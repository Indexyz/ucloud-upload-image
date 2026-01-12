package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/indexyz/ucloud-upload-image/pkgs/sshutil"
	"github.com/indexyz/ucloud-upload-image/pkgs/task"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

var zone = ""
var region = ""
var name = ""
var imageUrl = ""
var format = "raw"

var networkType = "Bgp"

func init() {
	flag.StringVar(&zone, "zone", "", "ucloud platform zone")
	flag.StringVar(&region, "region", "", "ucloud platform region")
	flag.StringVar(&name, "name", "", "image store name")
	flag.StringVar(&imageUrl, "image", "", "image local path or download url")
	flag.StringVar(&format, "format", "raw", "image format, available: raw, bz2, xz, zstd")
	flag.StringVar(&networkType, "network-type", "Bgp", "main interface network type, default Bgp")

	flag.Parse()
}

func main() {
	config := ucloud.NewConfig()
	config.Region = region
	credential := auth.NewCredential()
	credential.PublicKey = os.Getenv("UCLOUD_PUBLIC_KEY")
	credential.PrivateKey = os.Getenv("UCLOUD_PRIVATE_KEY")

	priv, pub, err := sshutil.GenerateKeyPair()
	if err != nil {
		panic(err)
	}

	hostClient := uhost.NewClient(&config, &credential)
	createKeyPairRequest := hostClient.NewImportUHostKeyPairsRequest()
	createKeyPairRequest.KeyPairName = ucloud.String(fmt.Sprintf("tmp-keypair-%d", os.Getpid()))
	createKeyPairRequest.PublicKeyBody = ucloud.String(string(pub))
	createKeyPairResponse, err := hostClient.ImportUHostKeyPairs(createKeyPairRequest)
	if err != nil {
		panic(err)
	}

	logrus.Infof("=== keypair created: %s", createKeyPairResponse.KeyPairId)

	searchImageRequest := hostClient.NewDescribeImageRequest()
	searchImageRequest.Limit = ucloud.Int(9999)
	searchImageRequest.Region = ucloud.String(region)
	imageResp, err := hostClient.DescribeImage(searchImageRequest)
	if err != nil {
		panic(err)
	}

	debianImage := lo.Filter(imageResp.ImageSet, func(item uhost.UHostImageSet, _ int) bool {
		if item.Zone != zone {
			return false
		}

		return strings.Contains(item.ImageName, "Debian 12")
	})

	if len(debianImage) == 0 {
		panic("image not found")
	}

	logrus.Infof("=== fetch create uhost image: %s", debianImage[0].ImageId)

	createHost := hostClient.NewCreateUHostInstanceRequest()
	createHost.Name = ucloud.String(fmt.Sprintf("tmp-uhost-%d", os.Getpid()))
	createHost.Zone = ucloud.String(zone)
	createHost.CPU = ucloud.Int(2)
	createHost.Memory = ucloud.Int(4096)
	createHost.ImageId = ucloud.String(debianImage[0].ImageId)
	createHost.LoginMode = ucloud.String("KeyPair")
	createHost.KeyPairId = ucloud.String(createKeyPairResponse.KeyPairId)
	createHost.Disks = []uhost.UHostDisk{
		{
			IsBoot: ucloud.String("True"),
			Type:   ucloud.String("CLOUD_RSSD"),
			Size:   ucloud.Int(20),
		},
	}

	createHost.NetworkInterface = []uhost.CreateUHostInstanceParamNetworkInterface{
		{
			EIP: &uhost.CreateUHostInstanceParamNetworkInterfaceEIP{
				Bandwidth:    ucloud.Int(1),
				PayMode:      ucloud.String("Bandwidth"),
				OperatorName: ucloud.String(networkType),
			},
		},
	}

	createHost.MachineType = ucloud.String("O")
	createHost.HotplugFeature = ucloud.Bool(false)
	createHost.Features = &uhost.CreateUHostInstanceParamFeatures{
		UNI: ucloud.Bool(false),
	}
	createHost.ChargeType = ucloud.String("Dynamic")
	createHost.NetCapability = ucloud.String("Normal")

	hostInstance, err := hostClient.CreateUHostInstance(createHost)
	if err != nil {
		panic(err)
	}

	logrus.Infof("=== host instance created: %s", hostInstance.UHostIds[0])

	queryInstance := func() (*uhost.DescribeUHostInstanceResponse, error) {
		request := hostClient.NewDescribeUHostInstanceRequest()
		request.Limit = ucloud.Int(1)
		request.Region = ucloud.String(region)
		request.Zone = ucloud.String(zone)
		request.UHostIds = hostInstance.UHostIds

		return hostClient.DescribeUHostInstance(request)
	}

	publicIp := ""
	for {
		hostInfo, err := queryInstance()
		if err != nil {
			panic(err)
		}

		publicIps := lo.Filter(hostInfo.UHostSet[0].IPSet, func(item uhost.UHostIPSet, index int) bool {
			return strings.EqualFold(item.Type, networkType)
			// return strings.ToLower(item.Type) == strings.ToLower(NetworkType)
		})

		if len(publicIps) == 0 {
			logrus.Warnf("failed to fetch public ip, wait for next round")
			time.Sleep(time.Second * 5)
			continue
		}

		publicIp = publicIps[0].IP
		break
	}

	logrus.Infof("=== public ip: %s", publicIp)

	err = task.RunStepHost(&task.TaskContext{
		ConnectIP:   publicIp,
		PrivateKey:  priv,
		ImagePath:   imageUrl,
		ImageFormat: format,
	})

	createImage := true

	if err != nil {
		logrus.Errorf("failed to write disk image: %v", err)
		createImage = false
	}

	for {
		info, err := queryInstance()

		if err != nil {
			logrus.Warnf("query instance error: %v", err)
			time.Sleep(time.Second * 10)
			continue
		}

		state := info.UHostSet[0].State
		if strings.EqualFold(state, "Stopped") {
			break
		}

		logrus.Infof("vm state %s, wait stop", state)
	}

	deleteKeyPair := hostClient.NewDeleteUHostKeyPairsRequest()
	deleteKeyPair.KeyPairIds = []string{createKeyPairResponse.KeyPairId}
	_, err = hostClient.DeleteUHostKeyPairs(deleteKeyPair)

	if err != nil {
		panic(err)
	}

	if createImage {
		imageRequest := hostClient.NewCreateCustomImageRequest()
		imageRequest.ImageName = ucloud.String(name)
		imageRequest.UHostId = ucloud.String(hostInstance.UHostIds[0])
		imageResponse, err := hostClient.CreateCustomImage(imageRequest)

		if err != nil {
			panic(err)
		}

		logrus.Infof("image id: %s", imageResponse.ImageId)

	imageRefresh:
		for {
			describeImage := hostClient.NewDescribeImageRequest()
			describeImage.Limit = ucloud.Int(1)
			describeImage.Region = ucloud.String(region)
			describeImage.Zone = ucloud.String(zone)
			describeImage.ImageIds = []string{imageResponse.ImageId}
			imageResp, err := hostClient.DescribeImage(describeImage)
			if err != nil {
				logrus.Errorf("describe image err: %v", err)
				time.Sleep(time.Second * 5)
				continue
			}

			for _, image := range imageResp.ImageSet {
				logrus.Infof("image id: %s => %s", image.ImageId, image.State)
				if strings.EqualFold(image.State, "Making") {
					time.Sleep(time.Second * 30)
					continue imageRefresh
				}
			}

			break
		}
	}

	termHostRequest := hostClient.NewTerminateUHostInstanceRequest()
	termHostRequest.UHostId = ucloud.String(hostInstance.UHostIds[0])
	termHostRequest.ReleaseEIP = ucloud.Bool(true)
	termHostRequest.ReleaseUDisk = ucloud.Bool(true)
	_, err = hostClient.TerminateUHostInstance(termHostRequest)
	if err != nil {
		panic(err)
	}
}
