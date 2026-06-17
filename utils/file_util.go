package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FindRawFiles 查询文件
func FindRawFiles(root string, start, end time.Time, site string, fileFormat int, ext string) []string {
	var files []string
	cur := start
	for cur.Before(end) {
		dir := rawDir(root, cur, fileFormat)
		fileExt := ext
		if fileExt == "" {
			fileExt = getDefaultExt(cur, fileFormat)
		}
		findFileName := site + fileExt
		if _, err := os.Stat(dir); err == nil {
			_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return nil
				}
				if d.IsDir() {
					return nil
				}
				filename := filepath.Base(path)
				if filename != findFileName {
					return nil
				}
				files = append(files, path)
				return nil
			})
		}
		if fileFormat == 1 {
			cur = cur.Add(time.Hour)
		} else {
			cur = cur.Add(time.Hour * 24)
		}
	}
	sort.Strings(files)
	return files
}

// FindRtcmFiles 查询RTCM文件
func FindRtcmFiles(root string, start, end time.Time, sites []string, fileFormat int, ext string) []string {
	var files []string
	cur := start
	for cur.Before(end) {
		dir := rawDir(root, cur, fileFormat)
		fileExt := ext
		if fileExt == "" {
			fileExt = getDefaultExt(cur, fileFormat)
		}
		if _, err := os.Stat(dir); err == nil {
			_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return nil
				}
				if d.IsDir() {
					return nil
				}
				filename := filepath.Base(path)
				if !strings.HasSuffix(filename, fileExt) {
					return nil
				}
				if len(sites) > 0 {
					match := false
					for _, s := range sites {
						if strings.HasPrefix(filename, s) {
							match = true
							break
						}
					}
					if !match {
						return nil
					}
				}
				files = append(files, path)
				return nil
			})
		}
		if fileFormat == 1 {
			cur = cur.Add(time.Hour)
		} else {
			cur = cur.Add(time.Hour * 24)
		}
	}
	sort.Strings(files)
	return files
}

func rawDir(root string, cur time.Time, fileFormat int) string {
	year := cur.Year()
	dayOfYear := cur.YearDay()
	hour := cur.Hour()
	switch fileFormat {
	case 1:
		return filepath.Join(
			root,
			fmt.Sprintf("%04d", year),
			fmt.Sprintf("%03d", dayOfYear),
			fmt.Sprintf("%02d", hour),
		)
	case 2:
		return filepath.Join(
			root,
			fmt.Sprintf("%04d", year),
			fmt.Sprintf("%03d", dayOfYear),
		)
	default:
		return root
	}
}

func getDefaultExt(cur time.Time, fileFormat int) string {
	if fileFormat == 1 {
		return fmt.Sprintf(".%d%03dbinRTCM3", cur.Year(), cur.YearDay())
	}
	if fileFormat == 2 {
		return fmt.Sprintf(".%do", cur.Year())
	}
	return fmt.Sprintf(".%d%03dbinRTCM3", cur.Year(), cur.YearDay())
}
