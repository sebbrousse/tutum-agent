package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/tutumcloud/tutum-agent/utils"
)

type RegResponseForm struct {
	UserCaCert      string `json:"user_ca_cert"`
	TutumUUID       string `json:"uuid"`
	CertCommonName  string `json:"external_fqdn"`
	DockerBinaryURL string `json:"docker_url"`
	NgrokBinaryURL  string `json:"ngrok_url"`
}

type RegPostForm struct {
	Version string `json:"agent_version"`
}

type RegPatchForm struct {
	Public_cert string `json:"public_cert"`
	Version     string `json:"agent_version"`
}

func PostToTutum(url, caFilePath, configFilePath string) error {
	form := RegPostForm{}
	form.Version = VERSION
	data, err := json.Marshal(form)
	if err != nil {
		Logger.Fatalln("Cannot marshal the POST form", err)
	}
	return Register(url, "POST", Conf.TutumToken, Conf.TutumUUID, caFilePath, configFilePath, data)
}

func PatchToTutum(url, caFilePath, certFilePath, configFilePath string) error {
	form := RegPatchForm{}
	form.Version = VERSION
	cert, err := GetCertificate(certFilePath)
	if err != nil {
		Logger.Fatal("Cannot read public certificate:", err)
		form.Public_cert = ""
	}
	form.Public_cert = *cert
	data, err := json.Marshal(form)
	if err != nil {
		Logger.Fatalln("Cannot marshal the PATCH form", err)
	}

	return Register(url, "PATCH", Conf.TutumToken, Conf.TutumUUID, caFilePath, configFilePath, data)
}

func Register(url, method, token, uuid, caFilePath, configFilePath string, data []byte) error {
	if token == "" {
		fmt.Fprintf(os.Stderr, "Tutum token is empty. Please run 'tutum-agent set TutumToken=xxx' first!")
		Logger.Fatalln("Tutum token is empty. Please run 'tutum-agent set TutumToken=xxx' first!")
	}

	for i := 1; ; i *= 2 {
		if i > MaxWaitingTime {
			i = 1
		}
		body, err := sendRegRequest(url, method, token, uuid, data)
		if err == nil {
			if err = handleRegResponse(body, caFilePath, configFilePath); err == nil {
				return nil
			}
		}
		if err.Error() == "Status: 404" {
			return err
		}
		Logger.Printf("Registration failed, %s. Retry in %d seconds\n", err.Error(), i)
		time.Sleep(time.Duration(i) * time.Second)
	}
}

func sendRegRequest(url, method, token, uuid string, data []byte) ([]byte, error) {
	headers := []string{"Authorization TutumAgentToken " + token,
		"Content-Type application/json"}
	return SendRequest(method, utils.JoinURL(url, uuid), data, headers)

}

func handleRegResponse(body []byte, caFilePath, configFilePath string) error {
	var responseForm RegResponseForm

	// Save ca cert file
	if err := json.Unmarshal(body, &responseForm); err != nil {
		return errors.New("Cannot unmarshal json from response")
	}
	if err := ioutil.WriteFile(caFilePath, []byte(responseForm.UserCaCert), 0644); err != nil {
		Logger.Print("Failed to save "+caFilePath, err)
	}
	// Update global Conf
	isModified := false
	if Conf.CertCommonName != responseForm.CertCommonName {
		Logger.Printf("Cert CommonName has been changed from %s to %s\n", Conf.CertCommonName, responseForm.CertCommonName)
		isModified = true
		Conf.CertCommonName = responseForm.CertCommonName
	}
	if Conf.TutumUUID != responseForm.TutumUUID {
		Logger.Printf("Tutum UUID has been changed from %s to %s\n", Conf.TutumUUID, responseForm.TutumUUID)
		isModified = true
		Conf.TutumUUID = responseForm.TutumUUID
	}

	DockerBinaryURL = responseForm.DockerBinaryURL

	if responseForm.NgrokBinaryURL != "" {
		NgrokBianryURL = responseForm.NgrokBinaryURL
	}
	// Save to configuration file
	if isModified {
		Logger.Println("Updating configraution file ...")
		return SaveConf(configFilePath, Conf)
	}
	return nil
}
