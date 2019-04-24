package libv2ray

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/2dust/AndroidLibV2rayLite/CoreI"
	"github.com/2dust/AndroidLibV2rayLite/Process/Escort"
	"github.com/2dust/AndroidLibV2rayLite/VPN"
	"github.com/2dust/AndroidLibV2rayLite/shippedBinarys"
	mobasset "golang.org/x/mobile/asset"

	v2core "v2ray.com/core"
	v2filesystem "v2ray.com/core/common/platform/filesystem"
	v2stats "v2ray.com/core/features/stats"
	v2serial "v2ray.com/core/infra/conf/serial"
	_ "v2ray.com/core/main/distro/all"
	v2internet "v2ray.com/core/transport/internet"
)

const (
	v2Assert    = "v2ray.location.asset"
	assetperfix = "/dev/libv2rayfs0/asset"
)

/*V2RayPoint V2Ray Point Server
This is territory of Go, so no getter and setters!
*/
type V2RayPoint struct {
	SupportSet   V2RayVPNServiceSupportsSet
	statsManager v2stats.Manager

	status   *CoreI.Status
	escorter *Escort.Escorting
	v2rayOP  *sync.Mutex

	PackageName          string
	DomainName           string
	ConfigureFileContent string
	EnableLocalDNS       bool
	ForwardIpv6          bool
}

/*V2RayVPNServiceSupportsSet To support Android VPN mode*/
type V2RayVPNServiceSupportsSet interface {
	GetVPNFd() int
	Setup(Conf string) int
	Prepare() int
	Shutdown() int
	Protect(int) int
	OnEmitStatus(int, string) int
}

/*RunLoop Run V2Ray main loop
 */
func (v *V2RayPoint) RunLoop() (err error) {
	v.v2rayOP.Lock()
	defer v.v2rayOP.Unlock()
	//Construct Context
	v.status.PackageName = v.PackageName

	if !v.status.IsRunning {
		err = v.pointloop()
	}
	return
}

/*StopLoop Stop V2Ray main loop
 */
func (v *V2RayPoint) StopLoop() (err error) {
	v.v2rayOP.Lock()
	defer v.v2rayOP.Unlock()
	if v.status.IsRunning {
		v.status.IsRunning = false
		go v.status.Vpoint.Close()
		go v.escorter.EscortingDown()
		v.SupportSet.Shutdown()
		v.SupportSet.OnEmitStatus(0, "Closed")
	}
	return
}

//Delegate Funcation
func (v *V2RayPoint) GetIsRunning() bool {
	return v.status.IsRunning
}

//Delegate Funcation
func (v V2RayPoint) QueryStats(tag string, direct string) int64 {
	if v.statsManager == nil {
		return 0
	}
	counter := v.statsManager.GetCounter(fmt.Sprintf("inbound>>>%s>>>traffic>>>%s", tag, direct))
	if counter == nil {
		return 0
	}
	return counter.Value()
}

func (v V2RayPoint) protectFd(network, address string, fd uintptr) error {
	if ret := v.SupportSet.Protect(int(fd)); ret != 0 {
		return fmt.Errorf("protectFd: fail to protect")
	}
	return nil
}

func (v *V2RayPoint) pointloop() error {
	log.Printf("EnableLocalDNS: %v\nForwardIpv6: %v\nDomainName: %s",
		v.EnableLocalDNS,
		v.ForwardIpv6,
		v.DomainName)

	dialer := &VPN.VPNProtectedDialer{
		DomainName: v.DomainName,
		SupportSet: v.SupportSet,
	}
	pch := make(chan bool)
	go dialer.PrepareDomain(pch)

	//TODO:Parse Configure File
	log.Println("loading v2ray config")
	config, err := v2serial.LoadJSONConfig(strings.NewReader(v.ConfigureFileContent))
	if err != nil {
		log.Println(err)
		return err
	}

	//TODO:Load Shipped Binary
	shipb := shippedBinarys.FirstRun{Status: v.status}
	if err := shipb.CheckAndExport(); err != nil {
		log.Println(err)
		return err
	}

	//New Start V2Ray Core
	log.Println("new v2ray core")
	inst, err := v2core.New(config)
	if err != nil {
		log.Println(err)
		return err
	}

	v.status.Vpoint = inst
	v.statsManager = inst.GetFeature(v2stats.ManagerType()).(v2stats.Manager)

	log.Println("start v2ray core")
	v.status.IsRunning = true
	if err := v.status.Vpoint.Start(); err != nil {
		v.status.IsRunning = false
		log.Println(err)
		return err
	}

	v.SupportSet.Prepare()
	select {
	case <-pch: // block until ready
	case <-time.After(3 * time.Second):
	}
	if !dialer.PreparedReady {
		v.SupportSet.OnEmitStatus(0, "Closed")
		return nil
	}
	v.SupportSet.Setup(v.status.GetVPNSetupArg(v.EnableLocalDNS, v.ForwardIpv6)) // vpnservice.establish()

	log.Println("run vpn apps")
	v.runTun2socks()
	// v2ray hooker to protect fd
	v2internet.UseAlternativeSystemDialer(dialer)
	v.SupportSet.OnEmitStatus(0, "Running")
	return nil
}

func initV2Env() {
	if os.Getenv(v2Assert) != "" {
		return
	}
	//Initialize asset API, Since Raymond Will not let notify the asset location inside Process,
	//We need to set location outside V2Ray
	os.Setenv(v2Assert, assetperfix)
	//Now we handle read
	v2filesystem.NewFileReader = func(path string) (io.ReadCloser, error) {
		if strings.HasPrefix(path, assetperfix) {
			p := path[len(assetperfix)+1:]
			//is it overridden?
			//by, ok := overridedAssets[p]
			//if ok {
			//	return os.Open(by)
			//}
			return mobasset.Open(p)
		}
		return os.Open(path)
	}

	// opt-in TLS 1.3 for Go1.12
	// TODO: remove this line when Go1.13 is released.
	os.Setenv("GODEBUG", "tls13=1")
}

//Delegate Funcation
func TestConfig(ConfigureFileContent string) error {
	initV2Env()
	_, err := v2serial.LoadJSONConfig(strings.NewReader(ConfigureFileContent))
	return err
}

/*NewV2RayPoint new V2RayPoint*/
func NewV2RayPoint() *V2RayPoint {
	initV2Env()
	_status := &CoreI.Status{}
	return &V2RayPoint{
		v2rayOP:  new(sync.Mutex),
		status:   _status,
		escorter: &Escort.Escorting{Status: _status},
	}
}

func (v V2RayPoint) runTun2socks() {
	// APP VPNService establish
	v.escorter.EscortingUp()
	go v.escorter.EscortRun(
		v.status.GetApp("tun2socks"),
		v.status.GetTun2socksArgs(v.SupportSet.GetVPNFd(), v.EnableLocalDNS, v.ForwardIpv6),
		"")
}

/*CheckVersion int
This func will return libv2ray binding version.
*/
func CheckVersion() int {
	return CoreI.CheckVersion()
}

/*CheckVersionX string
This func will return libv2ray binding version and V2Ray version used.
*/
func CheckVersionX() string {
	return fmt.Sprintf("Libv2rayLite V%d, Core V%s", CheckVersion(), v2core.Version())
}
