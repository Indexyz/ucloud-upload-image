package ucloud

import (
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	ucloudsdk "github.com/ucloud/ucloud-sdk-go/ucloud"
)

func (c *Client) CreateTempInstance(instanceName string, imageID string, keyPairID string) ([]string, error) {
	request := c.hostClient.NewCreateUHostInstanceRequest()
	request.Name = ucloudsdk.String(instanceName)
	request.Zone = ucloudsdk.String(c.zone)
	request.CPU = ucloudsdk.Int(2)
	request.Memory = ucloudsdk.Int(4096)
	request.ImageId = ucloudsdk.String(imageID)
	request.LoginMode = ucloudsdk.String("KeyPair")
	request.KeyPairId = ucloudsdk.String(keyPairID)
	request.Disks = []uhost.UHostDisk{
		{
			IsBoot: ucloudsdk.String("True"),
			Type:   ucloudsdk.String("CLOUD_RSSD"),
			Size:   ucloudsdk.Int(20),
		},
	}

	request.NetworkInterface = []uhost.CreateUHostInstanceParamNetworkInterface{
		{
			EIP: &uhost.CreateUHostInstanceParamNetworkInterfaceEIP{
				Bandwidth:    ucloudsdk.Int(1),
				PayMode:      ucloudsdk.String("Bandwidth"),
				OperatorName: ucloudsdk.String(c.networkType),
			},
		},
	}

	request.MachineType = ucloudsdk.String("O")
	request.HotplugFeature = ucloudsdk.Bool(false)
	request.Features = &uhost.CreateUHostInstanceParamFeatures{
		UNI: ucloudsdk.Bool(false),
	}
	request.ChargeType = ucloudsdk.String("Dynamic")
	request.NetCapability = ucloudsdk.String("Normal")

	response, err := c.hostClient.CreateUHostInstance(request)
	if err != nil {
		return nil, err
	}

	return response.UHostIds, nil
}

func (c *Client) DescribeInstance(uhostIDs []string) (*uhost.DescribeUHostInstanceResponse, error) {
	request := c.hostClient.NewDescribeUHostInstanceRequest()
	request.Limit = ucloudsdk.Int(1)
	request.Region = ucloudsdk.String(c.region)
	request.Zone = ucloudsdk.String(c.zone)
	request.UHostIds = uhostIDs

	return c.hostClient.DescribeUHostInstance(request)
}

func (c *Client) WaitPublicIP(uhostIDs []string) (string, error) {
	for {
		hostInfo, err := c.DescribeInstance(uhostIDs)
		if err != nil {
			return "", err
		}

		publicIPs := lo.Filter(hostInfo.UHostSet[0].IPSet, func(item uhost.UHostIPSet, _ int) bool {
			return strings.EqualFold(item.Type, c.networkType)
		})

		if len(publicIPs) == 0 {
			logrus.Warnf("failed to fetch public ip, wait for next round")
			time.Sleep(time.Second * 5)
			continue
		}

		return publicIPs[0].IP, nil
	}
}

func (c *Client) WaitStopped(uhostIDs []string) {
	for {
		info, err := c.DescribeInstance(uhostIDs)
		if err != nil {
			logrus.Warnf("query instance error: %v", err)
			time.Sleep(time.Second * 10)
			continue
		}

		state := info.UHostSet[0].State
		if strings.EqualFold(state, "Stopped") {
			return
		}

		logrus.Infof("vm state %s, wait stop", state)
	}
}

func (c *Client) TerminateInstance(uhostID string, releaseEIP bool, releaseUDisk bool) error {
	request := c.hostClient.NewTerminateUHostInstanceRequest()
	request.UHostId = ucloudsdk.String(uhostID)
	request.ReleaseEIP = ucloudsdk.Bool(releaseEIP)
	request.ReleaseUDisk = ucloudsdk.Bool(releaseUDisk)

	_, err := c.hostClient.TerminateUHostInstance(request)
	return err
}
