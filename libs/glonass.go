package libs

import (
	"fmt"
	"time"

	"github.com/wfu-work/nav-go-lib/domains"
	"github.com/wfu-work/nav-go-lib/utils"
)

// parseGLONASSNav 解析GLONASS的星历报文
func parseGLONASSNav(msgType uint16, payload []byte, trTime time.Time) any {
	eph := &domains.Ephe{}
	br := NewBitReader(payload)
	br.SkipBits(12)

	//eph.System = "R"
	satId := int(br.ReadBits(6)) // DF038 卫星号
	eph.Sat = domains.SatType(fmt.Sprintf("R%02d", satId))
	eph.FreqN = int(br.ReadBits(5)) - 7 // DF040 频率通道号（-7 ~ +6）
	//satHealth := br.ReadBits(1)         // DF104 历书健康状况
	//avail := br.ReadBits(1)             // DF105 历书健康状况可用性标志
	//eph.Svh = int((avail << 1) | satHealth)
	_ = br.ReadBits(2)
	_ = br.ReadBits(2)         // DF106 P1
	tkH := br.ReadBits(5)      // DF107 tk 电文时间
	tkM := br.ReadBits(6)      // DF107 tk 电文时间
	tkS := br.ReadBits(1) * 30 // DF107 tk 电文时间
	bn := br.ReadBits(1)       // DF108 Bn MSB
	_ = br.ReadBits(1)         // DF109 P2
	tb := br.ReadBits(7)       // DF110 tb 星历时间（单位：15分钟）

	// 解析速度、位置、加速度数据
	eph.VecX = br.ReadBitG(24) * utils.Pow2(-20)
	eph.PosX = br.ReadBitG(27) * utils.Pow2(-11)
	eph.AccX = br.ReadBitG(5) * utils.Pow2(-30)
	eph.VecY = br.ReadBitG(24) * utils.Pow2(-20)
	eph.PosY = br.ReadBitG(27) * utils.Pow2(-11)
	eph.AccY = br.ReadBitG(5) * utils.Pow2(-30)
	eph.VecZ = br.ReadBitG(24) * utils.Pow2(-20)
	eph.PosZ = br.ReadBitG(27) * utils.Pow2(-11)
	eph.AccZ = br.ReadBitG(5) * utils.Pow2(-30)

	_ = br.ReadBits(1)                             // DF120 P3
	eph.GammaN = br.ReadBitG(11) * utils.Pow2(-40) // DF121 γn(tb)
	_ = br.ReadBits(2)                             // DF122 P
	_ = br.ReadBits(1)                             // DF123 ln(3)
	eph.TauN = br.ReadBitG(22) * utils.Pow2(-30)   // DF124 τn(tb)
	//eph.Dtaun = float64(br.ReadSignedBits(5)) * utils.Pow2(-30) // DF125 Δτn
	br.SkipBits(5)                // 跳过Δτn，新结构中没有这个字段
	eph.Age = int(br.ReadBits(5)) // DF126 En
	_ = br.ReadBits(1)            // DF127 P4
	_ = br.ReadBits(4)            // DF128 FT
	_ = br.ReadBits(11)           // DF129 NT (帧号)
	_ = br.ReadBits(2)            // DF130 M
	_ = br.ReadBits(1)            // DF131 附加数据可用性
	_ = br.ReadBits(11)           // DF132 NA
	_ = br.ReadSignedBits(32)     // DF133 τC
	_ = br.ReadBits(5)            // DF134 N4
	_ = br.ReadBits(22)           // TGPS
	_ = br.ReadBits(1)
	_ = br.ReadBits(7)

	tbSec := float64(tb) * 900.0
	eph.Toe = *domains.NewGTime(utils.DecodeGlonassTime(trTime, tbSec))
	tkSec := float64(tkH)*3600 + float64(tkM)*60 + float64(tkS)
	eph.Toc = *domains.NewGTime(utils.DecodeGlonassTime(trTime, tkSec))

	eph.Iode = int(tb & 0x7F)
	eph.Svh = int(bn)
	return eph
}
