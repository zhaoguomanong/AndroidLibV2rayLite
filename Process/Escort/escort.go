package Escort

import (
	"os"
	"os/exec"

	"log"
)
import "github.com/2dust/AndroidLibV2rayLite/CoreI"

func (v *Escorting) EscortRun(proc string, pt []string, additionalEnv string) {
	log.Println(proc)
	log.Println(pt)
	count := 42
	for count > 0 {
		cmd := exec.Command(proc, pt...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if len(additionalEnv) > 0 {
			//additionalEnv := "FOO=bar"
			newEnv := append(os.Environ(), additionalEnv)
			cmd.Env = newEnv
		}

		if err := cmd.Start(); err != nil {
			log.Println("EscortRun cmd.Start err", err)
			goto CMDERROR
		}

		*v.escortProcess = append(*v.escortProcess, cmd.Process)
		log.Println("EscortRun Waiting....")
		if err := cmd.Wait(); err != nil {
			log.Println("EscortRun cmd.Wait err:", err)
		}

	CMDERROR:
		if v.status.IsRunning {
			log.Println("EscortRun Unexpected Exit")
			count--
		} else {
			log.Println("EscortRun Exit")
			break
		}
	}
}

func (v *Escorting) EscortingUPV() {
	if v.escortProcess != nil {
		return
	}
	v.escortProcess = new([](*os.Process))
}

func (v *Escorting) EscortingDown() {
	log.Println("escortingDown() Killing all escorted process ")
	if v.escortProcess == nil {
		return
	}
	for _, pr := range *v.escortProcess {
		pr.Kill()
		pr.Wait()
	}
	v.escortProcess = nil
}

func (v *Escorting) SetStatus(st *CoreI.Status) {
	v.status = st
}

func NewEscort() *Escorting {
	return &Escorting{}
}

type Escorting struct {
	escortProcess *[](*os.Process)
	status        *CoreI.Status
}
