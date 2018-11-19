package CoreI

import (
	"v2ray.com/core"
)

type Status struct {
	IsRunning       bool
	VpnSupportnodup bool
	PackageName     string

	Vpoint core.Server
}

func (v *Status) GetDataDir() string {
	return v.getDataDir()
}

func (v *Status) getDataDir() string {
	return v.PackageName
}

func (v *Status) GetTun2socks() string {
	return v.PackageName + "tun2socks"
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