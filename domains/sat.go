package domains

import (
	"sort"
	"strings"
)

// SatType Type representing satellite name like "G10"
type SatType string

type SatVar []SatType

func (p *SatVar) Set(s string) error {
	*p = []SatType{}
	for _, a := range strings.Split(s, ",") {
		*p = append(*p, SatType(a))
	}
	return nil
}

func (p *SatVar) String() string {
	return ""
}

// Sorted Sort the list of satellite names
func Sorted(s []SatType) []SatType {
	s2 := make([]SatType, len(s))
	copy(s2, s)
	sort.Slice(s2, func(i, j int) bool {
		m := map[byte]int{'G': 0, 'J': 1, 'E': 2, 'R': 3, 'C': 4, 'S': 5}
		if m[s2[i][0]] == m[s2[j][0]] {
			return s2[i] < s2[j]
		} else {
			return m[s2[i][0]] < m[s2[j][0]]
		}
	})
	return s2
}

// SatPairFType represents a satellite pair and frequency for double-difference processing
// It holds the reference satellite (highest elevation) and paired satellite information
type SatPairFType struct {
	S1 SatType // Reference satellite (highest elevation satellite)
	S2 SatType // Paired satellite for double-difference
	F  int     // Frequency index (0, 1, 2, 3)
}

// SatFType represents a satellite and frequency for single-difference ambiguity processing
// It holds satellite identification and frequency information
type SatFType struct {
	Sat SatType // Satellite identifier
	F   int     // Frequency index (0, ..., NFREQ-1)
}

func SortedSatF(s []SatFType) []SatFType {
	s2 := make([]SatFType, len(s))
	copy(s2, s)
	sort.Slice(s2, func(i, j int) bool {
		m := map[byte]int{'G': 0, 'J': 1, 'E': 2, 'R': 3, 'C': 4, 'S': 5}
		if s2[i].F == s2[j].F {
			if s2[i].Sat.Sys() == s2[j].Sat.Sys() {
				return s2[i].Sat.Num() < s2[j].Sat.Num()
			} else {
				return m[byte(s2[i].Sat.Sys())] < m[byte(s2[j].Sat.Sys())]
			}
		} else {
			return s2[i].F < s2[j].F
		}
	})
	return s2
}

// CodeType Type representing observation codes like C1C (3 or 2 characters)
type CodeType string

// T Returns observation type (C,L,D,S)
func (p *CodeType) T() byte {
	return (*p)[0]
}

// NA Returns frequency band and attributes of observation (1C,2P,5I etc.)
func (p *CodeType) NA() CodeType {
	return CodeType(*p)[1:]
}
