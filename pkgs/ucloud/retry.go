package ucloud

import (
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

func retryOnRetryableUcloudError[T any](op string, fn func() (T, error)) (T, error) {
	const (
		maxAttempts = 15
		minBackoff  = 1 * time.Second
		maxBackoff  = 30 * time.Second
	)

	var zero T
	backoff := minBackoff

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		var sdkErr uerr.Error
		if !errors.As(err, &sdkErr) || !sdkErr.Retryable() || attempt == maxAttempts {
			return zero, err
		}

		logrus.Warnf("%s got retryable error (attempt %d/%d), retry in %s: %v", op, attempt, maxAttempts, backoff, err)
		time.Sleep(backoff)

		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}

	return zero, errors.New("unreachable")
}
