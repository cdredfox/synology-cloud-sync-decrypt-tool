package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

// OpenSSL KDF - 基于密码生成密钥和IV的算法
func OpenSSLKDF(password, salt []byte, keySize, ivSize int) ([]byte, []byte, error) {
	count := 1
	if len(salt) > 0 {
		count = 1000
	}

	temp := []byte{}
	fd := []byte{}

	for len(fd) < keySize+ivSize {
		hashedCountTimes := append(temp, password...)
		hashedCountTimes = append(hashedCountTimes, salt...)

		for i := 0; i < count; i++ {
			hashedCountTimes = MD5Hash(hashedCountTimes)
		}

		temp = hashedCountTimes
		fd = append(fd, temp...)
	}

	key := fd[:keySize]
	iv := fd[keySize : keySize+ivSize]

	return key, iv, nil
}

func MD5Hash(data []byte) []byte {
	h := md5.New()
	h.Write(data)
	return h.Sum(nil)
}

// 去除PKCS7填充
func StripPKCS7Padding(data []byte) ([]byte, error) {
	if len(data)%16 != 0 {
		return nil, errors.New("invalid length")
	}

	pad := data[len(data)-1]
	if pad > 16 {
		return nil, fmt.Errorf("invalid padding byte: %d", pad)
	}

	// 验证所有填充字节
	for i := len(data) - int(pad); i < len(data); i++ {
		if data[i] != pad {
			return nil, fmt.Errorf("invalid padding at position %d", i)
		}
	}

	return data[:len(data)-int(pad)], nil
}

// 使用密码解密
func DecryptWithPassword(ciphertext, password, salt []byte) ([]byte, error) {
	decryptor, err := DecryptorWithPassword(password, salt)
	if err != nil {
		return nil, err
	}

	// 创建输出缓冲区
	decrypted := make([]byte, len(ciphertext))
	decryptor.CryptBlocks(decrypted, ciphertext)

	return StripPKCS7Padding(decrypted)
}


// 创建基于密码的解密器
func DecryptorWithPassword(password, salt []byte) (cipher.BlockMode, error) {
	key, iv, err := CSENCPBKDF(password, salt)
	if err != nil {
		return nil, err
	}

	return DecryptorWithKeyIV(key, iv)
}

// CSENC PBKDF - Synology特定的密钥派生函数
func CSENCPBKDF(password, salt []byte) ([]byte, []byte, error) {
	const aesKeySizeBits = 256
	const aesIVLengthBytes = 16 // AES block size

	return OpenSSLKDF(password, salt, aesKeySizeBits/8, aesIVLengthBytes)
}

// 使用密钥和IV创建解密器
func DecryptorWithKeyIV(key, iv []byte) (cipher.BlockMode, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return cipher.NewCBCDecrypter(block, iv), nil
}

// 使用私钥解密
func DecryptWithPrivateKey(ciphertext, privateKey []byte) ([]byte, error) {
	privKey, err := x509.ParsePKCS1PrivateKey(privateKey)
	if err != nil {
		// 尝试解析PKCS8格式
		privKeyInterface, err := x509.ParsePKCS8PrivateKey(privateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %v", err)
		}
		var ok bool
		privKey, ok = privKeyInterface.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not an RSA private key")
		}
	}

	return rsa.DecryptOAEP(sha1.New(), rand.Reader, privKey, ciphertext, nil)
}

// 加盐哈希
func SaltedHashOf(salt string, data []byte) string {
	h := md5.New()
	h.Write([]byte(salt))
	h.Write(data)
	return salt + hex.EncodeToString(h.Sum(nil))
}

// 验证加盐哈希
func IsSaltedHashCorrect(saltedHash string, data []byte) bool {
	if len(saltedHash) < 10 {
		return false
	}
	if len(saltedHash) < 10+32 { // salt + MD5 hash (32 hex chars)
		return false
	}
	// 找到salt的长度 - 总长度减去32个十六进制字符
	saltLen := len(saltedHash) - 32
	if saltLen > 10 {
		saltLen = 10 // 最多使用10个字符作为salt
	}
	expected := SaltedHashOf(saltedHash[:saltLen], data)
	return expected == saltedHash
}

// Base64解码辅助函数
func Base64Decode(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}