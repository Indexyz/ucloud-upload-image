package task

import (
	"context"
	"time"

	"github.com/5aaee9/ucloud-upload-image/pkgs/sshutil"
	"github.com/5aaee9/ucloud-upload-image/pkgs/steps/kexec"
	"github.com/5aaee9/ucloud-upload-image/pkgs/steps/power"
	"github.com/5aaee9/ucloud-upload-image/pkgs/steps/writedisk"
	"github.com/pingcap/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type TaskContext struct {
	ConnectIP  string
	PrivateKey []byte

	ImagePath   string
	ImageFormat string
}

func RunStepHost(task *TaskContext) error {
	signer, err := ssh.ParsePrivateKey(task.PrivateKey)
	if err != nil {
		return errors.WithStack(err)
	}

	sshClientConfig := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		// There is no way to get the host key of the rescue system beforehand
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	client, err := sshutil.Connect(context.Background(), task.ConnectIP, sshClientConfig)
	if err != nil {
		return errors.WithStack(err)
	}
	logrus.Infof("start run vm into kexec env")

	err = kexec.RunInstanceIntoKexec(client, true)
	if err != nil {
		return errors.WithStack(err)
	}

	logrus.Infof("client running kexec env")

	_ = client.Close()

	// Wait kexec start new kernel (script wait 7 second after success run)
	time.Sleep(time.Second * 10)

	// Wait reconnect
	client, err = sshutil.Connect(context.Background(), task.ConnectIP, sshClientConfig)
	if err != nil {
		panic(err)
	}

	err = writedisk.WriteDiskImage(client, task.ImagePath, task.ImageFormat)
	if err != nil {
		panic(err)
	}

	logrus.Infof("==> power off machine")
	err = power.PowerOff(client)
	if err != nil {
		panic(err)
	}

	_ = client.Close()

	return nil
}
