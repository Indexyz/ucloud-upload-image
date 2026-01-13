package ucloud

import ucloudsdk "github.com/ucloud/ucloud-sdk-go/ucloud"

func (c *Client) ImportKeyPair(keyPairName string, publicKeyBody string) (string, error) {
	request := c.hostClient.NewImportUHostKeyPairsRequest()
	request.KeyPairName = ucloudsdk.String(keyPairName)
	request.PublicKeyBody = ucloudsdk.String(publicKeyBody)

	response, err := c.hostClient.ImportUHostKeyPairs(request)
	if err != nil {
		return "", err
	}

	return response.KeyPairId, nil
}

func (c *Client) DeleteKeyPair(keyPairID string) error {
	request := c.hostClient.NewDeleteUHostKeyPairsRequest()
	request.KeyPairIds = []string{keyPairID}
	_, err := c.hostClient.DeleteUHostKeyPairs(request)
	return err
}
