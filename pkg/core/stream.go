package core

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

const (
	MagicHeader = "__CLOUDSYNC_ENC__"
)

// 流式解码器
type StreamDecoder struct {
	reader io.Reader
}

func NewStreamDecoder(reader io.Reader) *StreamDecoder {
	return &StreamDecoder{reader: reader}
}

// 验证魔数和哈希
func (sd *StreamDecoder) ValidateHeader() error {
	magic := make([]byte, len(MagicHeader))
	n, err := sd.reader.Read(magic)
	if err != nil {
		return err
	}
	if n != len(MagicHeader) {
		return errors.New("incomplete magic header")
	}

	if string(magic) != MagicHeader {
		return fmt.Errorf("invalid magic header: expected %s, got %s", MagicHeader, string(magic))
	}

	// 读取并验证魔数哈希
	hashBytes := make([]byte, 32)
	n, err = sd.reader.Read(hashBytes)
	if err != nil {
		return err
	}
	if n != 32 {
		return errors.New("incomplete magic hash")
	}

	expectedHash := md5.Sum([]byte(MagicHeader))
	expectedHashStr := hex.EncodeToString(expectedHash[:])

	if string(hashBytes) != expectedHashStr {
		return fmt.Errorf("invalid magic hash: expected %s, got %s", expectedHashStr, string(hashBytes))
	}

	return nil
}

// 从流中读取对象
func (sd *StreamDecoder) ReadObject() (interface{}, error) {
	headerByte := make([]byte, 1)
	n, err := sd.reader.Read(headerByte)
	if err != nil {
		return nil, err
	}
	if n != 1 {
		return nil, io.EOF
	}

	switch headerByte[0] {
	case 0x42: // OrderedDict
		return sd.readOrderedDict()
	case 0x40: // None/null
		return nil, nil
	case 0x11: // Bytes
		return sd.readBytes()
	case 0x10: // String
		return sd.readString()
	case 0x01: // Integer
		return sd.readInt()
	default:
		return nil, fmt.Errorf("unknown type byte: 0x%02X", headerByte[0])
	}
}

func (sd *StreamDecoder) readOrderedDict() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for {
		key, err := sd.ReadObject()
		if err != nil {
			return nil, err
		}
		if key == nil {
			break
		}

		keyStr, ok := key.(string)
		if !ok {
			return nil, errors.New("ordered dict key must be string")
		}

		value, err := sd.ReadObject()
		if err != nil {
			return nil, err
		}

		result[keyStr] = value
	}
	return result, nil
}

func (sd *StreamDecoder) readBytes() ([]byte, error) {
	lengthBytes := make([]byte, 2)
	n, err := sd.reader.Read(lengthBytes)
	if err != nil {
		return nil, err
	}
	if n != 2 {
		return nil, errors.New("incomplete length field")
	}

	length := binary.BigEndian.Uint16(lengthBytes)
	data := make([]byte, length)

	n, err = sd.reader.Read(data)
	if err != nil {
		return nil, err
	}
	if n != int(length) {
		return nil, errors.New("incomplete data")
	}

	return data, nil
}

func (sd *StreamDecoder) readString() (string, error) {
	data, err := sd.readBytes()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (sd *StreamDecoder) readInt() (int, error) {
	lengthByte := make([]byte, 1)
	n, err := sd.reader.Read(lengthByte)
	if err != nil {
		return 0, err
	}
	if n != 1 {
		return 0, errors.New("incomplete length byte")
	}

	length := int(lengthByte[0])
	if length > 8 {
		return 0, errors.New("integer too large")
	}

	if length == 0 {
		return 0, nil
	}

	data := make([]byte, length)
	n, err = sd.reader.Read(data)
	if err != nil {
		return 0, err
	}
	if n != length {
		return 0, errors.New("incomplete integer data")
	}

	// 大端序转换
	result := 0
	for i := 0; i < length; i++ {
		result = result<<8 | int(data[i])
	}

	return result, nil
}

// 解码 CloudSync 加密流
func DecodeCSEncStream(reader io.Reader) (<-chan StreamItem, error) {
	decoder := NewStreamDecoder(reader)
	if err := decoder.ValidateHeader(); err != nil {
		return nil, err
	}

	ch := make(chan StreamItem)
	go func() {
		defer close(ch)

		for {
			obj, err := decoder.ReadObject()
			if err != nil {
				if err != io.EOF {
					ch <- StreamItem{Error: err}
				}
				return
			}
			if obj == nil {
				continue
			}

			dict, ok := obj.(map[string]interface{})
			if !ok {
				ch <- StreamItem{Error: errors.New("expected dictionary object")}
				return
			}

			itemType, ok := dict["type"].(string)
			if !ok {
				ch <- StreamItem{Error: errors.New("missing type field")}
				return
			}

			switch itemType {
			case "metadata":
				for k, v := range dict {
					if k != "type" {
						ch <- StreamItem{Key: k, Value: v}
					}
				}
			case "data":
				if data, ok := dict["data"].([]byte); ok {
					ch <- StreamItem{Data: data}
				}
			}
		}
	}()

	return ch, nil
}

type StreamItem struct {
	Key   string
	Value interface{}
	Data  []byte
	Error error
}