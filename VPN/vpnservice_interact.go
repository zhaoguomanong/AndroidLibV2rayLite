package VPN

import (
	"github.com/2dust/AndroidLibV2rayLite/CoreI"
	"github.com/2dust/AndroidLibV2rayLite/Process/Escort"
	"golang.org/x/sys/unix"
	"v2ray.com/core/transport/internet"
)

type VPNSupport struct {
	prepareddomain           preparedDomain
	VpnSupportSet            V2RayVPNServiceSupportsSet
	status                   *CoreI.Status
	Estr                     *Escort.Escorting
}

type V2RayVPNServiceSupportsSet interface {
	GetVPNFd() int
	Setup(Conf string) int
	Prepare() int
	Shutdown() int
	Protect(int) int
}

func (v *VPNSupport) SetStatus(st *CoreI.Status, estr *Escort.Escorting) {
	v.status = st
	v.Estr = estr
}

func (v *VPNSupport) VpnSetup() {
		v.prepareDomainName()
		v.askSupportSetInit()
}

/*VpnSupportReady VpnSupportReady*/
func (v *VPNSupport) VpnSupportReady() {
	if !v.status.VpnSupportnodup {
		v.VpnSupportSet.Setup(v.status.GetVPNSetupArg())
		v.setV2RayDialer()
		v.startVPNRequire()
	}
}

func (v *VPNSupport) VpnShutdown() {	
	//if v.VpnSupportnodup {
	err := unix.Close(v.VpnSupportSet.GetVPNFd())
	println(err)
	//}
	v.VpnSupportSet.Shutdown()
	v.status.VpnSupportnodup = false
}

func (v *VPNSupport) LoadLocalDns() {
	if !v.status.VpnSupportnodup {	
		//v.Estr = Escort.NewEscort()
		v.Estr.SetStatus(v.status)
		v.Estr.EscortingUPV()
		go v.Estr.EscortRun(v.status.GetApp("overture"), v.status.GetOvertureArgs(), false, 0)		
	}
}

func (v *VPNSupport) startVPNRequire() {
	//v.Estr = Escort.NewEscort()
	v.Estr.SetStatus(v.status)
	v.Estr.EscortingUPV()
	go v.Estr.EscortRun(v.status.GetApp("tun2socks"), v.status.GetTun2socksArgs(v.VpnSupportSet.GetVPNFd()), false, v.VpnSupportSet.GetVPNFd())	
}

func (v *VPNSupport) askSupportSetInit() {
	v.VpnSupportSet.Prepare()
}

func (v *VPNSupport) setV2RayDialer() {
	protectedDialer := &vpnProtectedDialer{vp: v}
	internet.UseAlternativeSystemDialer(internet.WithAdapter(protectedDialer))
}