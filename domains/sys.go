package domains

import (
	"strconv"
	"strings"
)

// SysType Type representing satellite system like 'G'
type SysType byte

// Sys Extract satellite system from satellite name
func (p *SatType) Sys() SysType {
	return SysType((*p)[0])
}

// IsValid Check validity of satellite system
func (p *SysType) IsValid() bool {
	return *p == 'G' || *p == 'J' || *p == 'E' || *p == 'R' || *p == 'C' || *p == 'S'
}

// Num Extract satellite number from satellite name
func (p *SatType) Num() int {
	i, err := strconv.Atoi(string((*p)[1:3]))
	if err != nil {
		return 0
	}
	return i
}

type SysTypeList []SysType

func (p *SysTypeList) Set(s string) error {
	*p = []SysType{}
	for _, a := range strings.Split(s, ",") {
		*p = append(*p, SysType(a[:][0]))
	}
	return nil
}

func (p *SysTypeList) String() string {
	*p = []SysType{'G', 'J', 'E', 'R', 'C'}
	return ""
}

func (p *SysTypeList) Contains(s SysType) bool {
	for _, v := range *p {
		if s == v {
			return true
		}
	}
	return false
}

// ParseNavSys converts nav system string like "G,R,C" into []SysType
func ParseNavSys(nav string) []SysType {
	if nav == "" {
		return nil
	}
	parts := strings.Split(nav, ",")
	var sys []SysType
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if len(p) > 0 {
			s := SysType(p[0])
			if (&s).IsValid() {
				sys = append(sys, s)
			}
		}
	}
	return sys
}
