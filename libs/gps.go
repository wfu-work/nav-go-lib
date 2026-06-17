package libs

import (
	"fmt"
	"math"
	"time"

	"github.com/wfu-work/nav-go-lib/domains"
	"github.com/wfu-work/nav-go-lib/utils"
)

// ParseGPSNav 解析GPS的星历报文
func ParseGPSNav(msgType uint16, payload []byte, trTime time.Time) any {
	eph := &domains.Ephe{}
	br := NewBitReader(payload)
	// 按 RTCM 1019/1020 结构，前 12 bit 已为 msgType —— 如果需要从头开始解析，请先跳过 12
	br.SkipBits(12)

	//eph.System = "G"
	satId := int(br.ReadBits(6)) // DF009: GPS Satellite ID (1–32)
	eph.Sat = domains.SatType(fmt.Sprintf("G%02d", satId))
	week10 := int(br.ReadBits(10)) // DF076: GPS Week Number
	fullWeek := utils.GPSWeek10(week10, trTime.UTC())
	eph.Week = fullWeek
	eph.Sva = int(br.ReadBits(4))  // DF077: URA
	eph.Code = int(br.ReadBits(2)) // DF078: L2 Code

	eph.Idot = math.Pow(2, -43) * float64(br.ReadSignedBits(14)) * math.Pi // DF079
	eph.Iode = int(br.ReadBits(8))                                         // DF071
	tocRaw := br.ReadBits(16)                                              // DF081
	eph.Toc = domains.GTime{Week: fullWeek, Sec: float64(tocRaw) * 16.0}

	eph.Af2 = math.Pow(2, -55) * float64(br.ReadSignedBits(8))  // DF082
	eph.Af1 = math.Pow(2, -43) * float64(br.ReadSignedBits(16)) // DF083
	eph.Af0 = math.Pow(2, -31) * float64(br.ReadSignedBits(22)) // DF084

	eph.Iodc = int(br.ReadBits(10))                                          // DF085
	eph.Crs = math.Pow(2, -5) * float64(br.ReadSignedBits(16))               // DF086
	eph.DeltaN = math.Pow(2, -43) * float64(br.ReadSignedBits(16)) * math.Pi // DF087
	eph.M0 = math.Pow(2, -31) * float64(br.ReadSignedBits(32)) * math.Pi     // DF088
	eph.Cuc = math.Pow(2, -29) * float64(br.ReadSignedBits(16))              // DF089
	eph.Ecc = math.Pow(2, -33) * float64(br.ReadBits(32))                    // DF090
	eph.Cus = math.Pow(2, -29) * float64(br.ReadSignedBits(16))              // DF091

	sqrta := math.Pow(2, -19) * float64(br.ReadBits(32)) // DF092
	eph.SqrtA = sqrta

	toeRaw := float64(br.ReadBits(16)) * 16.0 // DF093
	//eph.Toes = toeRaw
	eph.Toe = domains.GTime{Week: fullWeek, Sec: toeRaw}

	eph.Cic = math.Pow(2, -29) * float64(br.ReadSignedBits(16))              // DF094
	eph.Omega0 = math.Pow(2, -31) * float64(br.ReadSignedBits(32)) * math.Pi // DF095
	eph.Cis = math.Pow(2, -29) * float64(br.ReadSignedBits(16))              // DF096
	eph.I0 = math.Pow(2, -31) * float64(br.ReadSignedBits(32)) * math.Pi     // DF097
	eph.Crc = math.Pow(2, -5) * float64(br.ReadSignedBits(16))               // DF098
	eph.Omega = math.Pow(2, -31) * float64(br.ReadSignedBits(32)) * math.Pi  // DF099
	eph.OmegaD = math.Pow(2, -43) * float64(br.ReadSignedBits(24)) * math.Pi // DF100

	eph.Tgd = math.Pow(2, -31) * float64(br.ReadSignedBits(8)) // DF101
	eph.Svh = int(br.ReadBits(6))                              // DF102
	eph.Flag = int(br.ReadBits(1))                             // DF103 (L2 P flag)
	fitFlag := br.ReadBits(1)                                  // DF137
	if fitFlag == 0 {
		eph.Fit = 4.0
	} else {
		eph.Fit = 6.0
	}
	eph.Tot = eph.Toe // 发射时间近似设为 Toe
	return eph
}
