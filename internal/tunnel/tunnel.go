package tunnel

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"

	sshconfig "github.com/kevinburke/ssh_config"
)

type Config struct {
	Host     string
	User     string
	Port     int
	KeyPath  string // optional override
	Password string // optional SSH password (skips interactive prompt)

	RemoteHost string
	RemotePort int
}

type Tunnel struct {
	localPort int
	listener  net.Listener
	client    *ssh.Client
	wg        sync.WaitGroup
	cancel    context.CancelFunc
}

func (t *Tunnel) LocalPort() int { return t.localPort }

func Open(ctx context.Context, cfg Config) (*Tunnel, error) {
	ctx, cancel := context.WithCancel(ctx)
	cfg = resolveSSHConfig(cfg)

	if cfg.Port == 0 {
		cfg.Port = 22
	}
	if cfg.RemoteHost == "" {
		cfg.RemoteHost = "localhost"
	}
	if cfg.RemotePort == 0 {
		cfg.RemotePort = 5432
	}

	sshCfg := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            buildAuthMethods(cfg.KeyPath, cfg.Password),
		HostKeyCallback: hostKeyCallback(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	client, err := ssh.Dial("tcp", addr, sshCfg)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("ssh dial %s: %w", addr, err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		cancel()
		client.Close()
		return nil, fmt.Errorf("local listen: %w", err)
	}

	t := &Tunnel{
		localPort: listener.Addr().(*net.TCPAddr).Port,
		listener:  listener,
		client:    client,
		cancel:    cancel,
	}

	// Keepalive: cancel ctx (and tear down the session) if SSH drops
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_, _, err := client.SendRequest("keepalive@openssh.com", true, nil)
				if err != nil {
					cancel() // propagate SSH drop to callers
					return
				}
			}
		}
	}()

	// Accept loop
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		for {
			local, err := listener.Accept()
			if err != nil {
				return
			}
			t.wg.Add(1)
			go func() {
				defer t.wg.Done()
				t.forward(local, cfg.RemoteHost, cfg.RemotePort)
			}()
		}
	}()

	return t, nil
}

func (t *Tunnel) Close() error {
	t.cancel()
	t.listener.Close()
	t.wg.Wait()
	return t.client.Close()
}

func (t *Tunnel) forward(local net.Conn, remoteHost string, remotePort int) {
	defer local.Close()

	remote, err := t.client.Dial("tcp", fmt.Sprintf("%s:%d", remoteHost, remotePort))
	if err != nil {
		return
	}
	defer remote.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); io.Copy(remote, local) }()
	go func() { defer wg.Done(); io.Copy(local, remote) }()
	wg.Wait()
}

func resolveSSHConfig(cfg Config) Config {
	if cfg.User == "" {
		if u := sshconfig.Get(cfg.Host, "User"); u != "" {
			cfg.User = u
		} else {
			cfg.User = os.Getenv("USER")
		}
	}
	if cfg.Port == 0 {
		if p := sshconfig.Get(cfg.Host, "Port"); p != "" {
			fmt.Sscanf(p, "%d", &cfg.Port)
		}
	}
	if cfg.KeyPath == "" {
		if id := sshconfig.Get(cfg.Host, "IdentityFile"); id != "" {
			if id[0] == '~' {
				home, _ := os.UserHomeDir()
				id = filepath.Join(home, id[1:])
			}
			cfg.KeyPath = id
		}
	}
	if h := sshconfig.Get(cfg.Host, "Hostname"); h != "" {
		cfg.Host = h
	}
	return cfg
}

func buildAuthMethods(keyPath, password string) []ssh.AuthMethod {
	var methods []ssh.AuthMethod

	// SSH agent
	if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
		if conn, err := net.Dial("unix", sock); err == nil {
			methods = append(methods, ssh.PublicKeysCallback(agent.NewClient(conn).Signers))
		}
	}

	// Explicit key path
	if keyPath != "" {
		if m := keyFromFile(keyPath); m != nil {
			methods = append(methods, m)
		}
	}

	// Default key paths
	home, _ := os.UserHomeDir()
	for _, name := range []string{"id_ed25519", "id_rsa", "id_ecdsa"} {
		p := filepath.Join(home, ".ssh", name)
		if p == keyPath {
			continue
		}
		if m := keyFromFile(p); m != nil {
			methods = append(methods, m)
		}
	}

	// Password auth
	if password != "" {
		methods = append(methods, ssh.Password(password))
	} else {
		// Interactive prompt fallback
		methods = append(methods, ssh.PasswordCallback(func() (string, error) {
			fmt.Print("SSH password: ")
			var pw string
			fmt.Scanln(&pw)
			return pw, nil
		}))
	}

	return methods
}

func keyFromFile(path string) ssh.AuthMethod {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	signer, err := ssh.ParsePrivateKey(data)
	if err != nil {
		// Try with passphrase
		if _, ok := err.(*ssh.PassphraseMissingError); ok {
			fmt.Printf("Passphrase for %s: ", path)
			var pw string
			fmt.Scanln(&pw)
			signer, err = ssh.ParsePrivateKeyWithPassphrase(data, []byte(pw))
			if err != nil {
				return nil
			}
			return ssh.PublicKeys(signer)
		}
		return nil
	}
	return ssh.PublicKeys(signer)
}

func hostKeyCallback() ssh.HostKeyCallback {
	home, _ := os.UserHomeDir()
	knownHostsFile := filepath.Join(home, ".ssh", "known_hosts")
	cb, err := knownhosts.New(knownHostsFile)
	if err != nil {
		return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			fmt.Fprintf(os.Stderr, "Warning: cannot read known_hosts, accepting key for %s\n", hostname)
			return nil
		}
	}
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := cb(hostname, remote, key)
		if err == nil {
			return nil
		}

		var keyErr *knownhosts.KeyError
		if errors.As(err, &keyErr) && len(keyErr.Want) > 0 {
			return fmt.Errorf("HOST KEY CHANGED for %s — possible MITM attack", hostname)
		}

		fmt.Fprintf(os.Stderr, "Unknown host %s. Fingerprint: %s\nAccept? [y/N]: ", hostname, ssh.FingerprintSHA256(key))
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			return fmt.Errorf("host key rejected by user")
		}

		line := knownhosts.Line([]string{knownhosts.Normalize(hostname)}, key)
		f, err := os.OpenFile(knownHostsFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return nil
		}
		defer f.Close()
		fmt.Fprintln(f, line)
		return nil
	}
}
