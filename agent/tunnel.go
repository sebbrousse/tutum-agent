package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/ActiveState/tail"
	"github.com/tutumcloud/tutum-agent/utils"
)

type TunnelPatchForm struct {
	Tunnel  string `json:"tunnel:"`
	Version string `json:"agent_version"`
}

func StartNatTunneling(url, ngrokPath, ngrokLogPath string) {
	if !utils.FileExist(ngrokPath) {
		Logger.Printf("Cannot find ngrok binary(%s), skip tunneling\n", ngrokPath)
		return
	}

	var commandStr string
	if *FlagNgrokToken != "" {
		Logger.Println("About to tunnel to public ngrok service")
		commandStr = fmt.Sprintf("%s -log stdout -authtoken %s -proto tcp 2375 > %s",
			ngrokPath, *FlagNgrokToken, ngrokLogPath)
	} else {
		Logger.Println("About to tunnel to private ngrok service")
		confPath := path.Join(TutumHome, NgrokConfName)
		if !utils.FileExist(confPath) {
			Logger.Println("Cannot find ngrok conf, skip tunneling")
			return
		}
		commandStr = fmt.Sprintf("%s -log stdout -proto tcp 2375 > %s",
			ngrokPath, ngrokLogPath)
	}

	os.RemoveAll(ngrokLogPath)
	Logger.Println("Starting tunneling:", commandStr)

	command := exec.Command("/bin/sh", "-c", commandStr)
	go runGronk(command)
	Logger.Println("Starting montoring tunnel:", commandStr)
	go montitorTunnels(url, ngrokLogPath)
}

func runGronk(command *exec.Cmd) bool {
	if err := command.Start(); err != nil {
		return true
	}
	command.Wait()
	return true
}

func montitorTunnels(url, ngrokLogPath string) {
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

	client := &http.Client{}
	req, err := http.NewRequest("PATCH", utils.JoinURL(url, Conf.TutumUUID), bytes.NewReader(data))
	if err != nil {
		Logger.Println(err)
	}
	req.Header.Add("Authorization", "TutumAgentToken "+Conf.TutumToken)
	req.Header.Add("Content-Type", "application/json")
	if *FlagDebugMode {
		Logger.Println("=======Request Info ======")
		Logger.Println("=> URL:", utils.JoinURL(url, Conf.TutumUUID))
		Logger.Println("=> Method:", "PATCH")
		Logger.Println("=> Headers:", req.Header)
		Logger.Println("=> Body:", string(data))
	}
	resp, err := client.Do(req)
	if err != nil {
		Logger.Println(err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200, 201, 202:
		Logger.Println(resp.Status)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Logger.Println(err)
		}
		if *FlagDebugMode {
			Logger.Println("=======Response Info ======")
			Logger.Println("=> Headers:", resp.Header)
			Logger.Println("=> Body:", string(body))
		}
	default:
		if *FlagDebugMode {
			Logger.Println("=======Response Info (ERROR) ======")
			Logger.Println("=> Headers:", resp.Header)
			b, _ := ioutil.ReadAll(resp.Body)
			Logger.Println("=> Body:", string(b))
		}
		Logger.Println(resp.Status)
	}
	Logger.Println("Patching tunnel address to Tutum is finished")
}
