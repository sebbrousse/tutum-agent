package agent

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

type Configuration struct {
	CertCommonName string
	DockerHost     string
	TutumHost      string
	TutumToken     string
	TutumUUID      string
}

func ParseFlag() {
	FlagDebugMode = flag.Bool("debug", false, "Enable debug mode")
	FlagLogToStdout = flag.Bool("stdout", false, "Print log to stdout")
	FlagStandalone = flag.Bool("standalone", false, "Standalone mode, skipping reg with tutum")
	FlagDockerHost = flag.String("docker-host", "", "Override 'DockerHost'")
	FlagDockerOpts = flag.String("docker-opts", "", "Add additional flags to run docker daemon")
	FlagTutumHost = flag.String("tutum-host", "", "Override 'TutumHost'")
	FlagTutumToken = flag.String("tutum-token", "", "Override 'TutumToken'")
	FlagTutumUUID = flag.String("tutum-uuid", "", "Override 'TutumUUID'")
	FlagNgrokToken = flag.String("ngrok-token", "", "ngrok token for NAT tunneling")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprint(os.Stderr, "   set: Set items in the config file and exit, supported items\n",
			"          CertCommonName=\"xxx\"\n",
			"          DockerHost=\"xxx\"\n",
			"          TutumHost=\"xxx\"\n",
			"          TutumToken=\"xxx\"\n",
			"          TutumUUID=\"xxx\"\n")
	}
	flag.Parse()
}

func SetConfigFile(configFilePath string) {
	// Set tutum config file content and exit, when "tutum-agent set" is called
	numberOfNonFlagArg := flag.NArg()
	if numberOfNonFlagArg == 0 {
		return
	} else if numberOfNonFlagArg == 1 {
		flag.Usage()
		os.Exit(1)
	} else {
		for i, param := range flag.Args() {
			if i == 0 {
				if param != "set" {
					flag.Usage()
					os.Exit(1)
				}
			} else {
				keyValue := strings.SplitN(param, "=", 2)
				if len(keyValue) != 2 {
					flag.Usage()
					os.Exit(1)
				}
				key := strings.TrimSpace(keyValue[0])
				value := strings.Trim(strings.TrimSpace(keyValue[1]), "\"'")
				if strings.ToLower(key) == strings.ToLower("CertCommonName") {
					Conf.CertCommonName = value
				} else if strings.ToLower(key) == strings.ToLower("DockerHost") {
					Conf.DockerHost = value
				} else if strings.ToLower(key) == strings.ToLower("TutumHost") {
					Conf.TutumHost = value
				} else if strings.ToLower(key) == strings.ToLower("TutumToken") {
					Conf.TutumToken = value
				} else if strings.ToLower(key) == strings.ToLower("TutumUUID") {
					Conf.TutumUUID = value
				} else {
					fmt.Fprintf(os.Stderr, "Unsupported item \"%s\" in \"tutum-agent set\" command", key)
					os.Exit(1)
				}
			}
		}
	}
	if err := SaveConf(configFilePath, Conf); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	Logger.Println("Tutum Agent configuration has been successfully updated!")
	os.Exit(0)
}

func LoadConf(configFile string) (*Configuration, error) {
	var conf Configuration
	f, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	//read and decode json format config file
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&conf)
	if err != nil {
		return nil, err
	}
	if conf.DockerHost == "" {
		conf.DockerHost = defaultDockerHost
	}

	if conf.TutumHost == "" {
		conf.TutumHost = defaultTutumHost
	}
	return &conf, nil
}

func SaveConf(configFile string, conf Configuration) error {
	f, err := os.OpenFile(configFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return errors.New("Failed to open config file for writing:" + err.Error())
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	err = encoder.Encode(conf)
	if err != nil {
		return errors.New("Failed to write the config file:" + err.Error())
	}
	return nil
}

func LoadDefaultConf() {
	if Conf.CertCommonName == "" {
		Conf.CertCommonName = defaultCertCommonName
	}
	if Conf.DockerHost == "" {
		Conf.DockerHost = defaultDockerHost
	}
	if Conf.TutumHost == "" {
		Conf.TutumHost = defaultTutumHost
	}
}

func SetLogger(logFile string) {
	if *FlagLogToStdout {
		Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
		Logger.Println("Set logger to stdout")

	} else {
		f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Println(err)
			log.Println("Log to stdout instead")
			f = os.Stdout
		}
		Logger = log.New(f, "", log.Ldate|log.Ltime)
		if f != os.Stdout {
			Logger.Println("Set logger to", logFile)
		}
	}
}
