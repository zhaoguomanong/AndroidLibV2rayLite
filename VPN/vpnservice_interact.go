package VPN

import (
	"github.com/2dust/AndroidLibV2rayLite/CoreI"
	"github.com/2dust/AndroidLibV2rayLite/Process/Escort"
	"golang.org/x/sys/unix"
	"v2ray.com/core/transport/internet"
)

type VPNSupport struct {
	prepareddomain preparedDomain
	VpnSupportSet  V2RayVPNServiceSupportsSet
	status         *CoreI.Status
	Estr           *Escort.Escorting
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
func (v *VPNSupport) VpnSupportReady(localDNS bool, enableIPv6 bool) {
	if !v.status.VpnSupportnodup {
		v.VpnSupportSet.Setup(v.status.GetVPNSetupArg(localDNS, enableIPv6))
		v.setV2RayDialer()
		v.startVPNRequire(localDNS, enableIPv6)
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

func (v *VPNSupport) startVPNRequire(localDNS bool, enableIPv6 bool) {
	//v.Estr = Escort.NewEscort()
	v.Estr.SetStatus(v.status)
	v.Estr.EscortingUPV()
	go v.Estr.EscortRun(
		v.status.GetApp("tun2socks"),
		v.status.GetTun2socksArgs(v.VpnSupportSet.GetVPNFd(), localDNS, enableIPv6),
		false,
		"")
}

func (v *VPNSupport) askSupportSetInit() {
	v.VpnSupportSet.Prepare()
}

func (v *VPNSupport) setV2RayDialer() {
	protectedDialer := &vpnProtectedDialer{vp: v}
	internet.UseAlternativeSystemDialer(internet.WithAdapter(protectedDialer))
}
