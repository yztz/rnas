package rnas

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
)

// generateRandomData generates a random byte slice of the given size
func generateRandomData(size int) []byte {
	data := make([]byte, size)
	rand.Read(data)
	return data
}

// 计算 MD5 哈希
func hashMD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:]) // 将结果转成十六进制字符串
}


