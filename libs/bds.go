package libs

import (
	"fmt"
	"math"
	"time"

	"navfirst.com/nav-go-lib/domains"
	"navfirst.com/nav-go-lib/utils"
)

// ParseBDSNav 解析BDS的星历报文
func ParseBDSNav(msgType uint16, payload []byte, trTime time.Time) any {
	br := NewBitReader(payload)
	br.SkipBits(12)
	eph := &domains.Ephe{}
	//eph.System = "C"
	satId := int(br.ReadBits(6))
	eph.Sat = domains.SatType(fmt.Sprintf("C%02d", satId))
	bdsWeek := int(br.ReadBits(13))
	eph.Sva = int(br.ReadBits(4))
	eph.Idot = utils.ScaleInt(br.ReadSignedBits(14), 1.0/(utils.Pow2(43))*math.Pi) // rad/s
	eph.Iode = int(br.ReadBits(5))
	toc := br.ReadBits(17)
	tocSec := float64(toc) * 8.0
	// 北斗时间转UTC时间
	tocUTC := utils.BdsTimeToUtc(bdsWeek, tocSec)
	// UTC时间转GPS时间，然后创建GTime
	tocTow, tocGPSWeek := utils.Time2Gpst(tocUTC)
	eph.Toc = domains.GTime{Week: tocGPSWeek, Sec: tocTow - 18}           // GPS时间
	eph.Af2 = utils.ScaleInt(br.ReadSignedBits(11), 1.0/(utils.Pow2(66))) // s/s^2
	eph.Af1 = utils.ScaleInt(br.ReadSignedBits(22), 1.0/(utils.Pow2(50))) // s/s
	eph.Af0 = utils.ScaleInt(br.ReadSignedBits(24), 1.0/(utils.Pow2(33))) // s
	eph.Iodc = int(br.ReadBits(5))
	eph.Crs = utils.ScaleInt(br.ReadSignedBits(18), 1.0/(utils.Pow2(6)))             // m
	eph.DeltaN = utils.ScaleInt(br.ReadSignedBits(16), 1.0/(utils.Pow2(43))*math.Pi) // rad/s
	eph.M0 = utils.ScaleInt(br.ReadSignedBits(32), 1.0/(utils.Pow2(31))*math.Pi)
	eph.Cuc = utils.ScaleInt(br.ReadSignedBits(18), 1.0/(utils.Pow2(31)))
	eph.Ecc = utils.ScaleInt(int64(br.ReadBits(32)), 1.0/(utils.Pow2(33)))
	eph.Cus = utils.ScaleInt(br.ReadSignedBits(18), 1.0/(utils.Pow2(31)))
	eph.SqrtA = utils.ScaleInt(int64(br.ReadBits(32)), 1.0/(utils.Pow2(19)))
	toe := br.ReadBits(17)
	toeSec := float64(toe) * 8.0
	// 北斗时间转UTC时间，再转GPS时间
	toeUTC := utils.BdsTimeToUtc(bdsWeek, toeSec)
	toeTow, toeGPSWeek := utils.Time2Gpst(toeUTC)
	eph.Week = bdsWeek
	eph.Toe = domains.GTime{Week: toeGPSWeek, Sec: toeTow - 18}
	eph.Cic = utils.ScaleInt(br.ReadSignedBits(18), 1.0/(utils.Pow2(31)))
	eph.Omega0 = utils.ScaleInt(br.ReadSignedBits(32), 1.0/(utils.Pow2(31))*math.Pi)
	eph.Cis = utils.ScaleInt(br.ReadSignedBits(18), 1.0/(utils.Pow2(31)))
	eph.I0 = utils.ScaleInt(br.ReadSignedBits(32), 1.0/(utils.Pow2(31))*math.Pi)
	eph.Crc = utils.ScaleInt(br.ReadSignedBits(18), 1.0/(utils.Pow2(6)))
	eph.Omega = utils.ScaleInt(br.ReadSignedBits(32), 1.0/(utils.Pow2(31))*math.Pi)
	eph.OmegaD = utils.ScaleInt(br.ReadSignedBits(24), 1.0/(utils.Pow2(43))*math.Pi)
	eph.Tgd = float64(br.ReadSignedBits(10)) * (1e-10)
	eph.Tgd2 = float64(br.ReadSignedBits(10)) * (1e-10)
	eph.Svh = int(br.ReadBits(1))
	eph.Tot = eph.Toe
	return eph
}
