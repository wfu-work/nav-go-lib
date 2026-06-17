package libs

import (
	"fmt"
	"strings"
	"time"
)

type ParserFunc func(msgType uint16, payload []byte, trTime time.Time) any

var parserRegistry = map[uint16]ParserFunc{}

func RegisterParser(msgType uint16, p ParserFunc) {
	parserRegistry[msgType] = p
}

func RegisterNavParsers() {
	RegisterParser(1019, ParseGPSNav)
	RegisterParser(1020, parseGLONASSNav)
	RegisterParser(1042, ParseBDSNav)
	RegisterParser(1045, parseGalileoFNav)
	RegisterParser(1046, parseGalileoINav)
}

// DefaultHandler default handler when no parser found
func DefaultHandler(msgType uint16, payload []byte) {
	fmt.Printf("MSG %d (len=%d bytes):\n", msgType, len(payload))
	fmt.Printf("payload bits preview: %s\n", bitsPreview(payload, 120))
	guess := guessConstellationByMsgType(msgType)
	fmt.Printf("guessed constellation: %s\n\n", guess)
}

// bitsPreview returns first N chars of bits representation
func bitsPreview(b []byte, maxChars int) string {
	var sb strings.Builder
	count := 0
	for i := 0; i < len(b) && count < maxChars; i++ {
		for bit := 7; bit >= 0 && count < maxChars; bit-- {
			if ((b[i] >> uint(bit)) & 1) == 1 {
				sb.WriteByte('1')
			} else {
				sb.WriteByte('0')
			}
			count++
		}
		if count < maxChars {
			sb.WriteByte(' ')
			count++
		}
	}
	return sb.String()
}

// simple heuristic to guess constellation (常见 RTCM message type 分布为示例)
func guessConstellationByMsgType(mt uint16) string {
	switch {
	case mt >= 1000 && mt < 1100:
		return "GPS / legacy (常见 GPS 消息位于 1001-1012/1019-1020 等)"
	case mt >= 1100 && mt < 1200:
		return "GLONASS (可能)"
	case mt >= 1200 && mt < 1300:
		return "Galileo / BeiDou (可能)"
	case mt >= 3000 && mt < 4000:
		return "MSM (多星座常用)"
	case mt >= 4000 && mt < 5000:
		return "SSR (卫星态势/星历等 SSR 消息)"
	default:
		return "未知（需参考 RTCM 标准或样本）"
	}
}
