package ucloud

import (
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	ucloudsdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type Options struct {
	Region      string
	Zone        string
	NetworkType string

	PublicKey  string
	PrivateKey string
}

type Client struct {
	region      string
	zone        string
	networkType string

	hostClient *uhost.UHostClient
}

func New(opts Options) *Client {
	config := ucloudsdk.NewConfig()
	config.Region = opts.Region

	credential := auth.NewCredential()
	credential.PublicKey = opts.PublicKey
	credential.PrivateKey = opts.PrivateKey

	return &Client{
		region:      opts.Region,
		zone:        opts.Zone,
		networkType: opts.NetworkType,
		hostClient:  uhost.NewClient(&config, &credential),
	}
}
