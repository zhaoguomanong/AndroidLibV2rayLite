package libv2ray

import (
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"v2ray.com/core"
	"v2ray.com/ext/sysio"
	"github.com/2dust/AndroidLibV2rayLite/CoreI"
	"github.com/2dust/AndroidLibV2rayLite/VPN"
	"github.com/2dust/AndroidLibV2rayLite/shippedBinarys"
	"github.com/2dust/AndroidLibV2rayLite/Process/Escort"
	mobasset "golang.org/x/mobile/asset"
	v2rayconf "v2ray.com/ext/tools/conf/serial"
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

	//Legacy prop, should use Context instead
	PackageName          string
	DomainName		     string
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
func (v *V2RayPoint) RunLoop() {
	v.v2rayOP.Lock()
	//Construct Context
	v.status.PackageName = v.PackageName
	v.status.DomainName = v.DomainName
	
	if !v.status.IsRunning {
		go v.pointloop()
	}
	v.v2rayOP.Unlock()
}


/*StopLoop Stop V2Ray main loop
 */
func (v *V2RayPoint) StopLoop() {
	v.v2rayOP.Lock()
	if v.status.IsRunning {
		/* TODO: Shutdown VPN
		v.vpnShutdown()
		*/
		go v.stopLoopW()
	}
	v.v2rayOP.Unlock()
}

/*NewV2RayPoint new V2RayPoint*/
func NewV2RayPoint() *V2RayPoint {
	//Initialize asset API, Since Raymond Will not let notify the asset location inside Process,
	//We need to set location outside V2Ray
	const assetperfix = "/dev/libv2rayfs0/asset"
	os.Setenv("v2ray.location.asset", assetperfix)
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
	//platform.ForceReevaluate()
	//panic("Creating VPoint")
	return &V2RayPoint{v2rayOP: new(sync.Mutex), status: &CoreI.Status{}, escorter: Escort.NewEscort(), VPNSupports: &VPN.VPNSupport{}}
} 

func (v *V2RayPoint) GetIsRunning() bool {
	return v.status.IsRunning
}

//Delegate Funcation
func (v *V2RayPoint) VpnSupportReady() {
	v.VPNSupports.VpnSupportReady()
}

//Delegate Funcation
func (v *V2RayPoint) SetVpnSupportSet(vs V2RayVPNServiceSupportsSet) {
	v.VPNSupports.VpnSupportSet = vs
}

func (v *V2RayPoint) pointloop() {
	v.status.VpnSupportnodup = false

	//TODO:Parse Configure File
	log.Println("loading v2ray config")
	var config core.Config
	configx, _ := v2rayconf.LoadJSONConfig(strings.NewReader(v.ConfigureFileContent))
	config = *configx
	
	var err error
	//TODO:Load Shipped Binary
	shipb := shippedBinarys.FirstRun{}
	shipb.SetCoreI(v.status)
	err = shipb.CheckAndExport()
	if err != nil {
		log.Println(err)
	}

	//New Start V2Ray Core
	log.Println("new v2ray core")
	v.status.Vpoint, err = core.New(&config)
	if err != nil {
		log.Println("VPoint Start Err:" + err.Error())

	}
	
	log.Println("start v2ray core")
	v.status.IsRunning = true
	v.status.Vpoint.Start()

	v.interuptDeferto = 1
	
	go func() {
		time.Sleep(5 * time.Second)
		v.interuptDeferto = 0
	}()
	//Set Necessary Props First
	
	log.Println("run vpn apps")

	v.VPNSupports.SetStatus(v.status)
	v.VPNSupports.VpnSetup()
 	
	v.Callbacks.OnEmitStatus(0, "Running")
}

func (v *V2RayPoint) stopLoopW() {
	v.status.IsRunning = false
	v.status.Vpoint.Close()	 
	v.VPNSupports.VpnShutdown()
	v.escorter.EscortingDown()	
	v.Callbacks.OnEmitStatus(0, "Closed")

}
