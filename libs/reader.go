package libs

import (
	"fmt"
	"io"

	"github.com/wfu-work/nav-go-lib/utils"
)

// BitReader 读取按位字段（按大端位序：先高位）
type BitReader struct {
	data []byte
	pos  int // 下一位的全局索引（从 0 开始，0 表示 data[0] 的最高位）
}

func NewBitReader(b []byte) *BitReader {
	return &BitReader{data: b, pos: 0}
}

func (br *BitReader) ReadBits(n int) uint64 {
	var val uint64
	totalBits := len(br.data) * 8
	if br.pos+n > totalBits {
		fmt.Printf("⚠️[BitReader] Attempt to read %d bits at pos %d (total %d bits)\n", n, br.pos, totalBits)
		n = totalBits - br.pos
		if n <= 0 {
			return 0
		}
	}
	for i := 0; i < n; i++ {
		bytePos := (br.pos + i) / 8
		bitPos := 7 - ((br.pos + i) % 8)
		bit := (br.data[bytePos] >> bitPos) & 1
		val = (val << 1) | uint64(bit)
	}
	br.pos += n
	return val
}

func (br *BitReader) ReadSignedBits(n int) int64 {
	if n <= 0 {
		return 0
	}
	if n >= 64 {
		u := br.ReadBits(64)
		return int64(u)
	}
	u := br.ReadBits(n)
	signMask := uint64(1) << uint(n-1)
	if u&signMask != 0 {
		u |= ^uint64(0) << uint(n)
	}
	return int64(u)
}

func (br *BitReader) ReadBitG(len int) float64 {
	if len <= 1 {
		return 0
	}
	sign := br.ReadBits(1)
	val := br.ReadBits(len - 1)
	if sign == 1 {
		return -float64(val)
	}
	return float64(val)
}

// ReadBits2 读取 n 位为无符号整数（n <= 64）
func (br *BitReader) ReadBits2(n int) uint64 {
	if n <= 0 || n > 64 {
		fmt.Printf("read bits err: %s\n", "n不在1-64之间")
		return 0
	}
	var val uint64 = 0
	for i := 0; i < n; i++ {
		byteIdx := br.pos / 8
		if byteIdx >= len(br.data) {
			fmt.Printf("read bits err: %s\n", "长度超了")
			return 0
		}
		bitInByte := 7 - (br.pos % 8)
		bit := (br.data[byteIdx] >> uint(bitInByte)) & 1
		val = (val << 1) | uint64(bit)
		br.pos++
	}
	return val
}

func (br *BitReader) ReadSignedBits2(n int) int64 {
	u := br.ReadBits(n)
	if n == 64 {
		return int64(u)
	}
	signMask := uint64(1) << uint(n-1)
	if (u & signMask) != 0 {
		return int64(u - (1 << uint(n)))
	}
	return int64(u)
}

func (br *BitReader) SkipBits(n int) { br.pos += n }

// ReadRTCM3Frames 从 reader 中读出单条 RTCM3 消息（处理粘包）
// 返回 payload（不含 header/CRC）和 msgType
func ReadRTCM3Frames(r io.Reader) (msgType uint16, payload []byte, err error) {
	buf := make([]byte, 1)
	for {
		_, err := r.Read(buf)
		if err != nil {
			return 0, nil, err
		}
		if buf[0] == 0xD3 {
			break
		}
	}

	// read next 2 bytes (length + reserved)
	h := make([]byte, 2)
	if _, err := io.ReadFull(r, h); err != nil {
		return 0, nil, err
	}
	lenWord := uint16(h[0])<<8 | uint16(h[1])
	length := lenWord & 0x03FF // 10-bit length
	if length == 0 {
		return 0, nil, fmt.Errorf("zero-length message")
	}

	full := make([]byte, int(length)+3)
	if _, err := io.ReadFull(r, full); err != nil {
		return 0, nil, err
	}
	payload = full[:length]
	crcBytes := full[length:]

	// calc CRC over preamble+header+payload
	crcData := append([]byte{0xD3, h[0], h[1]}, payload...)
	crc := utils.CalcCRC24Q(crcData)
	crcGot := uint32(crcBytes[0])<<16 | uint32(crcBytes[1])<<8 | uint32(crcBytes[2])
	if crc != crcGot {
		return 0, nil, fmt.Errorf("crc mismatch (want %06X got %06X)", crc, crcGot)
	}

	// message type: 前 12 位
	br := NewBitReader(payload)
	mt := br.ReadBits(12)
	return uint16(mt), payload, nil
}
