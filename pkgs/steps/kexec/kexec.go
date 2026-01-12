package kexec

import (
	"fmt"
	"net/http"

	"github.com/pingcap/errors"

	"github.com/indexyz/ucloud-upload-image/pkgs/sshutil"
	"github.com/indexyz/ucloud-upload-image/pkgs/utils"
	"golang.org/x/crypto/ssh"
)

var imageUrl = utils.EnvOr("KEXEC_IMAGE_OVERRIDE", "https://github.com/nix-community/nixos-images/releases/latest/download/nixos-kexec-installer-noninteractive-x86_64-linux.tar.gz")

func RunInstanceIntoKexec(client *ssh.Client, preferLocal bool) error {
	var err error
	if preferLocal {
		resp, err := http.Get(imageUrl)
		if err != nil {
			return errors.WithStack(err)
		}

		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return errors.Errorf("status code: %s", resp.Status)
		}

		_, err = sshutil.Run(client, "dd of=kexec.tar.gz", resp.Body)
		if err != nil {
			return errors.WithStack(err)
		}
	} else {
		err = sshutil.RunStdout(client, fmt.Sprintf("wget -O kexec.tar.gz %q", imageUrl), sshutil.NewNullReader())
		if err != nil {
			return errors.WithStack(err)
		}
	}

	_, err = sshutil.Run(client, "tar zxvf kexec.tar.gz && sync && ./kexec/run", sshutil.NewNullReader())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
