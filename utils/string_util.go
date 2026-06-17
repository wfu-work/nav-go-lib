package utils

import (
	"math"
	"strconv"
)

// BitsToString 输出 payload 的位字符串（用于调试）
func BitsToString(b []byte) string {
	out := make([]byte, 0, len(b)*8)
	for i := 0; i < len(b); i++ {
		for bit := 7; bit >= 0; bit-- {
			if ((b[i] >> uint(bit)) & 1) == 1 {
				out = append(out, '1')
			} else {
				out = append(out, '0')
			}
		}
		if i != len(b)-1 {
			out = append(out, ' ')
		}
	}
	return string(out)
}

func Pow2(n int) float64 { return math.Pow(2, float64(n)) }

func ScaleInt(v int64, scale float64) float64 {
	return float64(v) * scale
}

func ScaleFloat(v float64, scale float64) float64 {
	return v * scale
}

func Str2Int(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
