package VPN

import (

	"github.com/2dust/AndroidLibV2rayLite/CoreI"
	"github.com/2dust/AndroidLibV2rayLite/Process/Escort"

	"golang.org/x/sys/unix"

	"v2ray.com/core/transport/internet"
)

/*VpnSupportReady VpnSupportReady*/
func (v *VPNSupport) VpnSupportReady() {
	if !v.status.VpnSupportnodup {		
		VPNSetupArg := "m,1500 a,26.26.26.1,24 r,0.0.0.0,0"
		v.VpnSupportSet.Setup(VPNSetupArg)
		v.setV2RayDialer()
		v.startVPNRequire()
	}
}
func (v *VPNSupport) startVPNRequire() {
	v.Estr = Escort.NewEscort()
	v.Estr.SetStatus(v.status)
	v.Estr.EscortingUPV()
	go v.Estr.EscortRun(v.status.GetApp("tun2socks"), v.status.GetTun2socksArgs(), false, v.VpnSupportSet.GetVPNFd())	
	//go v.Estr.EscortRun(v.status.GetApp2("overture"), []string{}, false, 0)	
}

func (v *VPNSupport) askSupportSetInit() {
	v.VpnSupportSet.Prepare()
}

func (v *VPNSupport) VpnSetup() {
		v.prepareDomainName()
		v.askSupportSetInit()
}
func (v *VPNSupport) VpnShutdown() {	
	//if v.VpnSupportnodup {
	err := unix.Close(v.VpnSupportSet.GetVPNFd())
	println(err)
	//}
	v.VpnSupportSet.Shutdown()
	v.status.VpnSupportnodup = false
}

func (v *VPNSupport) setV2RayDialer() {
	protectedDialer := &vpnProtectedDialer{vp: v}
	internet.UseAlternativeSystemDialer(internet.WithAdapter(protectedDialer))
}

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

func (v *VPNSupport) SetStatus(st *CoreI.Status) {
	v.status = st
}
