package utils

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
)

// CalculateMD5 计算数据的MD5哈希值
func CalculateMD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// CalculateMD5FromReader 从Reader计算MD5哈希值
func CalculateMD5FromReader(reader io.Reader) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// CalculateCRC32 计算数据的CRC32校验值
func CalculateCRC32(data []byte) string {
	crc := crc32.ChecksumIEEE(data)
	return fmt.Sprintf("%08x", crc)
}

// CalculateHash 计算数据的哈希值（这里使用MD5作为示例）
func CalculateHash(data []byte) string {
	return CalculateMD5(data)
}

// EncodeBase64 将数据编码为Base64
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBase64 将Base64字符串解码为数据
func DecodeBase64(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

// EncodeImageToBase64 将图片数据编码为Base64格式，用于封面上传
func EncodeImageToBase64(imageData []byte, mimeType string) string {
	encoded := base64.StdEncoding.EncodeToString(imageData)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)
}

