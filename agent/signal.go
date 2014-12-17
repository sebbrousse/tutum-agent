package agent

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

var DockerProcess *os.Process
var c chan os.Signal

func HandleSig() {
	c = make(chan os.Signal, 1)

	SetHandler()
	go func() {
		for {
			s := <-c
			Logger.Println("Got signal:", s)
			if s == syscall.SIGCHLD {
				Logger.Println("Docker deamon has died")
				DockerProcess = nil
			} else if s == os.Interrupt {
				Logger.Println("User interrupt")
				if DockerProcess == nil {
					Logger.Println("Docker daemon is not running")
					Logger.Fatalln("tutum-agent is terminated")
				} else {
					Logger.Println("Docker daemon is running")
					Logger.Println("Start to shut down docker daemon gracefully")
					DockerProcess.Signal(syscall.SIGTERM)
				}
				syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
			} else {
				go func() {
					for {
						if DockerProcess != nil {
							time.Sleep(10 * time.Millisecond)
						} else {
							Logger.Println("Tutum agent exited")
							os.Exit(130)
						}
					}
				}()
			}
		}
	}()
}

func SetHandler() {
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGCHLD)
}

func UnsetHandler() {
	signal.Stop(c)
}
