package database

/**
 * 从用户名中获取基因数字
 * 基因用于定位数据中心 基因%数据中心数量（数据库分库总量）
 */

import (
	"errors"
	"strconv"
)

const dnaMaxBits int = 32

func Dna(userName string) (int, error) {
	if len(userName) <= 0 {
		return -1, errors.New("提取基因的字符串为空")
	}
	if number, err := strconv.Atoi(userName); err == nil {
		return dnaFromNumber(number)
	}
	var dna int = 0
	bts := []byte(userName)
	for _, bt := range bts {
		dna += int(bt)
	}
	return dnaFromNumber(dna)
}

func dnaFromNumber(userName int) (int, error) {

	// 长度判断
	dnsStr := strconv.Itoa(userName)
	dnaLen := len(dnsStr)
	if dnaLen <= dnaMaxBits {
		return userName, nil
	}
	// 长度超过指定最大长度，裁剪
	clipStr := dnsStr[dnaLen-dnaMaxBits : dnaLen]
	dna, err := strconv.Atoi(clipStr)
	if err != nil {
		return -1, err
	}
	return dna, nil
}
