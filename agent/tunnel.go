package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/ActiveState/tail"
	"github.com/tutumcloud/tutum-agent/utils"
)

type TunnelPatchForm struct {
	Tunnel  string `json:"tunnel:"`
	Version string `json:"agent_version"`
}

func runGronk(command *exec.Cmd) bool {
	if err := command.Start(); err != nil {
		return true
	}
	command.Wait()
	return true
}

func monitorTunnels(url, ngrokLogPath string) {
	update, _ := tail.TailFile(ngrokLogPath, tail.Config{
		Follow: true,
		ReOpen: true})
	for line := range update.Lines {
		if strings.Contains(line.Text, "[INFO] [client] Tunnel established at") {
			terms := strings.Split(line.Text, " ")
			tunnel := terms[len(terms)-1]
			Logger.Printf("Found new tunnel:%s\n", tunnel)
			patchTunnelToTutum(url, tunnel)
		}
	}
}

func patchTunnelToTutum(url, tunnel string) {
	Logger.Println("Patching tunnel address to Tutum")
	form := TunnelPatchForm{}
	form.Version = VERSION
	form.Tunnel = tunnel
	data, err := json.Marshal(form)
	if err != nil {
		Logger.Printf("Cannot marshal the TunnelPatch form:%s\f", err)
	}

	headers := []string{"Authorization TutumAgentToken " + Conf.TutumToken,
		"Content-Type", "application/json"}
	SendRequest("PATCH", utils.JoinURL(url, Conf.TutumUUID), data, headers)
	Logger.Println("Patching tunnel address to Tutum is finished")
}

func NatTunnel(url, ngrokPath, ngrokLogPath string) {
	counter := 0
	for {
		if counter > 10 {
			break
		}
		if DockerProcess == nil {
			time.Sleep(2 * time.Second)
			counter += 1

		} else {
			break
		}
	}

	Logger.Printf("Testing if port %d is publicly reachable ...\n", DockerHostPort)
	cmdStr := fmt.Sprintf("nc %s %d < /dev/null", Conf.CertCommonName, DockerHostPort)
	command := exec.Command("/bin/sh", "-c", cmdStr)
	command.Start()
	if err := command.Wait(); err != nil {
		Logger.Printf("Port %d is not publicly reachable, NAT tunne is needed", DockerHostPort)
	} else {
		Logger.Printf("Port %d is publicly reachable, NAT tunnel is not needed", DockerHostPort)
		return
	}

	if !utils.FileExist(ngrokPath) {
		Logger.Printf("Cannot find ngrok binary(%s), skipping NAT tunnel\n", ngrokPath)
		return
	}

	var commandStr string
	if *FlagNgrokToken != "" {
		Logger.Println("About to tunnel to public ngrok service")
		commandStr = fmt.Sprintf("%s -log stdout -authtoken %s -proto tcp %d > %s",
			ngrokPath, *FlagNgrokToken, DockerHostPort, ngrokLogPath)
	} else {
		Logger.Println("About to tunnel to private ngrok service")
		confPath := path.Join(TutumHome, NgrokConfName)
		if !utils.FileExist(confPath) {
			Logger.Println("Cannot find ngrok conf, skipping NAT tunnel")
			return
		}
		commandStr = fmt.Sprintf("%s -log stdout -proto tcp %d > %s",
			ngrokPath, DockerHostPort, ngrokLogPath)
	}

	os.RemoveAll(ngrokLogPath)
	Logger.Println("Starting montoring tunnel:", commandStr)
	go monitorTunnels(url, ngrokLogPath)
	Logger.Println("Starting NAT tunnel:", commandStr)

	for {
		command := exec.Command("/bin/sh", "-c", commandStr)
		runGronk(command)
		Logger.Println("Restarting NAT tunnel:", commandStr)
	}
}
