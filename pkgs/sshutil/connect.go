package sshutil

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func Connect(ctx context.Context, host string, config *ssh.ClientConfig) (client *ssh.Client, err error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			client, err = ssh.Dial("tcp", host+":22", config)
			if err == nil {
				return
			}

			logrus.Warnf("waiting connection: %v", err)
			time.Sleep(time.Second * 5)
		}
	}
}
