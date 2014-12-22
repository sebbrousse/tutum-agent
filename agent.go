package main

import (
	"os"
	"path"
	"runtime"
	"time"

	. "github.com/tutumcloud/tutum-agent/agent"
	"github.com/tutumcloud/tutum-agent/utils"
)

func init() {
	runtime.GOMAXPROCS(4)
}

func main() {
	dockerBinPath := path.Join(DockerDir, DockerBinaryName)
	dockerNewBinPath := path.Join(DockerDir, DockerNewBinaryName)
	dockerNewBinSigPath := path.Join(DockerDir, DockerNewBinarySigName)
	configFilePath := path.Join(TutumHome, ConfigFileName)
	keyFilePath := path.Join(TutumHome, KeyFileName)
	certFilePath := path.Join(TutumHome, CertFileName)
	caFilePath := path.Join(TutumHome, CAFileName)

	ParseFlag()
	SetLogger(path.Join(LogDir, TutumLogFileName))

	Logger.Println("Preparing directories and files...")
	PrepareFiles(configFilePath, dockerBinPath, keyFilePath, certFilePath)

	SetConfigFile(configFilePath)

	url := utils.JoinURL(Conf.TutumHost, RegEndpoint)
	if Conf.TutumUUID == "" {
		Logger.Printf("Registering in Tutum via POST: %s ...\n", url)
		PostToTutum(url, caFilePath, configFilePath)
	}

	Logger.Println("Checking if TLS certificate exists...")
	CreateCerts(keyFilePath, certFilePath, Conf.CertCommonName)

	Logger.Printf("Registering in Tutum via PATCH: %s ...\n", url+Conf.TutumUUID)
	PatchToTutum(url, caFilePath, certFilePath, configFilePath)

	Logger.Println("Check if docker binary exists...")
	DownloadDocker(Conf.DockerBinaryURL, dockerBinPath)

	Logger.Println("Setting system signals...")
	HandleSig()

	Logger.Println("Starting docker daemon...")
	StartDocker(dockerBinPath, keyFilePath, certFilePath, caFilePath)

	Logger.Println("Docker server started. Entering maintenance loop")
	for {
		time.Sleep(HeartBeatInterval * time.Second)
		UpdateDocker(dockerBinPath, dockerNewBinPath, dockerNewBinSigPath, keyFilePath, certFilePath, caFilePath)

		// try to restart docker daemon if it dies somehow
		if DockerProcess == nil {
			time.Sleep(HeartBeatInterval * time.Second)
			if DockerProcess == nil && ScheduleToTerminateDocker == false {
				Logger.Println("Respawning docker daemon")
				StartDocker(dockerBinPath, keyFilePath, certFilePath, caFilePath)
			}
		}
	}
}

func PrepareFiles(configFilePath, dockerBinPath, keyFilePath, certFilePath string) {
	Logger.Println("Creating all necessary folders...")
	_ = os.MkdirAll(TutumHome, 0755)
	_ = os.MkdirAll(DockerDir, 0755)
	_ = os.MkdirAll(LogDir, 0755)

	Logger.Println("Checking if config file exists...")
	if utils.FileExist(configFilePath) {
		Logger.Println("Config file exist, skipping")
	} else {
		Logger.Println("Creating a new config file")
		LoadDefaultConf()
		if err := SaveConf(configFilePath, Conf); err != nil {
			Logger.Fatalln(err)
		}
	}

	Logger.Println("Loading Configuration file...")
	conf, err := LoadConf(configFilePath)
	if err != nil {
		Logger.Fatalln("Failed to load configuration file:", err)
	} else {
		Conf = *conf
	}
}
