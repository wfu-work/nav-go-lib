package main

import (
	"fmt"
	"os"
	"strings"

	"navfirst.com/nav-go-lib/libs"
)

func main() {
	if len(os.Args) != 2 || isHelpArg(os.Args[1]) {
		usage()
		if len(os.Args) == 2 && isHelpArg(os.Args[1]) {
			return
		}
		os.Exit(2)
	}

	argv := os.Args[1]
	if !isValidNavArgs(argv) {
		_, _ = fmt.Fprintln(os.Stderr, "invalid nav args: expected <startTime>@<endTime>@<rawPath>@<outPath>@<sites>@<fileFormat>@<ext>")
		_, _ = fmt.Fprintln(os.Stderr)
		usage()
		os.Exit(2)
	}

	outPath := libs.Nav(argv)
	fmt.Println(outPath)
}

func isHelpArg(arg string) bool {
	return arg == "-h" || arg == "--help" || arg == "help"
}

func isValidNavArgs(argv string) bool {
	return strings.Count(argv, "@") >= 6
}

func usage() {
	_, _ = fmt.Fprint(os.Stderr, `nav-go-lib merges RTCM3 navigation messages into RINEX NAV.

Usage:
  go run ./cmd "<startTime>@<endTime>@<rawPath>@<outPath>@<sites>@<fileFormat>@<ext>"

字段说明:
  startTime   开始时间，UTC 文本时间，格式: 2006/01/02 15:04:05
  endTime     结束时间，UTC 文本时间，格式: 2006/01/02 15:04:05
  rawPath     RTCM 原始文件根目录
  outPath     RINEX NAV 输出根目录
  sites       站点名前缀，多个站点用英文逗号分隔
  fileFormat  原始文件目录格式: 0=<rawPath>, 1=<rawPath>/<YYYY>/<DOY>/<HH>, 2=<rawPath>/<YYYY>/<DOY>
  ext         指定文件后缀；为空时使用默认后缀

Example:
  go run ./cmd "2025/12/07 00:00:00@2025/12/07 09:00:00@/data/raw@/data/nav@PSNMJC014,PSNMJC074@1@"
`)
}
