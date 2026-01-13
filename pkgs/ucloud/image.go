package ucloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	ucloudsdk "github.com/ucloud/ucloud-sdk-go/ucloud"
)

func (c *Client) FindImageIDByNameContains(substr string) (string, error) {
	request := c.hostClient.NewDescribeImageRequest()
	request.Limit = ucloudsdk.Int(9999)
	request.Region = ucloudsdk.String(c.region)

	imageResp, err := c.hostClient.DescribeImage(request)
	if err != nil {
		return "", err
	}

	images := lo.Filter(imageResp.ImageSet, func(item uhost.UHostImageSet, _ int) bool {
		if item.Zone != c.zone {
			return false
		}

		return strings.Contains(item.ImageName, substr)
	})

	if len(images) == 0 {
		return "", fmt.Errorf("image not found")
	}

	return images[0].ImageId, nil
}

func (c *Client) CreateCustomImage(uhostID string, imageName string) (string, error) {
	return retryOnRetryableUcloudError("create custom image", func() (string, error) {
		request := c.hostClient.NewCreateCustomImageRequest()
		request.ImageName = ucloudsdk.String(imageName)
		request.UHostId = ucloudsdk.String(uhostID)

		response, err := c.hostClient.CreateCustomImage(request)
		if err != nil {
			return "", err
		}

		return response.ImageId, nil
	})
}

func (c *Client) WaitImageDone(imageID string) {
imageRefresh:
	for {
		request := c.hostClient.NewDescribeImageRequest()
		request.Limit = ucloudsdk.Int(1)
		request.Region = ucloudsdk.String(c.region)
		request.Zone = ucloudsdk.String(c.zone)
		request.ImageIds = []string{imageID}

		imageResp, err := c.hostClient.DescribeImage(request)
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

		return
	}
}
