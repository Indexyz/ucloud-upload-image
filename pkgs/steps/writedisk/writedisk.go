package writedisk

import (
	"fmt"
	"os"
	"strings"

	"github.com/indexyz/ucloud-upload-image/pkgs/sshutil"
	"github.com/pingcap/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func WriteDiskImage(client *ssh.Client, url string, format string) error {
	logrus.Infof("==> clean disk")
	err := sshutil.RunStdout(client, "wipefs -af /dev/vda", nil)
	if err != nil {
		return err
	}

	cmd := ""
	localMode := true

	if info, err := os.Stat(url); err != nil || info.Mode().IsDir() {
		cmd = fmt.Sprintf("wget --no-verbose -O - %q | ", url)
		localMode = false
	}

	switch strings.ToLower(format) {
	case "bz2":
		cmd += "bzip2 -cd | "
	case "xz":
		cmd += "xz -cd | "
	case "zstd":
		cmd += "zstd -cd | "
	case "raw":
	default:
		return errors.Errorf("format %s not supported", format)
	}

	cmd += "dd of=/dev/vda status=progress bs=4M"

	if localMode {
		file, err := os.OpenFile(url, os.O_RDONLY, 0o644)
		if err != nil {
			return errors.WithStack(err)
		}

		logrus.Infof("==> start write to disk from local")

		sshutil.RunStdout(client, cmd, file)
	} else {
		logrus.Infof("==> start write to disk from remote")

		sshutil.RunStdout(client, cmd, nil)
	}

	return nil
}
