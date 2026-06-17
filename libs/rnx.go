package libs

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"navfirst.com/nav-go-lib/domains"
	"navfirst.com/nav-go-lib/utils"
)

// writeRinexHeader 写入标准 RINEX 3.02 NAV 文件头
func writeRinexHeader(f *os.File) error {
	now := time.Now().UTC()
	// 格式：3.02版本，N: NAV DATA, M: Mixed
	_, _ = fmt.Fprintf(f, "     3.02           N: GNSS NAV DATA    M: Mixed            RINEX VERSION / TYPE\n")
	// PGM / RUN BY / DATE 行（左对齐+UTC时间）
	_, _ = fmt.Fprintf(f, "%-20s%-20s%20s UTC PGM / RUN BY / DATE \n",
		"NavGo",                       // 程序名
		"Wuhan Navfirst",              // 运行单位
		now.Format("20060102 150405"), // 日期格式
	)
	// 预留空行，用于未来扩展（可以添加注释）
	_, _ = fmt.Fprintf(f, "                                                            END OF HEADER       \n")
	return nil
}

func WriteNavRinex(nav *domains.Nav, outPath string) error {
	existingKeys := loadExistingNavKeys(outPath)
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(outPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, _ := f.Stat()
	if stat.Size() == 0 {
		if err := writeRinexHeader(f); err != nil {
			return err
		}
	}
	// 遍历nav map中的所有星历
	var sats []domains.SatType
	for sat := range *nav {
		sats = append(sats, sat)
	}
	for _, sat := range domains.Sorted(sats) {
		ephs := (*nav)[sat]
		for _, eph := range ephs {
			sys := string(sat.Sys())
			recordTime := eph.Toc.ToTime()
			if sys == "R" {
				recordTime = eph.Toe.ToTime()
			}
			if navKeyExists(existingKeys, sys, sat.Num(), recordTime) {
				continue
			}
			satStr := fmt.Sprintf("%s%02d", sys, sat.Num())
			switch sys {
			case "G", "C", "E": // GPS, BDS, Galileo
				utcTime := recordTime
				_, _ = fmt.Fprintf(f, "%-3s %4d %02d %02d %02d %02d %02d%s%s%s\n",
					satStr,
					utcTime.Year(), utcTime.Month(), utcTime.Day(),
					utcTime.Hour(), utcTime.Minute(), utcTime.Second(),
					formatRinexFloatD(eph.Af0),
					formatRinexFloatD(eph.Af1),
					formatRinexFloatD(eph.Af2),
				)
				writeRinexEphLine(f, float64(eph.Iode), eph.Crs, eph.DeltaN, eph.M0)
				writeRinexEphLine(f, eph.Cuc, eph.Ecc, eph.Cus, eph.SqrtA)
				writeRinexEphLine(f, eph.Toe.Sec, eph.Cic, eph.Omega0, eph.Cis)
				writeRinexEphLine(f, eph.I0, eph.Crc, eph.Omega, eph.OmegaD)
				writeRinexEphLine(f, eph.Idot, float64(eph.Code), float64(eph.Week), float64(eph.Flag))

				// 根据系统类型写入不同的参数
				switch sys {
				case "G": // GPS
					writeRinexEphLine(f, utils.UraValue(eph.Sva), float64(eph.Svh), eph.Tgd, float64(eph.Iodc))
					ttr := 0.0
					writeRinexEphLine(f, ttr, eph.Fit)
				case "C": // BDS
					writeRinexEphLine(f, utils.UraValue(eph.Sva), float64(eph.Svh), eph.Tgd, eph.Tgd2)
					ttr := -14.0
					writeRinexEphLine(f, ttr, float64(eph.Iodc))
				case "E": // Galileo
					writeRinexEphLine(f, utils.SisaValue(eph.Sva), float64(eph.Svh), eph.Tgd, eph.Tgd2)
					ttr := 0.0
					writeRinexEphLine(f, ttr, 0)
				default:
					break
				}
			case "R": // GLONASS
				messageFrameTime := utils.GlonassTkFromToc(&eph.Toc)
				_, _ = fmt.Fprintf(f, "%-3s %4d %02d %02d %02d %02d %02d%s%s%s\n",
					satStr,
					recordTime.Year(), recordTime.Month(), recordTime.Day(),
					recordTime.Hour(), recordTime.Minute(), recordTime.Second(),
					formatRinexFloatD(-eph.TauN),
					formatRinexFloatD(eph.GammaN),
					formatRinexFloatD(messageFrameTime),
				)
				writeRinexEphLine(f, eph.PosX, eph.VecX, eph.AccX, float64(eph.Svh))
				writeRinexEphLine(f, eph.PosY, eph.VecY, eph.AccY, float64(eph.FreqN))
				writeRinexEphLine(f, eph.PosZ, eph.VecZ, eph.AccZ, float64(eph.Age))
			}
		}
	}
	return nil
}

// 解析已存在文件中的 Nav 数据 (仅解析系统+卫星号+Toe)
func loadExistingNavKeys(outPath string) map[string]struct{} {
	keys := make(map[string]struct{})

	f, err := os.Open(outPath)
	if err != nil {
		return keys
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 3 {
			continue
		}
		sys := line[0:1]
		if len(line) < 23 || !(strings.HasPrefix(line, "G") || strings.HasPrefix(line, "C") || strings.HasPrefix(line, "E") || strings.HasPrefix(line, "R")) {
			continue
		}
		satStr := strings.TrimSpace(line[1:3])
		year := strings.TrimSpace(line[4:8])
		month := strings.TrimSpace(line[9:11])
		day := strings.TrimSpace(line[12:14])
		hour := strings.TrimSpace(line[15:17])
		mi := strings.TrimSpace(line[18:20])
		sec := strings.TrimSpace(line[21:23])
		key := navRecordKey(sys, satStr, year, month, day, hour, mi, sec)
		keys[key] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("warning: reading existing file failed: %v\n", err)
	}
	return keys
}

// 判断某颗星历是否已存在
func navKeyExists(keys map[string]struct{}, sys string, sat int, toe time.Time) bool {
	key := navRecordKey(
		sys,
		fmt.Sprintf("%02d", sat),
		fmt.Sprintf("%04d", toe.Year()),
		fmt.Sprintf("%02d", toe.Month()),
		fmt.Sprintf("%02d", toe.Day()),
		fmt.Sprintf("%02d", toe.Hour()),
		fmt.Sprintf("%02d", toe.Minute()),
		fmt.Sprintf("%02d", toe.Second()),
	)
	_, ok := keys[key]
	return ok
}

func navRecordKey(sys, sat, year, month, day, hour, minute, second string) string {
	return fmt.Sprintf("%s%02s%04s%02s%02s%02s%02s%02s", sys, sat, year, month, day, hour, minute, second)
}

// formatRinexFloatD 格式化为 RINEX D 格式（固定19宽，指数D±2位）
func formatRinexFloatD(val float64) string {
	if val == 0 {
		return "  .000000000000D+00"
	}
	s := fmt.Sprintf("%.12E", val)
	parts := strings.Split(s, "E")
	m, _ := strconv.ParseFloat(parts[0], 64)
	e, _ := strconv.Atoi(parts[1])
	m /= 10.0
	e++
	mStr := fmt.Sprintf("%.12f", m)
	if strings.HasPrefix(mStr, "-0") {
		mStr = "-" + mStr[2:]
	}
	if strings.HasPrefix(mStr, "0") {
		mStr = mStr[1:]
	}
	if strings.HasPrefix(mStr, "-") {
		return fmt.Sprintf(" %sD%+03d", mStr, e)
	} else {
		return fmt.Sprintf("  %sD%+03d", mStr, e)
	}
}

// writeRinexEphLine 写入最多4个参数一行，自动补齐缩进
func writeRinexEphLine(f *os.File, vals ...float64) {
	_, _ = f.WriteString("    ")
	for i, v := range vals {
		_, _ = f.WriteString(formatRinexFloatD(v))
		if i == len(vals)-1 {
			_, _ = f.WriteString("\n")
		}
	}
}

// 写 RINEX 严格对齐，D+xx 科学计数法
func formatRinexFloat(f float64) string {
	return fmt.Sprintf("%19.12E", f)
}
