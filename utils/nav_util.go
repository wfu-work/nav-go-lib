package utils

func Sqr(x float64) float64 {
	return x * x
}

func UraValue(sva int) float64 {
	uras := []float64{2.0, 2.8, 4.0, 5.7, 8.0, 11.3, 16.0, 32.0, 64.0, 128.0, 256.0, 512.0, 1024.0, 2048.0, 4096.0, 8192.0}
	if sva < 0 || sva > 14 {
		return 8192.0
	}
	return uras[sva]
}

func SisaValue(sisa int) float64 {
	switch {
	case sisa < 0:
		return -1.0
	case sisa <= 49:
		return float64(sisa) * 0.01
	case sisa <= 74:
		return 0.5 + float64(sisa-50)*0.02
	case sisa <= 99:
		return 1.0 + float64(sisa-75)*0.04
	case sisa <= 125:
		return 2.0 + float64(sisa-100)*0.16
	default:
		return -1.0
	}
}

func GetSvAccuracy(sys string, ura int) float64 {
	switch sys {
	case "G":
		gpsUra := []float64{
			2.4, 3.4, 4.85, 6.85, 9.65, 13.65, 24, 48,
			96, 192, 384, 768, 1536, 3072, 6144,
		}
		if ura < 0 || ura > 14 {
			return Sqr(6144.0)
		}
		return Sqr(gpsUra[ura])
	case "C":
		bdsUra := []float64{
			2.4, 3.4, 4.85, 6.85, 9.65, 13.65, 24, 48,
			96, 192, 384, 768, 1536, 3072, 6144,
		}
		if ura < 0 || ura > 14 {
			return Sqr(6144.0)
		}
		return Sqr(bdsUra[ura])
	case "E":
		if ura <= 49 {
			return Sqr(float64(ura) * 0.01)
		}
		if ura <= 74 {
			return Sqr(0.5 + float64(ura-50)*0.02)
		}
		if ura <= 99 {
			return Sqr(1.0 + float64(ura-75)*0.04)
		}
		if ura <= 125 {
			return Sqr(2.0 + float64(ura-100)*0.16)
		}
		return Sqr(500.0)
	case "R":
		gloUra := []float64{
			1.0, 2.0, 2.5, 4.0, 8.0, 16.0, 32.0,
		}
		if ura < 0 {
			return Sqr(32.0)
		}
		if ura >= len(gloUra) {
			return Sqr(gloUra[len(gloUra)-1])
		}
		return Sqr(gloUra[ura])
	}
	return Sqr(10.0)
}
