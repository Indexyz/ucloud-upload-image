package power

import (
	"github.com/indexyz/ucloud-upload-image/pkgs/sshutil"
	"golang.org/x/crypto/ssh"
)

func PowerOff(client *ssh.Client) error {
	_, err := sshutil.Run(client, "systemctl poweroff", nil)
	return err
}
