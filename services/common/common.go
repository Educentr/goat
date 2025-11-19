package common

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	testcontainers "github.com/testcontainers/testcontainers-go"
)

var proxyTrimList = []string{
	"docker.io/",
	"ghcr.io/",
}

type (
	ImageSubstitutor struct {
		proxy string
	}
)

func DockerProxy(u string) string {
	if proxy := os.Getenv("DOCKER_PROXY"); proxy != "" {
		fmt.Println("proxy=", proxy)

		for _, prefix := range proxyTrimList {
			if strings.HasPrefix(u, prefix) {
				u = strings.TrimPrefix(u, prefix)
				fmt.Println("trimmed prefix", prefix, "from", u)
			}
		}

		p, err := url.JoinPath(proxy, u)
		if err != nil {
			panic(err)
		}

		u = p
		fmt.Println("target image ", p)
	}

	return u
}

func (a *ImageSubstitutor) Description() string {
	m := "docker proxy substitutor %s"
	if a.proxy != "" {
		return fmt.Sprintf(m, fmt.Sprintf("to %s", a.proxy))
	}

	return fmt.Sprintf(m, "is disabled")
}

func (a *ImageSubstitutor) Substitute(image string) (string, error) {
	if a.proxy != "" && !strings.HasPrefix(image, a.proxy) {
		p, err := url.JoinPath(a.proxy, image)
		if err != nil {
			return "", err
		}

		image = p
		fmt.Println("target image ", image)
	}

	return image, nil
}

func NewImageSubstitutor() *ImageSubstitutor {
	return &ImageSubstitutor{
		proxy: os.Getenv("DOCKER_PROXY"), // I don't want to change CI config, so let's keep env name
	}
}

func ImageSubstitutors() []testcontainers.ImageSubstitutor {
	return []testcontainers.ImageSubstitutor{NewImageSubstitutor()}
}
