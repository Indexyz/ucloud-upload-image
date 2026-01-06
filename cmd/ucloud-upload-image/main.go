package main

import (
	"flag"
	"os"
	"path"
	"time"

	"github.com/5aaee9/ucloud-upload-image/pkgs/steps/kexec"
	"github.com/5aaee9/ucloud-upload-image/pkgs/steps/power"
	"github.com/5aaee9/ucloud-upload-image/pkgs/steps/writedisk"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

var zone = ""
var region = ""
var name = ""
var imageUrl = ""
var format = "raw"

func init() {
	flag.StringVar(&zone, "zone", "", "ucloud platform zone")
	flag.StringVar(&region, "region", "", "ucloud platform region")
	flag.StringVar(&name, "name", "", "image store name")
	flag.StringVar(&imageUrl, "image", "", "image local path or download url")
	flag.StringVar(&format, "format", "raw", "image format, available: raw, bz2, xz, zstd")

	flag.Parse()
}

func main() {
	// config := ucloud.NewConfig()
	// credential := auth.NewCredential()
	// credential.PrivateKey = os.Getenv("UCLOUD_PUBLIC_KEY")
	// credential.PublicKey = os.Getenv("UCLOUD_PRIVATE_KEY")

	// pub, priv, err := sshutil.GenerateKeyPair()
	// if err != nil {
	// 	panic(err)
	// }

	priv, err := os.ReadFile(path.Join(os.Getenv("HOME"), ".ssh/id_ed25519"))
	if err != nil {
		panic(err)
	}

	signer, err := ssh.ParsePrivateKey([]byte(priv))
	if err != nil {
		panic(err)
	}

	sshClientConfig := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		// There is no way to get the host key of the rescue system beforehand
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         1 * time.Minute,
	}

	host := "106.75.223.51"

	client, err := ssh.Dial("tcp", host+":22", sshClientConfig)
	if err != nil {
		panic(err)
	}
	logrus.Infof("start run vm into kexec env")
	err = kexec.RunInstanceIntoKexec(client, true)
	if err != nil {
		panic(err)
	}
	logrus.Infof("client running kexec env")

	_ = client.Close()

	// Wait kexec start new kernel (script wait 7 second after success run)
	time.Sleep(time.Second * 10)

	// Wait reconnect
	for {
		client, err = ssh.Dial("tcp", host+":22", sshClientConfig)
		if err == nil {
			break
		}

		logrus.Warnf("waiting connection: %v", err)
		time.Sleep(time.Second * 5)
	}

	err = writedisk.WriteDiskImage(client, imageUrl, format)
	if err != nil {
		panic(err)
	}

	logrus.Infof("==> power off machine")
	err = power.PowerOff(client)
	if err != nil {
		panic(err)
	}

	// hostClient := uhost.NewClient(&config, &credential)
	// createKeyPairRequest := hostClient.NewImportUHostKeyPairsRequest()
	// createKeyPairRequest.KeyPairName = ucloud.String(fmt.Sprintf("tmp-keypair-%d", os.Getpid()))
	// createKeyPairRequest.PublicKeyBody = ucloud.String(string(pub))
	// createKeyPairResponse, err := hostClient.ImportUHostKeyPairs(createKeyPairRequest)
	// if err != nil {
	// 	panic(err)
	// }

	// createHost := hostClient.NewCreateUHostInstanceRequest()
	// createHost.Name = ucloud.String(fmt.Sprintf("tmp-uhost-%d", os.Getpid()))
	// createHost.Zone = ucloud.String(zone)
	// createHost.CPU = ucloud.Int(2)
	// createHost.Memory = ucloud.Int(4096)
	// createHost.LoginMode = ucloud.String("KeyPair")
	// createHost.KeyPairId = ucloud.String(createKeyPairResponse.KeyPairId)
	// createHost.Disks = []uhost.UHostDisk{
	// 	{
	// 		IsBoot: ucloud.String("True"),
	// 		Type:   ucloud.String("CLOUD_RSSD"),
	// 		Size:   ucloud.Int(20),
	// 	},
	// }

	// createHost.MachineType = ucloud.String("O")
	// createHost.HotplugFeature = ucloud.Bool(false)
	// createHost.Features = &uhost.CreateUHostInstanceParamFeatures{
	// 	UNI: ucloud.Bool(false),
	// }
	// createHost.ChargeType = ucloud.String("Dynamic")

	// hostInstance, err := hostClient.CreateUHostInstance(createHost)
	// if err != nil {
	// 	panic(err)
	// }

	// for _, ip := range hostInstance.IPs {
	// 	println(string(ip))
	// }

	// deleteKeyPair := hostClient.NewDeleteUHostKeyPairsRequest()
	// deleteKeyPair.KeyPairIds = []string{createKeyPairResponse.KeyPairId}
	// _, err = hostClient.DeleteUHostKeyPairs(deleteKeyPair)
	// 	panic(err)
	// }
	// hostClient.DeleteUHostKeyPairs()
}
