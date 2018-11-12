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
	
	/*
	var datadir = "/data/data/com.v2ray.ang/"
	if v.PackageName != "" {
		datadir = "/data/user/" + strconv.Itoa(os.Getuid()/100000) + "/" + v.PackageName + "/"
		if !Exists(datadir) {
			datadir = "/data/data/" + v.PackageName + "/"
		}
	}
	return datadir
	*/
}