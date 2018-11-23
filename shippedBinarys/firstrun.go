package shippedBinarys

import (	
	"log"
	"os"
	
	"github.com/2dust/AndroidLibV2rayLite/CoreI"
)

type FirstRun struct {
	status *CoreI.Status
}

func (v *FirstRun) checkIfRcExist() error {
	datadir := v.status.GetDataDir()	
	if _, err := os.Stat(datadir + strconv.Itoa(CoreI.CheckVersion())); !os.IsNotExist(err) {
		log.Println("file exists")
		return nil
	}
	
	/*
	IndepDir, err := AssetDir("ArchIndep")
	log.Println(IndepDir)
	if err != nil {
		return err
	}
	for _, fn := range IndepDir {
		log.Println(datadir+"ArchIndep/"+fn)
		
		err := RestoreAsset(datadir, "ArchIndep/"+fn)
		log.Println(err)
		//GrantPremission
		os.Chmod(datadir+"ArchIndep/"+fn, 0700)
		log.Println(os.Symlink(datadir+"ArchIndep/"+fn, datadir+fn))			
	}*/
	
	DepDir, err := AssetDir("ArchDep")
	log.Println(DepDir)
	if err != nil {
		return err
	}
	for _, fn := range DepDir {
		DepDir2, err := AssetDir("ArchDep/" + fn)
		log.Println("ArchDep/" + fn)
		if err != nil {
			return err
		}
		for _, FND := range DepDir2 {
			log.Println(datadir+"ArchDep/"+fn+"/"+FND)
			
			RestoreAsset(datadir, "ArchDep/"+fn+"/"+FND)
			os.Chmod(datadir+"ArchDep/"+fn+"/"+FND, 0700)
			log.Println(os.Symlink(datadir+"ArchDep/"+fn+"/"+FND, datadir+FND))			
		}
	}
	s, _ := os.Create(datadir + strconv.Itoa(CoreI.CheckVersion()))
	s.Close()

	return nil
}

func (v *FirstRun) SetCoreI(status *CoreI.Status) {
	v.status = status
}

func (v *FirstRun) CheckAndExport() error {
	return v.checkIfRcExist()
}
