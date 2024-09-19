package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func hashMD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:]) // 将结果转成十六进制字符串
}


func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	data,err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("md5: %s\n", hashMD5(data))
}