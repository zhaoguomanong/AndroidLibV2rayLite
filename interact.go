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
	"github.com/2dust/AndroidLibV2rayLite/VPN"
	"github.com/2dust/AndroidLibV2rayLite/shippedBinarys"
	mobasset "golang.org/x/mobile/asset"
	"v2ray.com/core"
	"v2ray.com/ext/sysio"
	v2rayconf "v2ray.com/ext/tools/conf/serial"

	"v2ray.com/core/features/stats"
)

const (
	v2Assert    = "v2ray.location.asset"
	assetperfix = "/dev/libv2rayfs0/asset"
)

/*V2RayPoint V2Ray Point Server
This is territory of Go, so no getter and setters!
*/
type V2RayPoint struct {
	status          *CoreI.Status
	escorter        *Escort.Escorting
	Callbacks       V2RayCallbacks
	v2rayOP         *sync.Mutex
	VPNSupports     *VPN.VPNSupport
	interuptDeferto int64

	StatsManager stats.Manager
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
	//Construct Context
	v.status.PackageName = v.PackageName
	v.status.DomainName = v.DomainName

	if !v.status.IsRunning {
		err = v.pointloop()
	}
	v.v2rayOP.Unlock()
	return
}

/*StopLoop Stop V2Ray main loop
 */
func (v *V2RayPoint) StopLoop() (err error) {
	v.v2rayOP.Lock()
	if v.status.IsRunning {
		err = v.stopLoopW()
	}
	v.v2rayOP.Unlock()
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
	return &V2RayPoint{
		v2rayOP:     new(sync.Mutex),
		status:      &CoreI.Status{},
		escorter:    Escort.NewEscort(),
		VPNSupports: &VPN.VPNSupport{},
	}
}

func (v *V2RayPoint) GetIsRunning() bool {
	return v.status.IsRunning
}

//Delegate Funcation
func (v *V2RayPoint) VpnSupportReady(localDNS bool, enableIPv6 bool) {
	v.VPNSupports.VpnSupportReady(localDNS, enableIPv6)
}

//Delegate Funcation
func (v *V2RayPoint) SetVpnSupportSet(vs V2RayVPNServiceSupportsSet) {
	v.VPNSupports.VpnSupportSet = vs
}

//Delegate Funcation
func (v V2RayPoint) QueryStats(tag string, direct string) int64 {
	query := fmt.Sprintf("inbound>>>%s>>>traffic>>>%s", tag, direct)
	counter := v.StatsManager.GetCounter(query)
	if counter != nil {
		return counter.Value()
	}
	return 0
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
	shipb := shippedBinarys.FirstRun{}
	shipb.SetCoreI(v.status)
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
	v.StatsManager = inst.GetFeature(stats.ManagerType()).(stats.Manager)

	log.Println("start v2ray core")
	v.status.IsRunning = true
	if err := v.status.Vpoint.Start(); err != nil {
		v.status.IsRunning = false
		log.Println(err)
		return err
	}

	//Set Necessary Props First
	log.Println("run vpn apps")

	v.VPNSupports.SetStatus(v.status, v.escorter)
	v.VPNSupports.VpnSetup()

	v.Callbacks.OnEmitStatus(0, "Running")

	return nil
}

func (v *V2RayPoint) stopLoopW() (err error) {
	v.status.IsRunning = false
	err = v.status.Vpoint.Close()
	v.escorter.EscortingDown()
	v.VPNSupports.VpnShutdown()
	v.Callbacks.OnEmitStatus(0, "Closed")
	return
}
