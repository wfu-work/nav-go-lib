package libs

import (
	"sort"
	"time"

	"navfirst.com/nav-go-lib/domains"
)

func init() {
	RegisterNavParsers()
}

func HandleMessage(msgType uint16, payload []byte, endTime time.Time) any {
	if p, ok := parserRegistry[msgType]; ok {
		//fmt.Printf("Message Type: %d, payload bytes: %d\n", msgType, len(payload))
		r := p(msgType, payload, endTime)
		return r
	}
	return nil
}

func AddEphemeris(nav *domains.Nav, msgType uint16, res any) {
	switch msgType {
	case 1019, 1020, 1042, 1045, 1046: // GPS, GLONASS, BDS, Galileo
		if eph, ok := res.(*domains.Ephe); ok {
			if !isEphemerisExists(*nav, eph) {
				(*nav)[eph.Sat] = append((*nav)[eph.Sat], eph)
			}
		}
	}
}

// Check if ephemeris already exists based on Sat, Iode, and Toe
func isEphemerisExists(nav domains.Nav, newEph *domains.Ephe) bool {
	if ephs, ok := nav[newEph.Sat]; ok {
		for _, existingEph := range ephs {
			if existingEph.Toe.Week == newEph.Toe.Week &&
				existingEph.Toe.Sec == newEph.Toe.Sec {
				return true
			}
		}
	}
	return false
}

func SortEphemeris(nav *domains.Nav) {
	for sat, ephs := range *nav {
		sort.Slice(ephs, func(i, j int) bool {
			return ephs[i].Toe.Less(ephs[j].Toe, false)
		})
		(*nav)[sat] = ephs
	}
}
