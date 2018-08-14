package net

import (
	"fmt"

	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
)

func ParseMultiaddr(m ma.Multiaddr) (laddr string, code int, err error) {
	if !manet.IsThinWaist(m) {
		err = fmt.Errorf("%s is not a 'thin waist' address", m)
		return
	}

	var addr string
	var port string
	for _, p := range m.Protocols() {
		switch p.Code {
		case ma.P_IP4:
			addr, err = m.ValueForProtocol(ma.P_IP4)
		case ma.P_UDP:
			code = ma.P_UDP
			port, err = m.ValueForProtocol(ma.P_UDP)
		case ma.P_TCP:
			code = ma.P_TCP
			port, err = m.ValueForProtocol(ma.P_TCP)
		default:
			err = fmt.Errorf("not supported `%s`", p.Name)
		}

		if err != nil {
			return
		}
	}

	laddr = addr + ":" + port
	return
}
