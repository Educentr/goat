package common

import (
	"os"
	"testing"
)

func TestDockerProxy(t *testing.T) {
	// Store original env var to restore later
	originalProxy := os.Getenv("DOCKER_PROXY")
	defer func() {
		if originalProxy != "" {
			if err := os.Setenv("DOCKER_PROXY", originalProxy); err != nil {
				t.Logf("Failed to restore DOCKER_PROXY: %v", err)
			}
		} else {
			if err := os.Unsetenv("DOCKER_PROXY"); err != nil {
				t.Logf("Failed to unset DOCKER_PROXY: %v", err)
			}
		}
	}()

	tests := []struct {
		name        string
		input       string
		proxyEnv    string
		expected    string
		description string
	}{
		{
			name:        "no proxy env var",
			input:       "redis:7.2.2-alpine",
			proxyEnv:    "",
			expected:    "redis:7.2.2-alpine",
			description: "should return original URL when no DOCKER_PROXY is set",
		},
		{
			name:        "with proxy and no prefix to trim",
			input:       "redis:7.2.2-alpine",
			proxyEnv:    "my-registry.com",
			expected:    "my-registry.com/redis:7.2.2-alpine",
			description: "should prepend proxy URL when no trimming is needed",
		},
		{
			name:        "with proxy and docker.io prefix",
			input:       "docker.io/redis:7.2.2-alpine",
			proxyEnv:    "my-registry.com",
			expected:    "my-registry.com/redis:7.2.2-alpine",
			description: "should trim docker.io/ prefix and prepend proxy URL",
		},
		{
			name:        "with proxy and ghcr.io prefix",
			input:       "ghcr.io/foundry-rs/foundry:stable",
			proxyEnv:    "my-registry.com",
			expected:    "my-registry.com/foundry-rs/foundry:stable",
			description: "should trim ghcr.io/ prefix and prepend proxy URL",
		},
		{
			name:        "proxy with trailing slash",
			input:       "redis:7.2.2-alpine",
			proxyEnv:    "my-registry.com/",
			expected:    "my-registry.com/redis:7.2.2-alpine",
			description: "should handle proxy URL with trailing slash",
		},
		{
			name:        "complex image name with docker.io",
			input:       "docker.io/library/postgres:15.3-alpine",
			proxyEnv:    "internal-registry.company.com",
			expected:    "internal-registry.company.com/library/postgres:15.3-alpine",
			description: "should handle complex image names with docker.io prefix",
		},
		{
			name:        "image with tag and digest",
			input:       "nginx:1.20@sha256:abc123",
			proxyEnv:    "proxy.example.com",
			expected:    "proxy.example.com/nginx:1.20@sha256:abc123",
			description: "should handle images with both tag and digest",
		},
		{
			name:        "only ghcr.io prefix without additional path",
			input:       "ghcr.io/some-image",
			proxyEnv:    "my-proxy.com",
			expected:    "my-proxy.com/some-image",
			description: "should trim ghcr.io/ prefix for simple image names",
		},
		{
			name:        "image name that starts with but is not exactly a trimmed prefix",
			input:       "docker.io.example.com/my-image:latest",
			proxyEnv:    "proxy.com",
			expected:    "proxy.com/docker.io.example.com/my-image:latest",
			description: "should not trim prefixes that are not exact matches",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.proxyEnv != "" {
				if err := os.Setenv("DOCKER_PROXY", tt.proxyEnv); err != nil {
					t.Fatalf("Failed to set DOCKER_PROXY: %v", err)
				}
			} else {
				if err := os.Unsetenv("DOCKER_PROXY"); err != nil {
					t.Fatalf("Failed to unset DOCKER_PROXY: %v", err)
				}
			}

			// Run the function
			result := DockerProxy(tt.input)

			// Check result
			if result != tt.expected {
				t.Errorf("DockerProxy(%q) = %q, want %q. %s", tt.input, result, tt.expected, tt.description)
			}
		})
	}
}

func TestDockerProxyWithInvalidURL(t *testing.T) {
	// Store original env var to restore later
	originalProxy := os.Getenv("DOCKER_PROXY")
	defer func() {
		if originalProxy != "" {
			if err := os.Setenv("DOCKER_PROXY", originalProxy); err != nil {
				t.Logf("Failed to restore DOCKER_PROXY: %v", err)
			}
		} else {
			if err := os.Unsetenv("DOCKER_PROXY"); err != nil {
				t.Logf("Failed to unset DOCKER_PROXY: %v", err)
			}
		}
	}()

	// Test with invalid proxy URL that will cause url.JoinPath to fail
	if err := os.Setenv("DOCKER_PROXY", "://invalid-url"); err != nil {
		t.Fatalf("Failed to set DOCKER_PROXY: %v", err)
	}

	// This test checks that the function panics for invalid URL paths
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected DockerProxy to panic with invalid proxy URL, but it didn't")
		}
	}()

	// This should cause url.JoinPath to fail and panic
	DockerProxy("some-image:latest")
}
