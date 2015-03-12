package agent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ActiveState/tail"
	"github.com/tutumcloud/tutum-agent/utils"
)

type TunnelPatchForm struct {
	Tunnel  string `json:"tunnel"`
	Version string `json:"agent_version"`
}

func NatTunnel(url, ngrokPath, ngrokLogPath, ngrokConfPath string) {
	if !utils.FileExist(ngrokPath) {
		Logger.Printf("Cannot find ngrok binary(%s), skipping NAT tunnel", ngrokPath)
		return
	}

	if !isNodeNated() {
		return
	}

	updateNgrokHost(url)
	createNgrokConfFile(ngrokConfPath)

	var cmd *exec.Cmd
	if *FlagNgrokToken != "" {
		Logger.Println("About to tunnel to public ngrok service")
		cmd = exec.Command(ngrokPath,
			"-log", "stdout",
			"-authtoken", *FlagNgrokToken,
			"-proto", "tcp",
			DockerHostPort)
	} else {
		Logger.Println("About to tunnel to private ngrok service")
		if !utils.FileExist(ngrokConfPath) {
			Logger.Println("Cannot find ngrok conf, skipping NAT tunnel")
			return
		}
		cmd = exec.Command(ngrokPath,
			"-config", ngrokConfPath,
			"-log", "stdout",
			"-proto", "tcp",
			DockerHostPort)
	}

	os.RemoveAll(ngrokLogPath)
	logFile, err := os.OpenFile(ngrokLogPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		Logger.Println(err)
	} else {
		defer logFile.Close()
		cmd.Stdout = logFile
	}

	Logger.Println("Starting montoring tunnel:", cmd.Args)
	go monitorTunnels(url, ngrokLogPath)
	Logger.Println("Starting NAT tunnel:", cmd.Args)

	for {
		runGronk(cmd)
		time.Sleep(10 * time.Second)
		Logger.Println("Restarting NAT tunnel:", cmd.Args)
	}
}

func runGronk(cmd *exec.Cmd) bool {
	if err := cmd.Start(); err != nil {
		return true
	}
	cmd.Wait()
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
			Logger.Printf("Found new tunnel:%s", tunnel)
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
	_, err = SendRequest("PATCH", utils.JoinURL(url, Conf.TutumUUID), data, headers)
	if err != nil {
		Logger.Println("Failed to patch tunnel address to Tutum,", err)
	} else {
		Logger.Println("Successfully Patched tunnel address to Tutum")
	}
}

func DownloadNgrok(url, ngrokBinPath string) {
	if utils.FileExist(ngrokBinPath) {
		Logger.Printf("Found ngrok locally(%s), skip downloading", ngrokBinPath)
	} else {
		Logger.Println("No ngrok binary is found locally. Starting to download ngrok...")
		downloadFile(url, ngrokBinPath, "ngrok")
	}
}

func createNgrokConfFile(ngrokConfPath string) {
	ngrokConfStr := fmt.Sprintf("server_addr: %s\ntrust_host_root_certs: false", NgrokHost)
	Logger.Printf("Creating ngrok config file in %s ...", ngrokConfPath)
	if err := ioutil.WriteFile(ngrokConfPath, []byte(ngrokConfStr), 0666); err != nil {
		Logger.Println("Cannot create ngrok config file:", err)
	}
}

func updateNgrokHost(url string) {
	if NgrokHost != "" {
		return
	}

	headers := []string{"Authorization TutumAgentToken " + Conf.TutumToken,
		"Content-Type application/json"}
	body, err := SendRequest("GET", utils.JoinURL(url, Conf.TutumUUID), nil, headers)
	if err != nil {
		Logger.Printf("Get registration info error, %s", err)
	} else {
		var form RegGetForm
		if err = json.Unmarshal(body, &form); err != nil {
			Logger.Println("Cannot unmarshal the response", err)
		} else {
			if form.NgrokHost != "" {
				NgrokHost = form.NgrokHost
				Logger.Println("Set ngrok server address to", NgrokHost)
			}
		}
	}
}

func isNodeNated() bool {
	Logger.Printf("Testing if port %s is publicly reachable ...", DockerHostPort)
	Logger.Println("Waiting for the startup of docker ...")
	for {
		cmdstring := fmt.Sprintf("nc -w 10 127.0.0.1 %s < /dev/null", DockerHostPort)
		command := exec.Command("/bin/sh", "-c", cmdstring)
		command.Start()
		if err := command.Wait(); err == nil {
			Logger.Println("Docker daemon has started, testing if it's publicly reachable")
			break
		} else {
			Logger.Println("Docker daemon has not started yet. Retrying in 2 seconds")
			time.Sleep(2 * time.Second)
		}
	}

	cmdstring := fmt.Sprintf("nc -w 10 %s %s < /dev/null", Conf.CertCommonName, DockerHostPort)
	command := exec.Command("/bin/sh", "-c", cmdstring)
	command.Start()
	if err := command.Wait(); err != nil {
		Logger.Printf("Port %s is not publicly reachable", DockerHostPort)
		return true
	} else {
		Logger.Printf("Port %s is publicly reachable, skipping NAT tunnel", DockerHostPort)
		return false
	}
}
