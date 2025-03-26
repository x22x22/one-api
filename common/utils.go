package common

import (
	"fmt"
	"github.com/songquanpeng/one-api/common/config"
	"strings"
)

func LogQuota(quota int64) string {
	if config.DisplayInCurrencyEnabled {
		return fmt.Sprintf("＄%.6f 额度", float64(quota)/config.QuotaPerUnit)
	} else {
		return fmt.Sprintf("%d 点额度", quota)
	}
}

func MaskToken(token string) string {
	// 移除前缀
	token = strings.TrimPrefix(token, "Bearer ")

	// 检查token长度
	if len(token) < 12 {
		return token
	}

	// 保留前4位和后4位,中间用*替换
	masked := token[:6] + strings.Repeat("*", 4) + token[len(token)-8:]

	// 如果有Bearer前缀,加回去
	if strings.HasPrefix(token, "Bearer ") {
		masked = "Bearer " + masked
	}

	return masked
}

func SubString(str string, start, end int) string {
	if start < 0 || end > len([]rune(str)) || start > end {
		return ""
	}
	runes := []rune(str)
	return string(runes[start:end])
}
