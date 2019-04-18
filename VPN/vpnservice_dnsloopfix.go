package VPN

import (
	"log"
	"net"
)

type preparedDomain struct {
	tcpprepared map[string](*net.TCPAddr)
	udpprepared map[string](*net.UDPAddr)
}

func (p *preparedDomain) Init() {
	p.tcpprepared = make(map[string](*net.TCPAddr))
	p.udpprepared = make(map[string](*net.UDPAddr))
}

func (v *VPNSupport) prepareDomainName() {
	if v.VpnSupportSet == nil {
		return
	}
	for _, domainName := range v.status.GetDomainNameList() {
		log.Println("Preparing DNS,", domainName)
		if addr, err := net.ResolveTCPAddr("tcp", domainName); err == nil {
			log.Println("Prepared %s, %s,", domainName, addr)
			v.prepareddomain.tcpprepared[domainName] = addr
		} else {
			log.Println(err)
		}

		if addr, err := net.ResolveUDPAddr("udp", domainName); err == nil {
			log.Println("Prepared %s, %s,", domainName, addr)
			v.prepareddomain.udpprepared[domainName] = addr
		} else {
			log.Println(err)
		}
	}
}
