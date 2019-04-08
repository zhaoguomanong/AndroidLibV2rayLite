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
		var err error
		v.prepareddomain.tcpprepared[domainName], err = net.ResolveTCPAddr("tcp", domainName)
		if err != nil {
			log.Println(err)
		}
		v.prepareddomain.udpprepared[domainName], err = net.ResolveUDPAddr("udp", domainName)
		if err != nil {
			log.Println(err)
		}
	}
}
