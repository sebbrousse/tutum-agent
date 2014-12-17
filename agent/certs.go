package agent

import (
	"crypto/rand"
	"crypto/rsa"
	_ "crypto/sha1"
	_ "crypto/sha256"
	_ "crypto/sha512"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	"github.com/tutumcloud/tutum-agent/utils"
)

func CreateCerts(keyFilePath, certFilePath, host string) {
	if isCertificateExist(keyFilePath, certFilePath) {
		Logger.Println("TLS certificate exists, skipping")
	} else {
		if Conf.CertCommonName == "" {
			Logger.Fatalln("CertCommonName is empty. This may be caused by failure on POSTing to Tutum.")
		}
		Logger.Println("No tls certificate founds, creating a new one using CN:", Conf.CertCommonName)
		genCetificate(keyFilePath, certFilePath, host)
		Logger.Println("New tls certificate is generated")
	}
}

func isCertificateExist(keyFilePath, certFilePath string) (isExist bool) {
	if utils.FileExist(keyFilePath) && utils.FileExist(certFilePath) {
		return true
	}
	return false
}

func genCetificate(keyFilePath, certFilePath, host string) {
	validFor := 10 * 365 * 24 * time.Hour
	isCA := true
	rsaBits := 2048

	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		Logger.Fatalf("Failed to generate private key: %s", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		Logger.Fatalf("Failed to generate serial number: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Tutum Self-Signed Host"},
			CommonName:   host,
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		Logger.Fatalf("Failed to create certificate: %s", err)
	}

	certOut, err := os.Create(certFilePath)
	if err != nil {
		Logger.Fatalf("Failed to open cert.pem for writing: %s", err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()
	Logger.Printf("Written certificate to %s\n", certFilePath)

	keyOut, err := os.OpenFile(keyFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		Logger.Print("Failed to open key.pem for writing:", err)
		return
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()
	Logger.Printf("Written %s\n", keyFilePath)
}

func GetCertificate(certFilePath string) (*string, error) {
	content, err := ioutil.ReadFile(certFilePath)
	if err != nil {
		return nil, err
	}
	cert := string(content[:])
	return &cert, nil

}
