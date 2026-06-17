package utils

import (
	"math"
	"time"

	"navfirst.com/nav-go-lib/domains"
)

var CurrentLeapSeconds = 18

const GLONASS_UTC_OFFSET = 3 * 3600

// Time2Gpst 将 UTC time 转为 GPST seconds-of-week 和 full week
func Time2Gpst(t time.Time) (tow float64, week int) {
	// GPS epoch
	base := time.Date(1980, 1, 6, 0, 0, 0, 0, time.UTC)
	// 把 t 转为 UTC，然后加上 (GPS - UTC) 的闰秒差 => GPS time = UTC + leapSeconds
	// 如果 t 已经是 GPST，则这一步会多加；所以确保 t 是 UTC 表示
	gpsTime := t.UTC().Add(time.Duration(CurrentLeapSeconds) * time.Second)
	// seconds since GPS epoch
	secs := gpsTime.Sub(base).Seconds()
	week = int(secs) / (7 * 24 * 3600)
	tow = secs - float64(week*7*24*3600)
	// 确保tow是整数秒，避免时间计算精度问题
	intTow := int64(tow)
	fracTow := tow - float64(intTow)
	if fracTow >= 0.5 {
		intTow++
		// 如果进位后超过一周的秒数，调整到下一周
		if intTow >= 7*24*3600 {
			intTow = 0
			week++
		}
	}
	return float64(intTow), week
}

// Gpst2Time 给定 week 和 tow 返回 GPS time (time.Time in UTC)
// result is GPS time converted back to UTC (i.e., subtract leapSeconds)
func Gpst2Time(week int, tow float64) time.Time {
	base := time.Date(1980, 1, 6, 0, 0, 0, 0, time.UTC)
	secs := float64(week*(7*24*3600)) + tow
	gpsTime := base.Add(time.Duration(math.Floor(secs)) * time.Second).Add(time.Duration((secs-math.Floor(secs))*1e9) * time.Nanosecond)
	// convert GPS -> UTC by subtracting leap seconds
	utc := gpsTime.Add(-time.Duration(CurrentLeapSeconds) * time.Second)
	return utc
}

// GpstTimeToUtc GPS epoch: 1980-01-06 00:00:00 UTC
func GpstTimeToUtc(fullWeek int, tow float64) time.Time {
	base := time.Date(1980, 1, 6, 0, 0, 0, 0, time.UTC)
	// 分离整数秒和小数部分，避免精度损失
	intSec := int64(tow)
	fracSec := tow - float64(intSec)
	// 添加周、秒和小数秒
	t := base.Add(time.Duration(fullWeek) * 7 * 24 * time.Hour)
	t = t.Add(time.Duration(intSec) * time.Second)
	t = t.Add(time.Duration(fracSec*1e9) * time.Nanosecond)
	// 转为 UTC（GPS时间 - 闰秒）
	t = t.Add(-time.Duration(CurrentLeapSeconds) * time.Second)
	return t
}

// GalileoTimeToUtc Galileo epoch: 1999-08-22
func GalileoTimeToUtc(week int, tow float64) time.Time {
	base := time.Date(1999, 8, 22, 0, 0, 0, 0, time.UTC)
	return base.Add(time.Duration(week*7*24*3600)*time.Second + time.Duration(tow)*time.Second)
}

// RestoreGalileoWeek Galileo 周数 恢复成连续周数
func RestoreGalileoWeek(week12 int, toc time.Time) int {
	// Galileo epoch 1999-08-22
	galEpoch := time.Date(1999, 8, 22, 0, 0, 0, 0, time.UTC)
	// 当前应当的连续周数（根据时间算的）
	trueWeeks := int(toc.Sub(galEpoch).Hours() / 24 / 7)
	// 合并高位：保持上面的周期不变，只替换低12bit
	base := trueWeeks &^ 4095 // 清掉底12bit
	candidate := base | week12
	// 距离太大时修正（避免跨周期错误）
	if candidate-trueWeeks > 2048 {
		candidate -= 4096
	} else if trueWeeks-candidate > 2048 {
		candidate += 4096
	}
	return candidate
}

// BdsTimeToUtc BeiDou epoch: 2006-01-01
// BDT (BeiDou Time) = UTC时间 + 8小时
// BDT周数从2006年1月1日开始计算
func BdsTimeToUtc(week int, tow float64) time.Time {
	// BDT时间基线：2006年1月1日0时0分0秒UTC
	base := time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC)
	// 将tow转换为整数秒，确保时间计算精确
	intTow := int64(tow)
	fracTow := tow - float64(intTow)
	bdtTime := base.Add(time.Duration(week*7*24*3600) * time.Second).Add(time.Duration(intTow) * time.Second).Add(time.Duration(fracTow*1e9) * time.Nanosecond)
	return bdtTime
}

// GPSWeek10 将10-bit周号扩展为full GPS week（基于当前时间，取最近的候选）
func GPSWeek10(week10 int, ref time.Time) int {
	base := time.Date(1980, 1, 6, 0, 0, 0, 0, time.UTC)
	// 当前参考周号（full）
	curWeek := int(ref.Sub(base).Seconds()) / (7 * 24 * 3600)
	// 计算 k，使得 fullWeek = week10 + k*1024 接近 curWeek
	// k = round((curWeek - week10) / 1024)
	k := int(math.Round(float64(curWeek-week10) / 1024.0))
	fullWeek := week10 + k*1024
	// 保险修正：若差距仍大于 512，修正方向
	if fullWeek-curWeek > 512 {
		fullWeek -= 1024
	} else if curWeek-fullWeek > 512 {
		fullWeek += 1024
	}
	return fullWeek
}

// AlignToNearestDay 将给定秒数（当天的秒）与参考时间 ref 对齐到“距离 ref 最近”的那一个日内时刻
func AlignToNearestDay(ref time.Time, secondsOfDay float64) time.Time {
	// 以 UTC 为基准
	ref = ref.UTC()
	dayStart := time.Date(ref.Year(), ref.Month(), ref.Day(), 0, 0, 0, 0, time.UTC)
	// 分离整数秒和小数部分，保留纳秒精度
	intSec := int64(math.Floor(secondsOfDay))
	frac := secondsOfDay - float64(intSec)
	t := dayStart.Add(time.Duration(intSec) * time.Second).Add(time.Duration(frac*1e9) * time.Nanosecond)
	// 如果候选时间与参考时间相差超过 12 小时，则向前或向后移动 24 小时以得到最近的那一天
	diff := t.Sub(ref)
	if diff > 12*time.Hour {
		t = t.Add(-24 * time.Hour)
	} else if diff < -12*time.Hour {
		t = t.Add(24 * time.Hour)
	}
	return t.UTC()
}

// getDayStartUTC 返回某天00:00:00的UTC时间
func getDayStartUTC(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// DayStartTtr 从 Toes 和 Toc 计算当天零点 Ttr
func DayStartTtr(toes float64, toc time.Time) float64 {
	hour, mi, sec := toc.Clock()
	secElapsed := float64(hour*3600 + mi*60 + sec)
	dayStart := toes - secElapsed
	if dayStart < 0 {
		dayStart += 7 * 24 * 3600 // 保证在 [0,604800)
	}
	return dayStart
}

// DecodeGlonassTime 处理GLONASS的Toe/Tof跨天
func DecodeGlonassTime(trTime time.Time, sec float64) time.Time {
	dayStart := getDayStartUTC(trTime)
	sec -= GLONASS_UTC_OFFSET
	t := dayStart.Add(time.Duration(sec) * time.Second)
	// 如果偏差 > 12 小时，修正跨天
	if t.Sub(trTime) > 12*time.Hour {
		t = t.Add(-24 * time.Hour)
	}
	if trTime.Sub(t) > 12*time.Hour {
		t = t.Add(24 * time.Hour)
	}
	return t
}

// GlonassTkFromToc 计算 GLONASS 导航电文 tk（秒）
func GlonassTkFromToc(toc *domains.GTime) float64 {
	gpsEpoch := time.Date(1980, 1, 6, 0, 0, 0, 0, time.UTC).Unix()
	gpsUnix := gpsEpoch + int64(toc.Week*7*86400) + int64(toc.Sec)
	utcUnix := gpsUnix
	utc := time.Unix(utcUnix, 0).UTC()
	tkH := utc.Hour()
	tkM := utc.Minute()
	tkS := utc.Second()
	return float64(tkH*3600 + tkM*60 + tkS)
}

// ParseTime 根据给定的时间字符串将其转换为 time.Time 类型
func ParseTime(timeStr string) (time.Time, error) {
	layout := "2006/01/02 15:04:05"
	t, err := time.Parse(layout, timeStr)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}
