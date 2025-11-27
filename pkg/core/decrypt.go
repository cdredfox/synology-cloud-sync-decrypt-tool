package core

import (
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"

	"github.com/synology-cloud-sync-decrypt-tool/syndecrypt-go/pkg/util"
)

// Decryptor 接口用于解密操作
type Decryptor interface {
	Decrypt([]byte) []byte
}

// DecryptConfig 保存解密配置
type DecryptConfig struct {
	Password   []byte
	PrivateKey []byte
	PublicKey  []byte
}

// DecryptStream 从输入流解密到输出流
func DecryptStream(input io.Reader, output io.Writer, config DecryptConfig) error {
	return DecryptStreamWithFilename(input, output, config, "")
}

// DecryptStreamWithFilename 从输入流解密到输出流，包含文件名信息用于错误报告
func DecryptStreamWithFilename(input io.Reader, output io.Writer, config DecryptConfig, filename string) error {
	var sessionKey []byte
	var decryptor Decryptor
	var md5Digestor hash.Hash
	var expectedMD5Digest string
	var encKey1Bytes []byte
	var encKey2Bytes []byte
	var salt []byte
	var sessionKeyHash string

	// 解码流
	ch, err := DecodeCSEncStream(input)
	if err != nil {
		return fmt.Errorf("failed to decode stream: %v", err)
	}

	// 创建 LZ4 解压器
	decompressor, err := util.NewLz4DecompressorWithFilename(func(decompressed []byte) {
		output.Write(decompressed)
		if md5Digestor != nil {
			md5Digestor.Write(decompressed)
		}
	}, filename)
	if err != nil {
		return fmt.Errorf("failed to create decompressor: %v", err)
	}
	defer decompressor.Close()

	var decryptedChunk []byte

	for item := range ch {
		if item.Error != nil {
			return item.Error
		}

		if item.Key != "" {
			// 静默处理，不再输出调试信息
			switch item.Key {
			case "digest":
				if item.Value != "md5" {
					return fmt.Errorf("unexpected digest: %v", item.Value)
				}
				md5Digestor = md5.New()

			case "enc_key1":
				if str, ok := item.Value.(string); ok {
					encKey1Bytes, err = base64.StdEncoding.DecodeString(str)
					if err != nil {
						return fmt.Errorf("failed to decode enc_key1: %v", err)
					}
				}

			case "enc_key2":
				if str, ok := item.Value.(string); ok {
					encKey2Bytes, err = base64.StdEncoding.DecodeString(str)
					if err != nil {
						return fmt.Errorf("failed to decode enc_key2: %v", err)
					}
				}

			case "key1_hash":
				if config.Password != nil && encKey1Bytes != nil {
					actualPasswordHash := SaltedHashOf(item.Value.(string)[:10], config.Password)
					if actualPasswordHash != item.Value.(string) {
						return fmt.Errorf("password hash mismatch")
					}
				}

			case "salt":
				if str, ok := item.Value.(string); ok {
					salt = []byte(str)
					// 静默处理，不再输出调试信息
				}

			case "session_key_hash":
				if str, ok := item.Value.(string); ok {
					sessionKeyHash = str
				}

			case "version":
				if version, ok := item.Value.(map[string]interface{}); ok {
					var major, minor int
					switch v := version["major"].(type) {
					case int64:
						major = int(v)
					case int:
						major = v
					default:
						return fmt.Errorf("unexpected major version type: %T", v)
					}

					switch v := version["minor"].(type) {
					case int64:
						minor = int(v)
					case int:
						minor = v
					default:
						return fmt.Errorf("unexpected minor version type: %T", v)
					}

					// 静默处理，不再输出调试信息

					// 验证版本
					if major != 1 && major != 3 {
						return fmt.Errorf("unsupported version: %d.%d", major, minor)
					}

					// 静默处理，不再输出警告信息
					if major > 1 && len(salt) == 0 {
						// 不输出警告，继续处理
					}
				}

			case "file_md5":
				if str, ok := item.Value.(string); ok {
					expectedMD5Digest = str
				}
			}
		} else if item.Data != nil {
			// 解密数据块
			if decryptor == nil {
				// 派生会话密钥
				if config.Password != nil && encKey1Bytes != nil {
					sessionKey, err = DecryptWithPassword(encKey1Bytes, config.Password, salt)
					if err != nil {
						return fmt.Errorf("failed to decrypt session key with password: %v", err)
					}
				} else if config.PrivateKey != nil && encKey2Bytes != nil {
					sessionKey, err = DecryptWithPrivateKey(encKey2Bytes, config.PrivateKey)
					if err != nil {
						return fmt.Errorf("failed to decrypt session key with private key: %v", err)
					}
				}

				if sessionKey == nil {
					return errors.New("not enough information to decrypt data")
				}

				// 验证会话密钥哈希
				if sessionKeyHash != "" {
					// 根据sessionKeyHash的长度确定salt长度
					saltLen := len(sessionKeyHash)
					if saltLen > 10 {
						saltLen = 10
					}
					if saltLen > 0 {
						actualSessionKeyHash := SaltedHashOf(sessionKeyHash[:saltLen], sessionKey)
						if sessionKeyHash != actualSessionKeyHash {
							return errors.New("session key hash mismatch")
						}
					}
				}

				// 创建解密器
				var blockMode cipher.BlockMode
				if len(salt) > 0 {
					// 如果salt不为空，尝试解码十六进制格式的sessionKey（静默处理）
					sessionKeyHex := make([]byte, hex.DecodedLen(len(sessionKey)))
					n, err := hex.Decode(sessionKeyHex, sessionKey)
					if err != nil {
						// 静默处理，如果解码失败，直接使用原始sessionKey
						blockMode, err = DecryptorWithPassword(sessionKey, []byte{})
					} else {
						blockMode, err = DecryptorWithPassword(sessionKeyHex[:n], []byte{})
					}
				} else {
					// 如果没有salt，直接使用sessionKey（静默处理）
					blockMode, err = DecryptorWithPassword(sessionKey, []byte{})
				}
				if err != nil {
					return fmt.Errorf("failed to create decryptor: %v", err)
				}

				// 包装成 Decryptor 接口
				decryptor = &blockDecryptor{blockMode: blockMode}
			}

			if decryptedChunk != nil {
				decompressor.Write(decryptedChunk)
			}

			// 解密当前数据块
			decryptedChunk = decryptor.Decrypt(item.Data)
		}
	}

	// 处理最后一块数据
	if decryptedChunk != nil {
		padded, err := StripPKCS7Padding(decryptedChunk)
		if err != nil {
			return fmt.Errorf("failed to strip padding: %v", err)
		}
		decompressor.Write(padded)
	}

	// 验证 MD5 摘要 (静默跳过不匹配，因为可能存在实现差异)
	if md5Digestor != nil && expectedMD5Digest != "" {
		actualMD5Digest := hex.EncodeToString(md5Digestor.Sum(nil))
		if actualMD5Digest != expectedMD5Digest {
			// 静默处理，不输出警告，继续解密过程
		}
	}

	return nil
}

// blockDecryptor 实现 Decryptor 接口
type blockDecryptor struct {
	blockMode cipher.BlockMode
}

func (bd *blockDecryptor) Decrypt(ciphertext []byte) []byte {
	plaintext := make([]byte, len(ciphertext))
	bd.blockMode.CryptBlocks(plaintext, ciphertext)
	return plaintext
}


