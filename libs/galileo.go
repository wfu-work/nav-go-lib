package libs

import (
	"fmt"
	"math"
	"time"

	"github.com/wfu-work/nav-go-lib/domains"
	"github.com/wfu-work/nav-go-lib/utils"
)

// parseGalileoFNav 解析Galileo的星历报文
func parseGalileoFNav(msgType uint16, payload []byte, trTime time.Time) any {
	br := NewBitReader(payload)
	br.SkipBits(12)
	eph := &domains.Ephe{}
	//eph.System = "E"
	satId := int(br.ReadBits(6)) // DF252
	eph.Sat = domains.SatType(fmt.Sprintf("E%02d", satId))
	galWeek12 := int(br.ReadBits(12))
	eph.Iode = int(br.ReadBits(10))
	sisa := br.ReadBits(8)
	eph.Sva = int(sisa)                                                            // SIS 精度参数（可忽略或存储备用）
	eph.Idot = float64(br.ReadSignedBits(14)) * (1.0 / (utils.Pow2(43) / math.Pi)) // DF292 IDOT (rad/s)
	toc := br.ReadBits(14)
	tocSec := float64(toc) * 60.0
	// 恢复伽利略连续周数
	galWeek := utils.RestoreGalileoWeek(galWeek12, trTime)
	// 伽利略时间转UTC时间
	tocUTC := utils.GalileoTimeToUtc(galWeek, tocSec)
	// UTC时间转GPS时间，然后创建GTime
	tocTow, tocGPSWeek := utils.Time2Gpst(tocUTC)
	eph.Week = tocGPSWeek
	eph.Toc = domains.GTime{Week: tocGPSWeek, Sec: tocTow - 18} // 秒 (单位为分钟)
	eph.Af2 = float64(br.ReadSignedBits(6)) * utils.Pow2(-59)
	eph.Af1 = float64(br.ReadSignedBits(21)) * utils.Pow2(-46)
	eph.Af0 = float64(br.ReadSignedBits(31)) * utils.Pow2(-34)
	eph.Crs = float64(br.ReadSignedBits(16)) * utils.Pow2(-5)
	eph.DeltaN = float64(br.ReadSignedBits(16)) * utils.Pow2(-43) * math.Pi
	eph.M0 = float64(br.ReadSignedBits(32)) * utils.Pow2(-31) * math.Pi
	eph.Cuc = float64(br.ReadSignedBits(16)) * utils.Pow2(-29)
	eph.Ecc = float64(br.ReadSignedBits(32)) * utils.Pow2(-33)
	eph.Cus = float64(br.ReadSignedBits(16)) * utils.Pow2(-29)
	eph.SqrtA = float64(br.ReadBits(32)) * utils.Pow2(-19)
	toe := float64(br.ReadBits(14)) * 60.0 // DF304 toe (秒, 单位=分钟)
	// 伽利略时间转UTC时间，再转GPS时间
	toeUTC := utils.GalileoTimeToUtc(galWeek, toe)
	toeTow, toeGPSWeek := utils.Time2Gpst(toeUTC)
	eph.Toe = domains.GTime{Week: toeGPSWeek, Sec: toeTow - 18}
	eph.Cic = float64(br.ReadSignedBits(16)) * utils.Pow2(-29)
	eph.Omega0 = float64(br.ReadSignedBits(32)) * utils.Pow2(-31) * math.Pi
	eph.Cis = float64(br.ReadSignedBits(16)) * utils.Pow2(-29)
	eph.I0 = float64(br.ReadSignedBits(32)) * utils.Pow2(-31) * math.Pi
	eph.Crc = float64(br.ReadSignedBits(16)) * utils.Pow2(-5)
	eph.Omega = float64(br.ReadSignedBits(32)) * utils.Pow2(-31) * math.Pi
	eph.OmegaD = float64(br.ReadSignedBits(24)) * utils.Pow2(-43) * math.Pi
	eph.Tgd = float64(br.ReadSignedBits(10)) * utils.Pow2(-32) // BGD(E1/E5a)
	//eph.Tgd2 = float64(br.ReadSignedBits(10)) * utils.Pow2(-32) // BGD(E1/E5b)
	br.SkipBits(2) // E5a health
	br.SkipBits(1) // E5a data validity
	br.SkipBits(7) // reserved
	eph.Tot = eph.Toe
	eph.Code = (1 << 1) + (1 << 8)
	eph.Iodc = eph.Iode
	return eph
}

// parseGalileoINav 解析Galileo的星历报文
func parseGalileoINav(msgType uint16, payload []byte, trTime time.Time) any {
	br := NewBitReader(payload)
	br.SkipBits(12)
	eph := &domains.Ephe{}
	//eph.System = "E"
	satId := int(br.ReadBits(6)) // DF252
	eph.Sat = domains.SatType(fmt.Sprintf("E%02d", satId))
	galWeek12 := int(br.ReadBits(12))                                              // DF289
	eph.Iode = int(br.ReadBits(10))                                                // DF290 IODnav
	sisa := br.ReadBits(8)                                                         // DF291
	eph.Sva = int(sisa)                                                            // SIS 精度参数（可忽略或存储备用）
	eph.Idot = float64(br.ReadSignedBits(14)) * (1.0 / (utils.Pow2(43) / math.Pi)) // DF292 IDOT (rad/s)
	// 恢复伽利略连续周数
	galWeek := utils.RestoreGalileoWeek(galWeek12, trTime)
	toc := float64(br.ReadBits(14)) * 60.0 // DF293 toc (秒, 单位=分钟)
	// 伽利略时间转UTC时间
	tocUTC := utils.GalileoTimeToUtc(galWeek, toc)
	// UTC时间转GPS时间，然后创建GTime
	tocTow, tocGPSWeek := utils.Time2Gpst(tocUTC)
	eph.Week = tocGPSWeek
	eph.Toc = domains.GTime{Week: tocGPSWeek, Sec: tocTow - 18}
	eph.Af2 = float64(br.ReadSignedBits(6)) * math.Pow(2, -59)  // DF294 af2 (s/s^2)
	eph.Af1 = float64(br.ReadSignedBits(21)) * math.Pow(2, -46) // DF295 af1 (s/s)
	eph.Af0 = float64(br.ReadSignedBits(31)) * math.Pow(2, -34) // DF296 af0 (s)

	// 轨道参数
	eph.Crs = float64(br.ReadSignedBits(16)) * math.Pow(2, -5)               // DF297 Crs (m)
	eph.DeltaN = float64(br.ReadSignedBits(16)) * math.Pow(2, -43) * math.Pi // DF298 Δn (rad/s)
	eph.M0 = float64(br.ReadBits(32)) * math.Pow(2, -31) * math.Pi           // DF299 M0 (rad)
	eph.Cuc = float64(br.ReadSignedBits(16)) * math.Pow(2, -29)              // DF300 Cuc (rad)
	eph.Ecc = float64(br.ReadBits(32)) * math.Pow(2, -33)                    // DF301 e
	eph.Cus = float64(br.ReadSignedBits(16)) * math.Pow(2, -29)              // DF302 Cus (rad)
	eph.SqrtA = float64(br.ReadBits(32)) * math.Pow(2, -19)                  // DF303 sqrt(A)
	toe := float64(br.ReadBits(14)) * 60.0                                   // DF304 toe (秒, 单位=分钟)
	// 伽利略时间转UTC时间，再转GPS时间
	toeUTC := utils.GalileoTimeToUtc(galWeek, toe)
	toeTow, toeGPSWeek := utils.Time2Gpst(toeUTC)
	eph.Toe = domains.GTime{Week: toeGPSWeek, Sec: toeTow - 18}
	eph.Cic = float64(br.ReadSignedBits(16)) * math.Pow(2, -29)              // DF305 Cic (rad)
	eph.Omega0 = float64(br.ReadSignedBits(32)) * math.Pow(2, -31) * math.Pi // DF306 Ω0 (rad)
	eph.Cis = float64(br.ReadSignedBits(16)) * math.Pow(2, -29)              // DF307 Cis (rad)
	eph.I0 = float64(br.ReadSignedBits(32)) * math.Pow(2, -31) * math.Pi     // DF308 i0 (rad)
	eph.Crc = float64(br.ReadSignedBits(16)) * math.Pow(2, -5)               // DF309 Crc (m)
	eph.Omega = float64(br.ReadSignedBits(32)) * math.Pow(2, -31) * math.Pi  // DF310 ω (rad)
	eph.OmegaD = float64(br.ReadSignedBits(24)) * math.Pow(2, -43) * math.Pi // DF311 ΩDOT (rad/s)
	eph.Tgd = float64(br.ReadSignedBits(10)) * math.Pow(2, -32)              // DF312 BGD(E1/E5a) (s)
	eph.Tgd2 = float64(br.ReadSignedBits(10)) * math.Pow(2, -32)             // DF312 BGD(E1/E5a) (s)
	e5aHealth := br.ReadBits(2)                                              // DF314
	e5aDValid := br.ReadBits(1)                                              // DF315
	_ = e5aHealth
	_ = e5aDValid
	eph.Iodc = eph.Iode
	eph.Code = (1 << 0) + (1 << 2) + (1 << 9)
	eph.Tot = eph.Toe
	return eph
}
