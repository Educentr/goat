package singbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	errors "github.com/go-faster/errors"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	common "github.com/Educentr/goat/services/common"
)

// Env contains the sing-box container and connection details
type Env struct {
	testcontainers.Container
	SOCKS5ProxyURL string // socks5://host:port
	HTTPProxyURL   string // http://host:port
	Host           string
	SOCKS5Port     string
	HTTPPort       string
}

// getDefaultImage returns the sing-box image to use.
// It checks the SINGBOX_IMAGE environment variable first,
// falling back to the custom GHCR image if not set.
func getSingBoxImage() string {
	if img := os.Getenv("SINGBOX_IMAGE"); img != "" {
		return img
	}
	// Use custom sing-box image from GHCR with additional tools (curl, wget, time, etc.)
	// This image is built from Dockerfile-sing-box with all necessary utilities for VPN testing
	return "ghcr.io/sagernet/sing-box"
}

// WithNetworks returns a customizer that connects the container to the specified networks
func WithNetworks(networks []string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.ContainerRequest.Networks = networks
		return nil
	}
}

// CreateNetworkWithMTU creates a Docker bridge network with specified MTU
// This is important for VPN connections which may have MTU < 1500 on the host
// Returns network ID and cleanup function
func CreateNetworkWithMTU(ctx context.Context, mtu int) (networkID string, cleanup func() error, err error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", nil, errors.Errorf("failed to create docker client: %w", err)
	}

	networkName := fmt.Sprintf("singbox-mtu%d-%d", mtu, time.Now().Unix())

	networkResp, err := cli.NetworkCreate(ctx, networkName, network.CreateOptions{
		Driver: "bridge",
		Options: map[string]string{
			"com.docker.network.driver.mtu": fmt.Sprintf("%d", mtu),
		},
		Labels: map[string]string{
			"goat.singbox": "true",
		},
	})
	if err != nil {
		cli.Close()
		return "", nil, errors.Errorf("failed to create network: %w", err)
	}

	cleanup = func() error {
		defer cli.Close()
		return cli.NetworkRemove(ctx, networkResp.ID)
	}

	return networkName, cleanup, nil
}

// singBoxConfig represents the structure of sing-box config we need to parse
type singBoxConfig struct {
	Inbounds []struct {
		Type       string `json:"type"`
		Listen     string `json:"listen"`
		ListenPort int    `json:"listen_port"`
	} `json:"inbounds"`
}

// parseConfigPorts reads the sing-box config and extracts SOCKS5 and HTTP proxy ports
// Returns 0 for ports if not found (e.g., for TUN-only configs)
func parseConfigPorts(configPath string) (socks5Port, httpPort int, hasTUN bool, err error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return 0, 0, false, errors.Errorf("failed to read config file: %w", err)
	}

	var config singBoxConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return 0, 0, false, errors.Errorf("failed to parse config JSON: %w", err)
	}

	// Debug: print inbounds found
	fmt.Printf("[sing-box DEBUG] Found %d inbounds in config\n", len(config.Inbounds))
	for i, inbound := range config.Inbounds {
		fmt.Printf("[sing-box DEBUG] Inbound %d: type=%s, listen_port=%d\n", i, inbound.Type, inbound.ListenPort)
	}

	// Find SOCKS5, HTTP, and TUN inbounds
	for _, inbound := range config.Inbounds {
		switch inbound.Type {
		case "socks":
			socks5Port = inbound.ListenPort
		case "http":
			httpPort = inbound.ListenPort
		case "mixed": // Mixed port supports both SOCKS5 and HTTP
			if socks5Port == 0 {
				socks5Port = inbound.ListenPort
			}
			if httpPort == 0 {
				httpPort = inbound.ListenPort
			}
		case "tun":
			hasTUN = true
		}
	}

	// TUN-only config is valid (for testing production configs)
	if socks5Port == 0 && httpPort == 0 && !hasTUN {
		return 0, 0, false, errors.New("no SOCKS5, HTTP, or TUN inbound found in config")
	}

	return socks5Port, httpPort, hasTUN, nil
}

// WithConfigFile sets a custom config file path for sing-box.
// This option is REQUIRED - sing-box will not start without a config file.
func WithConfigFile(configPath string) testcontainers.ContainerCustomizer {
	return testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
		// Add the config file to the container
		req.Files = append(req.Files, testcontainers.ContainerFile{
			HostFilePath:      configPath,
			ContainerFilePath: "/etc/sing-box/config.json",
			FileMode:          0644,
		})
		return nil
	})
}

// Run starts a sing-box container with the provided configuration
func Run(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Env, error) {
	// Create initial request
	req := testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:             getSingBoxImage(),
			ImageSubstitutors: common.ImageSubstitutors(),
			Cmd:               []string{"run", "-c", "/etc/sing-box/config.json"},
			Env: map[string]string{
				"ENABLE_DEPRECATED_SPECIAL_OUTBOUNDS": "true",
			},
		},
	}

	// Apply user-provided customizations
	for _, e := range opts {
		_ = e.Customize(&req) //nolint:errcheck // options pattern, errors handled during container creation
	}

	// Validate that config file was provided
	hasConfig := false
	var configPath string
	for _, file := range req.Files {
		if file.ContainerFilePath == "/etc/sing-box/config.json" {
			hasConfig = true
			configPath = file.HostFilePath
			break
		}
	}
	if !hasConfig {
		return nil, errors.New("sing-box config file is required: use singbox.WithConfigFile() to specify the config path")
	}

	// Parse config to determine which ports to expose and if TUN is present
	socks5Port, httpPort, hasTUN, err := parseConfigPorts(configPath)
	if err != nil {
		return nil, errors.Errorf("failed to parse config ports: %w", err)
	}

	// Build list of exposed ports
	exposedPorts := []string{}
	var waitPort nat.Port
	var waitStrategy wait.Strategy

	if socks5Port != 0 {
		portStr := fmt.Sprintf("%d/tcp", socks5Port)
		exposedPorts = append(exposedPorts, portStr)
		waitPort = nat.Port(portStr)
	}
	if httpPort != 0 {
		portStr := fmt.Sprintf("%d/tcp", httpPort)
		exposedPorts = append(exposedPorts, portStr)
		if waitPort == "" {
			waitPort = nat.Port(portStr)
		}
	}

	// Determine wait strategy
	if waitPort != "" {
		// If we have SOCKS5 or HTTP port, wait for it
		waitStrategy = wait.ForListeningPort(waitPort)
	} else if hasTUN {
		// For TUN-only config, wait for TUN interface to be created
		// Use ip addr show to check for tun interface presence
		waitStrategy = wait.ForExec([]string{"ip", "addr", "show"}).
			WithResponseMatcher(func(body io.Reader) bool {
				output, err := io.ReadAll(body)
				if err != nil {
					return false
				}
				// Check if output contains "tun" which indicates TUN interface is present
				return bytes.Contains(output, []byte("tun"))
			}).
			WithPollInterval(1 * time.Second).
			WithStartupTimeout(60 * time.Second)
	} else {
		return nil, errors.New("cannot determine wait strategy: no proxy ports or TUN inbound")
	}

	// Set exposed ports and wait strategy
	req.ContainerRequest.ExposedPorts = exposedPorts
	req.ContainerRequest.WaitingFor = waitStrategy

	// Add Privileged mode for TUN support
	// TUN requires access to /dev/net/tun and various network capabilities
	// Privileged mode is the most reliable way to enable TUN in containers for testing
	if hasTUN {
		req.ContainerRequest.Privileged = true
		fmt.Printf("[sing-box DEBUG] Enabling privileged mode for TUN support\n")
	}

	// Start the container
	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		// Try to get container logs if container was created but failed to start
		var logsStr string
		if container != nil {
			logs, logErr := container.Logs(ctx)
			if logErr == nil {
				logBytes, readErr := io.ReadAll(logs)
				if readErr == nil && len(logBytes) > 0 {
					// Limit logs to 2000 characters for readability
					logsStr = string(logBytes)
					if len(logsStr) > 2000 {
						logsStr = logsStr[:2000] + "\n... (truncated)"
					}
				}
			}
		}

		if logsStr != "" {
			return nil, errors.Errorf("failed to start sing-box container: %w\nContainer logs:\n%s", err, logsStr)
		}
		return nil, errors.Errorf("failed to start sing-box container: %w", err)
	}

	// Get host and mapped ports
	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx) //nolint:errcheck // best effort cleanup on error
		return nil, errors.Errorf("failed to get host: %w", err)
	}

	env := &Env{
		Container: container,
		Host:      host,
	}

	// Get mapped SOCKS5 port
	if socks5Port != 0 {
		mappedPort, err := container.MappedPort(ctx, nat.Port(fmt.Sprintf("%d/tcp", socks5Port)))
		if err != nil {
			_ = container.Terminate(ctx) //nolint:errcheck // best effort cleanup on error
			return nil, errors.Errorf("failed to get mapped SOCKS5 port: %w", err)
		}
		env.SOCKS5Port = mappedPort.Port()
		env.SOCKS5ProxyURL = fmt.Sprintf("socks5://%s:%s", host, env.SOCKS5Port)
	}

	// Get mapped HTTP port
	if httpPort != 0 {
		mappedPort, err := container.MappedPort(ctx, nat.Port(fmt.Sprintf("%d/tcp", httpPort)))
		if err != nil {
			_ = container.Terminate(ctx) //nolint:errcheck // best effort cleanup on error
			return nil, errors.Errorf("failed to get mapped HTTP port: %w", err)
		}
		env.HTTPPort = mappedPort.Port()
		env.HTTPProxyURL = fmt.Sprintf("http://%s:%s", host, env.HTTPPort)
	}

	return env, nil
}
