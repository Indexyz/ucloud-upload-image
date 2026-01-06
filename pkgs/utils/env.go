package utils

import "os"

func EnvOr(env string, def string) string {
	envVar := os.Getenv(env)
	if len(envVar) > 0 {
		return envVar
	}

	return def
}
