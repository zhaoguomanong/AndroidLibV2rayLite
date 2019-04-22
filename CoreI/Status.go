package CoreI

import (
	"strconv"

	v2core "v2ray.com/core"
)

type Status struct {
	IsRunning   bool
	PackageName string
	DomainName  string

	Vpoint v2core.Server
}

func CheckVersion() int {
	return 19
}

func (v *Status) GetDataDir() string {
	return v.PackageName
}

func (v *Status) GetApp(name string) string {
	return v.PackageName + name
}

func (v *Status) GetTun2socksArgs(fd int, localDNS bool, enableIPv6 bool) (ret []string) {
	ret = []string{"--netif-ipaddr",
		"26.26.26.2",
		"--netif-netmask",
		"255.255.255.252",
		"--socks-server-addr",
		"127.0.0.1:10808",
		"--tunfd",
		strconv.Itoa(fd),
		"--tunmtu",
		"1500",
		"--loglevel",
		"info",
		"--enable-udprelay"}

	if enableIPv6 {
		ret = append(ret, "--netif-ip6addr", "fd26:2626::2")
	}

	if localDNS {
		ret = append(ret, "--dnsgw", "127.0.0.1:10807")
	}

	return
}

func (v *Status) GetVPNSetupArg(localDNS bool, enableIPv6 bool) (ret string) {
	ret = "m,1500 a,26.26.26.1,30 r,0.0.0.0,0"

	if enableIPv6 {
		ret += " a,fd26:2626::1,126 r,::,0"
	}
	if localDNS {
		ret += " d,26.26.26.2"
	}
	return
}
