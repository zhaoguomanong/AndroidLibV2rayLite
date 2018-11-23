package CoreI

import (
	"v2ray.com/core"
)

type Status struct {
	IsRunning       bool
	VpnSupportnodup bool
	PackageName     string
	DomainName     string

	Vpoint core.Server
}

func CheckVersion() int {
	return 10
}

func (v *Status) GetDataDir() string {
	return v.getDataDir()
}

func (v *Status) getDataDir() string {
	return v.PackageName
}

func (v *Status) GetApp(name string) string {
	return v.PackageName + name
}

func (v *Status) GetTun2socksArgs() []string {
	return []string{"--netif-ipaddr",
                    "26.26.26.2",
                    "--netif-netmask",
                    "255.255.255.0",
                    "--socks-server-addr",
                    "127.0.0.1:10808",
                    "--tunfd",
                    "3",
                    "--tunmtu",
                    "1500",
                    "--sock-path",
                    "/dev/null",
                    "--loglevel",
                    "4",
                    "--enable-udprelay"}
}

func (v *Status) GetOvertureArgs() []string {
	var dynaArr []string
	dynaArr = append(dynaArr, "-c")
	dynaArr = append(dynaArr, v.PackageName + "config.json")
	dynaArr = append(dynaArr, "-l")
	dynaArr = append(dynaArr, v.PackageName + "overture.log")
	
	return dynaArr
}

func (v *Status) GetDomainNameList() []string {
	var dynaArr []string
	if v.DomainName != "" {
		dynaArr = append(dynaArr, v.DomainName)
	}	
	return dynaArr
}

