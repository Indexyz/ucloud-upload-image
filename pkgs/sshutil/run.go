package sshutil

import (
	"io"
	"os"

	"golang.org/x/crypto/ssh"
)

func Run(client *ssh.Client, cmd string, stdin io.Reader) ([]byte, error) {
	sess, err := client.NewSession()

	if err != nil {
		return nil, err
	}
	defer func() { _ = sess.Close() }()

	if stdin != nil {
		sess.Stdin = stdin
	}
	return sess.CombinedOutput(cmd)
}

func RunStdout(client *ssh.Client, cmd string, stdin io.Reader) error {
	sess, err := client.NewSession()

	if err != nil {
		return err
	}
	defer func() { _ = sess.Close() }()

	if stdin != nil {
		sess.Stdin = stdin
		sess.Stdout = os.Stdout
	}

	return sess.Run(cmd)
}
