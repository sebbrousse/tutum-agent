package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/tutumcloud/tutum-agent/utils"
)

type ResponseForm struct {
	UserCaCert      string `json:"user_ca_cert"`
	TutumUUID       string `json:"uuid"`
	CertCommonName  string `json:"external_fqdn"`
	DockerBinaryURL string `json:"docker_url"`
}

type PostForm struct {
	Version string `json:"agent_version"`
}

type PatchForm struct {
	Public_cert string `json:"public_cert"`
	Version     string `json:"agent_version"`
}

func PostToTutum(url, caFilePath, configFilePath string) error {
	form := PostForm{}
	form.Version = VERSION
	data, err := json.Marshal(form)
	if err != nil {
		Logger.Fatalln("Cannot marshal the POST form", err)
	}
	return Register(url, "POST", Conf.TutumToken, Conf.TutumUUID, caFilePath, configFilePath, data)
}

func PatchToTutum(url, caFilePath, certFilePath, configFilePath string) error {
	form := PatchForm{}
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
			if err = handleResponse(body, caFilePath, configFilePath); err == nil {
				return nil
			}
		}
		if err.Error() == "Error 404" {
			return err
		}
		Logger.Printf("Registration failed: %s. Retry in %d seconds\n", err.Error(), i)
		time.Sleep(time.Duration(i) * time.Second)
	}
}

func sendRegRequest(url, method, token, uuid string, data []byte) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, utils.JoinURL(url, uuid), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "TutumAgentToken "+token)
	req.Header.Add("Content-Type", "application/json")
	if *FlagDebugMode {
		Logger.Println("=======Request Info ======")
		Logger.Println("=> URL:", utils.JoinURL(url, uuid))
		Logger.Println("=> Method:", method)
		Logger.Println("=> Headers:", req.Header)
		Logger.Println("=> Body:", string(data))
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200, 201, 202:
		Logger.Println(resp.Status)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if *FlagDebugMode {
			Logger.Println("=======Response Info ======")
			Logger.Println("=> Headers:", resp.Header)
			Logger.Println("=> Body:", string(body))
		}
		return body, nil
	case 404:
		return nil, errors.New("Error 404")
	default:
		if *FlagDebugMode {
			Logger.Println("=======Response Info (ERROR) ======")
			Logger.Println("=> Headers:", resp.Header)
			b, _ := ioutil.ReadAll(resp.Body)
			Logger.Println("=> Body:", string(b))
		}
		return nil, errors.New(resp.Status)
	}
}

func handleResponse(body []byte, caFilePath, configFilePath string) error {
	var responseForm ResponseForm

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

	// Save to configuration file
	if isModified {
		Logger.Println("Updating configraution file ...")
		return SaveConf(configFilePath, Conf)
	}
	return nil
}
