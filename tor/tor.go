package tor

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/net/proxy"
)

// Config holds Tor configuration parameters.
type Config struct {
	Enabled     bool
	ProxyAddr   string
	ControlPort string
	DataDir     string
}

// Client represents a Tor client connection.
type Client struct {
	config  Config
	dialer  proxy.Dialer
	process *exec.Cmd
}

// NewClient creates a new Tor client and starts Tor process if needed.
func NewClient(config Config) (*Client, error) {
	if !config.Enabled {
		return &Client{
			config: config,
			dialer: proxy.Direct,
		}, nil
	}

	client := &Client{
		config: config,
	}

	// Start Tor process
	if err := client.startTor(); err != nil {
		return nil, fmt.Errorf("failed to start Tor: %v", err)
	}

	// Wait for Tor to be ready
	if err := client.waitForTor(30 * time.Second); err != nil {
		client.Stop()
		return nil, fmt.Errorf("Tor failed to start: %v", err)
	}

	// Create SOCKS5 dialer for Tor
	dialer, err := proxy.SOCKS5("tcp", config.ProxyAddr, nil, proxy.Direct)
	if err != nil {
		client.Stop()
		return nil, fmt.Errorf("failed to create SOCKS5 dialer: %v", err)
	}
	client.dialer = dialer

	fmt.Println("Tor process started successfully")
	return client, nil
}

// Dial connects to an address through Tor.
func (c *Client) Dial(network, address string) (net.Conn, error) {
	if !c.config.Enabled {
		return net.Dial(network, address)
	}

	return c.dialer.Dial(network, address)
}

// DialTimeout connects to an address through Tor with a timeout.
func (c *Client) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	if !c.config.Enabled {
		return net.DialTimeout(network, address, timeout)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return c.DialContext(ctx, network, address)
}

// DialContext connects to an address through Tor with a context.
func (c *Client) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if !c.config.Enabled {
		var d net.Dialer
		return d.DialContext(ctx, network, address)
	}

	// Use goroutine to support context cancellation
	type result struct {
		conn net.Conn
		err  error
	}

	ch := make(chan result, 1)
	go func() {
		conn, err := c.dialer.Dial(network, address)
		ch <- result{conn, err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		return res.conn, res.err
	}
}

// testConnection tests if Tor is reachable.
func (c *Client) testConnection() error {
	if !c.config.Enabled {
		return nil
	}

	// Try to connect to Tor SOCKS proxy
	conn, err := net.DialTimeout("tcp", c.config.ProxyAddr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("cannot connect to Tor proxy at %s: %v", c.config.ProxyAddr, err)
	}
	conn.Close()

	return nil
}

// IsEnabled returns whether Tor is enabled.
func (c *Client) IsEnabled() bool {
	return c.config.Enabled
}

// GetProxyAddr returns the Tor proxy address.
func (c *Client) GetProxyAddr() string {
	return c.config.ProxyAddr
}

// startTor starts the Tor process.
func (c *Client) startTor() error {
	// Create data directory
	dataDir := c.config.DataDir
	if dataDir == "" {
		// Use DATA_DIR environment variable or default
		envDataDir := os.Getenv("DATA_DIR")
		if envDataDir != "" {
			dataDir = filepath.Join(envDataDir, "tor")
		} else {
			// Use /var/lib/tor for container environments, fallback to /tmp/tor-data
			if _, err := os.Stat("/var/lib"); err == nil {
				dataDir = "/var/lib/tor"
			} else {
				dataDir = "/tmp/tor-data"
			}
		}
	}
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	// Create torrc configuration file
	torrcPath := filepath.Join(dataDir, "torrc")
	torrcContent := fmt.Sprintf(`SOCKSPort %s
DataDirectory %s
Log notice stdout
`, c.config.ProxyAddr, dataDir)

	// Write torrc file with proper permissions for tor user
	if err := os.WriteFile(torrcPath, []byte(torrcContent), 0644); err != nil {
		return fmt.Errorf("failed to write torrc: %v", err)
	}

	// Change ownership to tor user if running as root
	if os.Geteuid() == 0 {
		if err := os.Chown(torrcPath, 100, 100); err != nil {
			// If chown fails, try to make file readable by all
			os.Chmod(torrcPath, 0644)
		}
	}

	// Start Tor process as tor user if running as root
	cmd := exec.Command("tor", "-f", torrcPath)

	// Run as tor user if we're root
	if os.Geteuid() == 0 {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{
				Uid: 100, // tor user UID
				Gid: 100, // tor group GID
			},
		}
	}

	c.process = cmd
	c.process.Stdout = os.Stdout
	c.process.Stderr = os.Stderr

	if err := c.process.Start(); err != nil {
		return fmt.Errorf("failed to start tor process: %v", err)
	}

	fmt.Printf("Started Tor process (PID: %d)\n", c.process.Process.Pid)
	return nil
}

// waitForTor waits for Tor to be ready.
func (c *Client) waitForTor(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", c.config.ProxyAddr, 1*time.Second)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for Tor to start")
}

// Stop stops the Tor process.
func (c *Client) Stop() error {
	if c.process != nil && c.process.Process != nil {
		fmt.Println("Stopping Tor process...")
		if err := c.process.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill tor process: %v", err)
		}
		c.process.Wait()
		fmt.Println("Tor process stopped")
	}
	return nil
}
