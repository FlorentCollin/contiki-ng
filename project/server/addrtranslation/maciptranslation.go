package addrtranslation

import (
	"net"
	"strings"
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
	ip  IPString
}

type MacIPTranslation struct {
	MacIPaddrs []*MacIPPair
}

func NewMacIPTranslation() MacIPTranslation {
	return MacIPTranslation{MacIPaddrs: make([]*MacIPPair, 0)}
}

func (macIPTranslation *MacIPTranslation) Add(macAddr *MacAddr, ip IPString) {
	macIPTranslation.MacIPaddrs = append(macIPTranslation.MacIPaddrs, &MacIPPair{
		mac: macAddr,
		ip:  ip,
	})
}

func (macIPTranslation *MacIPTranslation) Find(macAddr *MacAddr) (IPString, bool) {
	for _, macip := range macIPTranslation.MacIPaddrs {
		if macip.mac.Equal(macAddr) {
			return macip.ip, true
		}
	}
	return "", false
}

func (macIPTranslation *MacIPTranslation) FindMac(ipAddr IPString) (*MacAddr, bool) {
	for _, macip := range macIPTranslation.MacIPaddrs {
		if macip.ip == ipAddr {
			return macip.mac, true
		}
	}
	return nil, false
}

type IPString string

func AddrToIPString(addr *net.UDPAddr) IPString {
	return IPString(addr.IP.String())
}

func NetIPToIPString(addr net.IP) IPString {
	return IPString(addr.String())
}

func (ipString IPString) GlobalToLinkLocal() IPString {
	// TODO: Normally this should not be done, we need a better way to translate link local to global addresses.
	return IPString(strings.Replace(string(ipString), "fd00", "fe80", 1))
}

func (ipString IPString) LinkLocalToGlobal() IPString {
	// TODO: This is a kind of hack since normally we should not be able to get the local address from the link-local addr
	// This certainly only works in Cooja
	return IPString(strings.Replace(string(ipString), "fe80", "fd00", 1))
}
