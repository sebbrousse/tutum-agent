package agent

import (
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"
)

func HandleSig() {
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for {
			s := <-c
			Logger.Println("Got signal:", s)
			if s == os.Interrupt {
				Logger.Println("User interrupt")
				if DockerProcess == nil {
					Logger.Println("Docker daemon is not running")
					os.RemoveAll(TutumPidFile)
					Logger.Fatal("Exiting agent")
				} else {
					Logger.Println("Docker daemon is running")
					Logger.Println("Starting to shut down docker daemon gracefully")
					ScheduleToTerminateDocker = true
					DockerProcess.Signal(syscall.SIGTERM)
				}
				syscall.Kill(os.Getpid(), syscall.SIGTERM)
			} else if s == syscall.SIGHUP {
				go ReloadLogger(path.Join(LogDir, TutumLogFileName), path.Join(LogDir, DockerLogFileName))
			} else {
				go func() {
					for {
						if DockerProcess != nil {
							time.Sleep(10 * time.Millisecond)
						} else {
							Logger.Println("Exiting agent")
							os.RemoveAll(TutumPidFile)
							os.Exit(130)
						}
					}
				}()
			}
		}
	}()
}
