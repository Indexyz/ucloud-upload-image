package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/indexyz/ucloud-upload-image/pkgs/sshutil"
	"github.com/indexyz/ucloud-upload-image/pkgs/task"
	"github.com/indexyz/ucloud-upload-image/pkgs/ucloud"
	"github.com/sirupsen/logrus"
)

var zone = ""
var region = ""
var name = ""
var imageUrl = ""
var format = "raw"
var imageWriteOutFile = ""

var networkType = "Bgp"

func init() {
	flag.StringVar(&zone, "zone", "", "ucloud platform zone")
	flag.StringVar(&region, "region", "", "ucloud platform region")
	flag.StringVar(&name, "name", "", "image store name")
	flag.StringVar(&imageUrl, "image", "", "image local path or download url")
	flag.StringVar(&format, "format", "raw", "image format, available: raw, bz2, xz, zstd")
	flag.StringVar(&networkType, "network-type", "Bgp", "main interface network type, default Bgp")
	flag.StringVar(&imageWriteOutFile, "image-write-out-file", "", "write image id to this file when set")

	flag.Parse()
}

func main() {
	uc := ucloud.New(ucloud.Options{
		Region:      region,
		Zone:        zone,
		NetworkType: networkType,
		PublicKey:   os.Getenv("UCLOUD_PUBLIC_KEY"),
		PrivateKey:  os.Getenv("UCLOUD_PRIVATE_KEY"),
	})

	priv, pub, err := sshutil.GenerateKeyPair()
	if err != nil {
		panic(err)
	}

	keyPairID, err := uc.ImportKeyPair(
		fmt.Sprintf("tmp-keypair-%d", os.Getpid()),
		string(pub),
	)
	if err != nil {
		panic(err)
	}

	logrus.Infof("=== keypair created: %s", keyPairID)

	debianImageID, err := uc.FindImageIDByNameContains("Debian 12")
	if err != nil {
		panic(err)
	}

	logrus.Infof("=== fetch create uhost image: %s", debianImageID)

	hostIDs, err := uc.CreateTempInstance(
		fmt.Sprintf("tmp-uhost-%d", os.Getpid()),
		debianImageID,
		keyPairID,
	)
	if err != nil {
		panic(err)
	}

	logrus.Infof("=== host instance created: %s", hostIDs[0])

	publicIP, err := uc.WaitPublicIP(hostIDs)
	if err != nil {
		panic(err)
	}

	logrus.Infof("=== public ip: %s", publicIP)

	err = task.RunStepHost(&task.TaskContext{
		ConnectIP:   publicIP,
		PrivateKey:  priv,
		ImagePath:   imageUrl,
		ImageFormat: format,
	})

	createImage := true

	if err != nil {
		logrus.Errorf("failed to write disk image: %v", err)
		createImage = false
	}

	err = uc.DeleteKeyPair(keyPairID)
	if err != nil {
		logrus.Warnf("delete keypair error: %v", err)
	}

	if createImage {
		uc.WaitStopped(hostIDs)
		imageID, err := uc.CreateCustomImage(hostIDs[0], name)
		if err != nil {
			panic(err)
		}

		logrus.Infof("image id: %s", imageID)
		uc.WaitImageDone(imageID)

		if len(imageWriteOutFile) > 0 {
			err := os.WriteFile(imageWriteOutFile, []byte(imageID), 0o644)
			if err != nil {
				panic(err)
			}
		}
	}

	err = uc.TerminateInstance(hostIDs[0], true, true)
	if err != nil {
		panic(err)
	}

}
