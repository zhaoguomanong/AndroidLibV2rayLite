package libv2ray

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/2dust/AndroidLibV2rayLite/CoreI"
	"github.com/2dust/AndroidLibV2rayLite/Process/Escort"
	"github.com/2dust/AndroidLibV2rayLite/shippedBinarys"
	mobasset "golang.org/x/mobile/asset"

	"v2ray.com/core"
	"v2ray.com/ext/sysio"
	"v2ray.com/core/features/stats"
	"v2ray.com/core/transport/internet"
	v2rayconf "v2ray.com/ext/tools/conf/serial"
)

const (
	v2Assert    = "v2ray.location.asset"
	assetperfix = "/dev/libv2rayfs0/asset"
)

/*V2RayPoint V2Ray Point Server
This is territory of Go, so no getter and setters!
*/
type V2RayPoint struct {
	Callbacks       V2RayCallbacks
	SupportSet  	V2RayVPNServiceSupportsSet
	statsManager 	stats.Manager

	status          *CoreI.Status
	escorter        *Escort.Escorting
	v2rayOP         *sync.Mutex

	//Legacy prop, should use Context instead
	PackageName          string
	DomainName           string
	ConfigureFileContent string
}

/*V2RayVPNServiceSupportsSet To support Android VPN mode*/
type V2RayVPNServiceSupportsSet interface {
	GetVPNFd() int
	Setup(Conf string) int
	Prepare() int
	Shutdown() int
	Protect(int) int
}

/*V2RayCallbacks a Callback set for V2Ray
 */
type V2RayCallbacks interface {
	OnEmitStatus(int, string) int
}

/*RunLoop Run V2Ray main loop
 */
func (v *V2RayPoint) RunLoop() (err error) {
	v.v2rayOP.Lock()
	defer v.v2rayOP.Unlock()
	//Construct Context
	v.status.PackageName = v.PackageName
	v.status.DomainName = v.DomainName

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
		err = v.stopLoopW()
	}
	return
}

func initV2Env() {
	if os.Getenv(v2Assert) != "" {
		return
	}
	//Initialize asset API, Since Raymond Will not let notify the asset location inside Process,
	//We need to set location outside V2Ray
	os.Setenv(v2Assert, assetperfix)
	//Now we handle read
	sysio.NewFileReader = func(path string) (io.ReadCloser, error) {
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
}

//Delegate Funcation
func TestConfig(ConfigureFileContent string) error {
	initV2Env()
	_, err := v2rayconf.LoadJSONConfig(strings.NewReader(ConfigureFileContent))
	return err
}

/*NewV2RayPoint new V2RayPoint*/
func NewV2RayPoint() *V2RayPoint {
	initV2Env()
	_status := &CoreI.Status{}
	return &V2RayPoint{
		v2rayOP:     new(sync.Mutex),
		status:      _status,
		escorter:    &Escort.Escorting{ Status: _status },
	}
}

//Delegate Funcation
func (v *V2RayPoint) GetIsRunning() bool {
	return v.status.IsRunning
}

//Delegate Funcation
func (v *V2RayPoint) VpnSupportReady(localDNS bool, enableIPv6 bool) {
	// APP VPNService establish
	v.SupportSet.Setup(v.status.GetVPNSetupArg(localDNS, enableIPv6))
	v.escorter.EscortingUp()
	go v.escorter.EscortRun(
		v.status.GetApp("tun2socks"),
		v.status.GetTun2socksArgs(v.SupportSet.GetVPNFd(), localDNS, enableIPv6),
		"")
}

//Delegate Funcation
func (v *V2RayPoint) SetVpnSupportSet(vs V2RayVPNServiceSupportsSet) {
	v.SupportSet = vs
}

//Delegate Funcation
func (v V2RayPoint) QueryStats(tag string, direct string) int64 {
	if v.statsManager == nil {
		return 0
	}
	query := fmt.Sprintf("inbound>>>%s>>>traffic>>>%s", tag, direct)
	counter := v.statsManager.GetCounter(query)
	if counter == nil {
		return 0
	}
	return counter.Value()
}

func (v *V2RayPoint) pointloop() error {

	//TODO:Parse Configure File
	log.Println("loading v2ray config")
	config, err := v2rayconf.LoadJSONConfig(strings.NewReader(v.ConfigureFileContent))
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
	inst, err := core.New(config)
	if err != nil {
		log.Println(err)
		return err
	}


	v.status.Vpoint = inst
	v.statsManager = inst.GetFeature(stats.ManagerType()).(stats.Manager)

	// v2ray hooker to protect fd
	protectfunc := func(network, address string, fd uintptr) error {
		if ret := v.SupportSet.Protect(int(fd)); ret != 0 {
			return fmt.ErrorF("protectfunc: fail to protect.")
		}
	}
	internet.RegisterDialerController(protectfunc)
	internet.RegisterListenerController(protectfunc)

	log.Println("start v2ray core")
	v.status.IsRunning = true
	if err := v.status.Vpoint.Start(); err != nil {
		v.status.IsRunning = false
		log.Println(err)
		return err
	}

	log.Println("run vpn apps")
	v.SupportSet.Prepare() // app will call V2rayPoint.VpnSupportReady
	v.Callbacks.OnEmitStatus(0, "Running")
	return nil
}

func (v *V2RayPoint) stopLoopW() (err error) {
	v.escorter.EscortingDown()
	v.SupportSet.Shutdown()
	v.status.IsRunning = false
	err = v.status.Vpoint.Close()
	v.Callbacks.OnEmitStatus(0, "Closed")
	return
}
