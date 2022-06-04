package addrtranslation

// This package allows to keep a one-to-one relation between an IP address of a node and its LinkLocalAddr.

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
	if _, in := macIPTranslation.Find(macAddr); !in {
		macIPTranslation.MacIPaddrs = append(macIPTranslation.MacIPaddrs, &MacIPPair{
			mac: macAddr,
			ip:  ip,
		})
	}
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

func (ipString IPString) LinkLocalToGlobal() IPString {
	// Normally this should not be done, we need a better way to translate link local to global addresses.
	// For simulations though it's totally usable because the translation between an IP address and the LinkLocal
	// address is deterministic.
	return IPString(strings.Replace(string(ipString), "fe80", "fd00", 1))
}
