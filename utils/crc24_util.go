package utils

var crc24qTable [256]uint32

func init() {
	for i := 0; i < 256; i++ {
		var crc = uint32(i) << 16
		for j := 0; j < 8; j++ {
			if (crc & 0x800000) != 0 {
				crc = (crc << 1) ^ 0x1864CFB
			} else {
				crc <<= 1
			}
		}
		crc24qTable[i] = crc & 0xFFFFFF
	}
}

// CalcCRC24Q 计算数据的 24-bit CRC（返回低 24 位）
func CalcCRC24Q(data []byte) uint32 {
	var crc uint32 = 0
	for _, b := range data {
		index := byte((crc>>16)^uint32(b)) & 0xFF
		crc = ((crc << 8) & 0xFFFFFF) ^ crc24qTable[index]
	}
	return crc & 0xFFFFFF
}
