package libs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wfu-work/nav-go-lib/domains"
	"github.com/wfu-work/nav-go-lib/utils"
)

const navUsage = "❌Usage: Nav <startTime> <endTime> <rawPath> <outPath> <sites> <fileFormat> <ext>"

type navArgs struct {
	startTime  time.Time
	endTime    time.Time
	rawPath    string
	outPath    string
	sites      []string
	fileFormat int
	ext        string
}

type navStats struct {
	gps     int
	bds     int
	glonass int
	galileo int
}

func (s navStats) total() int {
	return s.gps + s.bds + s.glonass + s.galileo
}

// Nav "%s@%s@%s@%s@%s"  PSNMJC014,PSNMJC074,PSNMJC057,PSJK94,PSJK88,PSJK84,HJSK01,ZSSK01,GSSK01,PSKHJC17,PSKHJC15,PSKHJC06,PSYMJC007,PSYMJC008
func Nav(argv string) string {
	args, err := parseNavArgs(argv)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	nav := collectNav(args, true)
	stats := printNavSummary(nav)
	if stats.total() == 0 {
		return ""
	}

	SortEphemeris(nav)

	outPath := navOutputPath(args.outPath, args.endTime)
	err = WriteNavRinex(nav, outPath)
	if err != nil {
		fmt.Println("❌write nav:", err)
		return ""
	}
	fmt.Println("✅nav merge result:", outPath)
	return outPath
}

func NavData(argv string) *domains.Nav {
	args, err := parseNavArgs(argv)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	nav := collectNav(args, false)
	printNavSummary(nav)
	SortEphemeris(nav)

	return nav
}

func parseNavArgs(argv string) (*navArgs, error) {
	parts := strings.Split(argv, "@")
	if len(parts) < 7 {
		return nil, fmt.Errorf("%s", navUsage)
	}
	startTime, err := utils.ParseTime(parts[0])
	if err != nil {
		return nil, err
	}
	endTime, err := utils.ParseTime(parts[1])
	if err != nil {
		return nil, err
	}
	return &navArgs{
		startTime:  startTime,
		endTime:    endTime,
		rawPath:    parts[2],
		outPath:    parts[3],
		sites:      splitSites(parts[4]),
		fileFormat: utils.Str2Int(parts[5]),
		ext:        parts[6],
	}, nil
}

func splitSites(sites string) []string {
	var siteList []string
	for _, site := range strings.Split(sites, ",") {
		site = strings.TrimSpace(site)
		if site != "" {
			siteList = append(siteList, site)
		}
	}
	return siteList
}

func collectNav(args *navArgs, logMessage bool) *domains.Nav {
	nav := make(domains.Nav)
	trTime := args.startTime
	rtcmFiles := utils.FindRtcmFiles(args.rawPath, args.startTime, args.endTime, args.sites, args.fileFormat, args.ext)
	for _, rtcmFile := range rtcmFiles {
		f, err := os.Open(rtcmFile)
		if err != nil {
			fmt.Println("❌open rtcm file:", err)
			continue
		}
		for {
			msgType, payload, err := ReadRTCM3Frames(f)
			if err != nil {
				if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
					break
				}
				//fmt.Println("read frame error:", err)
			}
			if payload == nil {
				continue
			}
			res := HandleMessage(msgType, payload, trTime)
			if res == nil {
				continue
			}
			if logMessage {
				fmt.Printf("✅message result: %+v\n", res)
			}
			if eph, ok := res.(*domains.Ephe); ok {
				trTime = eph.Toe.ToTime()
			}
			AddEphemeris(&nav, msgType, res)
		}
		_ = f.Close()
	}
	return &nav
}

func countNav(nav *domains.Nav) navStats {
	stats := navStats{}
	if nav == nil {
		return stats
	}
	for sat, ephs := range *nav {
		switch sat.Sys() {
		case 'G':
			stats.gps += len(ephs)
		case 'C':
			stats.bds += len(ephs)
		case 'R':
			stats.glonass += len(ephs)
		case 'E':
			stats.galileo += len(ephs)
		}
	}
	return stats
}

func printNavSummary(nav *domains.Nav) navStats {
	stats := countNav(nav)
	fmt.Printf("\n✅星历汇总:\nGPS星历: %d\nBDS星历: %d\nGLONASS星历: %d\nGalileo星历: %d\n",
		stats.gps, stats.bds, stats.glonass, stats.galileo)
	return stats
}

func navOutputPath(outRoot string, endTime time.Time) string {
	return filepath.Join(outRoot, fmt.Sprintf("%04d", endTime.Year()), fmt.Sprintf("BRDM%d0.rnx", endTime.YearDay()))
}
