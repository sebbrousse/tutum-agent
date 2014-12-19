package agent

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/tutumcloud/tutum-agent/utils"
)

type DockerMgmtDef struct {
	Version             string `json:"version"`
	Download_url        string `json: "download_url"`
	Checksum_md5_url    string `json: "checksum_md5_url"`
	Checksum_sha256_url string `json: "checksum_sha256_url"`
}

func StartDocker(dockerBinPath, keyFilePath, certFilePath, caFilePath string) {
	var command *exec.Cmd

	command = exec.Command(dockerBinPath, "-d",
		"-H", Conf.DockerHost,
		"-H", DockerDefaultHost,
		"--tlscert", certFilePath,
		"--tlskey", keyFilePath,
		"--tlscacert", caFilePath,
		"--tlsverify")

	go func(cmd *exec.Cmd) {
		//open file to log docker logs
		dockerLog := path.Join(LogDir, DockerLogFileName)
		Logger.Println("Set docker log to", dockerLog)
		f, err := os.OpenFile(dockerLog, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			Logger.Println(err)
			Logger.Println("Cannot set docker log to", dockerLog)
		} else {
			defer f.Close()
			command.Stdout = f
			command.Stderr = f
		}

		Logger.Println("Starting docker daemon:", command.Args)
		if err := cmd.Start(); err != nil {
			Logger.Println("Cannot start docker daemon:", err)
		}
		DockerProcess = cmd.Process
		if err := cmd.Wait(); err != nil {
			Logger.Println("Docker daemon is terminated:", err)
			DockerProcess = nil
		}
	}(command)
}

func StopDocker() {
	if DockerProcess != nil {
		DockerProcess.Signal(syscall.SIGTERM)
		for {
			if DockerProcess != nil {
				time.Sleep(500 * time.Millisecond)
			} else {
				break
			}
		}
	}
}
func DownloadDocker(url, dockerBinPath string) {
	if utils.FileExist(dockerBinPath) {
		Logger.Printf("Found docker locally(%s), skip downloading\n", dockerBinPath)
	} else {
		Logger.Println("No docker binary found locally. Starting to download docker...")

		Logger.Println("Downloading docker definition from", url)
		def := downloadDockerDef(url)
		Logger.Println("Successfully downloaded docker definition")

		Logger.Println("Downloading docker binary from", def.Download_url)
		binary := downloadDockerBin(def)
		Logger.Println("Successfully downloaded docker binary")

		Logger.Println("Writing docker binary to", dockerBinPath)
		writeDockerFile(binary, dockerBinPath)
		createDockerSymlink(dockerBinPath, DockerSymbolicLink)
	}
}

func UpdateDocker(dockerBinPath, dockerNewBinPath, dockerNewBinSigPath, keyFilePath, certFilePath, caFilePath string) {
	if utils.FileExist(dockerNewBinPath) {
		Logger.Printf("New Docker binary(%s) found\n", dockerNewBinPath)
		if verifyDockerSig(dockerNewBinPath, dockerNewBinSigPath) {
			Logger.Println("Stopping docker daemon")
			StopDocker()
			Logger.Println("Removing old docker binary")
			if err := os.RemoveAll(dockerBinPath); err != nil {
				Logger.Println("Cannot remove old docker binary:", err)
			}
			Logger.Println("Renaming new docker binary")
			if err := os.Rename(dockerNewBinPath, dockerBinPath); err != nil {
				Logger.Println("Cannot rename docker binary:", err)
			}
			Logger.Println("Removing the signature file ", dockerNewBinSigPath)
			if err := os.RemoveAll(dockerNewBinSigPath); err != nil {
				Logger.Println(err.Error())
			}
			createDockerSymlink(dockerBinPath, DockerSymbolicLink)
			StartDocker(dockerBinPath, keyFilePath, certFilePath, caFilePath)
		} else {
			Logger.Println("New docker binary signature cannot be verified. Update is rejected!")
			Logger.Println("Removing the invalid docker binary ", dockerNewBinPath)
			if err := os.RemoveAll(dockerNewBinPath); err != nil {
				Logger.Println(err.Error())
			}
			Logger.Println("Removing the invalid signature file ", dockerNewBinSigPath)
			if err := os.RemoveAll(dockerNewBinSigPath); err != nil {
				Logger.Println(err.Error())
			}
		}
	}
}

func downloadDockerDef(url string) *DockerMgmtDef {
	def, err := getDockerDef(url)
	for i := 1; ; i *= 2 {
		if i > MaxWaitingTime {
			i = 1
		}
		if err != nil || def == nil {
			Logger.Printf("Cannot get docker definition: %s. Retry in %d second\n", err.Error(), i)
			time.Sleep(time.Duration(i) * time.Second)
			def, err = getDockerDef(url)

		} else {
			break
		}
	}
	return def
}

func getDockerDef(url string) (*DockerMgmtDef, error) {
	var def DockerMgmtDef
	body, err := getBodyFromURL(url)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(body, &def); err != nil {
		return nil, err
	}
	if def == (DockerMgmtDef{}) {
		return nil, errors.New("Wrong docker defniniton")
	}
	return &def, nil
}

func downloadDockerBin(def *DockerMgmtDef) []byte {
	b, err := getDockerBin(def)
	for i := 1; ; i *= 2 {
		if i > MaxWaitingTime {
			i = 1
		}
		if err != nil {
			Logger.Printf("Cannot get docker binary: %s. Retry in %d second\n", err.Error(), i)
			time.Sleep(time.Duration(i) * time.Second)
			b, err = getDockerBin(def)

		} else {
			break
		}
	}
	return b
}

func getDockerBin(def *DockerMgmtDef) ([]byte, error) {
	b, err := getBodyFromURL(def.Download_url)
	if err != nil {
		return nil, err
	}

	//validate md5 checksum of the docker binary
	md5hasher := md5.New()
	md5hasher.Write(b)
	md5s := hex.EncodeToString(md5hasher.Sum(nil))
	Logger.Println("Checksum of the downloaded docker binary, md5:", md5s)
	md5b, err := getBodyFromURL(def.Checksum_md5_url)
	if err != nil {
		Logger.Println("Failed to get md5 for the docker binary")
		return nil, err
	} else {
		if !strings.Contains(string(md5b), md5s) {
			return nil, errors.New("Failed to pass md5 checksum test")
		}
	}
	Logger.Println("Docker binary passed md5 checksum check")

	//validate sha256 checksum of the docker binary
	sha256hasher := sha256.New()
	sha256hasher.Write(b)
	sha256s := hex.EncodeToString(sha256hasher.Sum(nil))
	Logger.Println("Checksum of the downloaded docker binary, shar256:", sha256s)
	sha256b, err := getBodyFromURL(def.Checksum_sha256_url)
	if err != nil {
		Logger.Println("Failed to get sha256 for the docker binary")
		return nil, err
	} else {
		if !strings.Contains(string(sha256b), sha256s) {
			return nil, errors.New("Failed to pass sha256 checksum test")
		}
	}
	Logger.Println("Docker binary passed sha256 checksum check")

	return b, nil
}

func writeDockerFile(binary []byte, dockerBinPath string) {
	err := ioutil.WriteFile(dockerBinPath, binary, 0755)
	for i := 1; ; i *= 2 {
		if i > MaxWaitingTime {
			i = 1
		}
		if err != nil {
			Logger.Printf("Failed to save docker binary: %s. Retry in %d second\n", err.Error(), i)
			time.Sleep(time.Duration(i) * time.Second)
			err = ioutil.WriteFile(dockerBinPath, binary, 0755)
		} else {
			break
		}
	}
}

func getBodyFromURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, errors.New(resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func verifyDockerSig(dockerNewBinPath, dockerNewBinSigPath string) bool {
	cmd := exec.Command("gpg", "--verify", dockerNewBinSigPath, dockerNewBinPath)
	UnsetHandler()
	err := cmd.Run()
	SetHandler()
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

func createDockerSymlink(dockerBinPath, dockerSymbolicLink string) {
	Logger.Println("Removing the docker symbolic from ", dockerSymbolicLink)
	if err := os.RemoveAll(DockerSymbolicLink); err != nil {
		Logger.Println(err.Error())
	}
	Logger.Println("Creating the docker symbolic to ", dockerSymbolicLink)
	if err := os.Symlink(dockerBinPath, DockerSymbolicLink); err != nil {
		Logger.Println(err.Error())
	}
}
