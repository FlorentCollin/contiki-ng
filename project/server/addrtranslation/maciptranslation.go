package addrtranslation

import (
	"coap-server/udpack"
)

type MacAddr [8]byte

func (macaddr *MacAddr) Equal(other *MacAddr) bool {
	for i, b := range macaddr {
		if b != other[i] {
			return false
		}
	}
	return true
}

type MacIPPair struct {
	mac *MacAddr
	ip  udpack.IPString
}

type MacIPTranslation struct {
	MacIPaddrs []*MacIPPair
}

func NewMacIPTranslation() MacIPTranslation {
	return MacIPTranslation{MacIPaddrs: make([]*MacIPPair, 0)}
}

func (macIPTranslation *MacIPTranslation) Add(macAddr *MacAddr, ip udpack.IPString) {
	macIPTranslation.MacIPaddrs = append(macIPTranslation.MacIPaddrs, &MacIPPair{
		mac: macAddr,
		ip:  ip,
	})
}

func (macIPTranslation *MacIPTranslation) Find(macAddr *MacAddr) (udpack.IPString, bool) {
	for _, macip := range macIPTranslation.MacIPaddrs {
		if macip.mac.Equal(macAddr) {
			return macip.ip, true
		}
	}
	return "", false
}

func (macIPTranslation *MacIPTranslation) FindMac(ipAddr udpack.IPString) (*MacAddr, bool) {
	for _, macip := range macIPTranslation.MacIPaddrs {
		if macip.ip == ipAddr {
			return macip.mac, true
		}
	}
	return nil, false
}
