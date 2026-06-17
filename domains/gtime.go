package domains

import (
	"math"
	"time"
)

type GTime struct {
	Week int
	Sec  float64
}

func NewGTime(dt time.Time) *GTime {
	t := dt.Unix()
	t -= time.Date(1980, 1, 6, 0, 0, 0, 0, time.UTC).Unix() // Elapsed seconds since 1980/1/6 00:00:00
	return &GTime{
		Week: int(t / (3600 * 24 * 7)),
		Sec:  float64(t%(3600*24*7)) + float64(dt.Nanosecond())/1000000000,
	}
}

func (p *GTime) ToTime() time.Time {
	o := time.Date(1980, 1, 6, 0, 0, 0, 0, time.UTC).Unix() // GPS time starts from 1980/1/6 00:00:00
	i := int64(math.Trunc(p.Sec))
	t := int64(3600*24*7*p.Week) + i + o
	n := int64((p.Sec - float64(i)) * 1e9)
	return time.Unix(t, n).UTC() // Unix time is the elapsed seconds since 1970/1/1 00:00:00
}

func (p *GTime) Less(b GTime, roundSec bool) bool {
	if p.Week == b.Week {
		if roundSec {
			return math.Round(p.Sec) < math.Round(b.Sec)
		} else {
			return p.Sec < b.Sec
		}
	} else {
		return p.Week < b.Week
	}
}

func (p *GTime) LessOrEqual(b GTime, roundSec bool) bool {
	if p.Week == b.Week {
		if roundSec {
			return math.Round(p.Sec) <= math.Round(b.Sec)
		} else {
			return p.Sec <= b.Sec
		}
	} else {
		return p.Week < b.Week
	}
}

func (p *GTime) Before(t time.Time, roundSec bool) bool {
	return p.Less(*NewGTime(t), roundSec)
}

func (p *GTime) After(t time.Time, roundSec bool) bool {
	return NewGTime(t).Less(*p, roundSec)
}

func (p *GTime) Divisible(sec int) bool {
	return int(math.Round(p.Sec))%sec == 0
}
