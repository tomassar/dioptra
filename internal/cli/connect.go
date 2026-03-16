package cli

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tomassar/dioptra/internal/browser"
	"github.com/tomassar/dioptra/internal/config"
	"github.com/tomassar/dioptra/internal/db"
	"github.com/tomassar/dioptra/internal/server"
	"github.com/tomassar/dioptra/internal/tunnel"
	"github.com/tomassar/dioptra/ui"
	"golang.org/x/term"
)

var connectFlags struct {
	host      string
	sshUser   string
	sshPort   int
	sshKey    string
	sshPass   string
	dbHost    string
	dbPort    int
	dbName    string
	dbUser    string
	dbPass    string
	write     bool
	noBrowser bool
}

var connectCmd = &cobra.Command{
	Use:   "connect [connection-name | user@host/db]",
	Short: "Open a dashboard for a Postgres database via SSH tunnel",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runConnect,
}

func init() {
	f := connectCmd.Flags()
	f.StringVar(&connectFlags.host, "host", "", "SSH host")
	f.StringVar(&connectFlags.sshUser, "ssh-user", "", "SSH user")
	f.IntVar(&connectFlags.sshPort, "ssh-port", 0, "SSH port")
	f.StringVar(&connectFlags.sshKey, "ssh-key", "", "Path to SSH private key")
	f.StringVar(&connectFlags.sshPass, "ssh-password", "", "SSH password (or use SSHPASS env)")
	f.StringVar(&connectFlags.dbHost, "db-host", "", "Remote DB host (default: localhost)")
	f.IntVar(&connectFlags.dbPort, "db-port", 0, "Remote DB port (default: 5432)")
	f.StringVar(&connectFlags.dbName, "db", "", "Database name")
	f.StringVar(&connectFlags.dbUser, "db-user", "", "Database user")
	f.StringVar(&connectFlags.dbPass, "db-password", "", "Database password (or use PGPASSWORD env)")
	f.BoolVar(&connectFlags.write, "write", false, "Allow write queries")
	f.BoolVar(&connectFlags.noBrowser, "no-browser", false, "Don't open browser automatically")

	rootCmd.AddCommand(connectCmd)
}

func runConnect(cmd *cobra.Command, args []string) error {
	conn, err := resolveConnection(args)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// SSH tunnel
	fmt.Fprintf(os.Stderr, "→ Connecting to %s...\n", conn.SSHHost)
	sshPass := conn.sshPass
	if sshPass == "" {
		sshPass = os.Getenv("SSHPASS")
	}
	tun, err := tunnel.Open(ctx, tunnel.Config{
		Host:       conn.SSHHost,
		User:       conn.SSHUser,
		Port:       conn.SSHPort,
		KeyPath:    conn.SSHKeyPath,
		Password:   sshPass,
		RemoteHost: conn.DBHost,
		RemotePort: conn.DBPort,
	})
	if err != nil {
		return fmt.Errorf("ssh tunnel: %w", err)
	}
	defer tun.Close()
	fmt.Fprintf(os.Stderr, "→ SSH tunnel open on port %d\n", tun.LocalPort())

	// DB password
	dbPass := conn.DBPassword()

	// Database
	fmt.Fprintf(os.Stderr, "→ Connecting to database %q...\n", conn.DBName)
	database, err := db.Connect(ctx, tun.LocalPort(), conn.DBUser, dbPass, conn.DBName, !connectFlags.write)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer database.Close()
	fmt.Fprintf(os.Stderr, "→ Database connected")
	if !connectFlags.write {
		fmt.Fprintf(os.Stderr, " (read-only)")
	}
	fmt.Fprintln(os.Stderr)

	// HTTP server
	staticFS, err := ui.FS()
	if err != nil {
		return fmt.Errorf("embedded ui: %w", err)
	}

	srv, err := server.New(database, staticFS)
	if err != nil {
		return fmt.Errorf("server: %w", err)
	}

	go func() {
		if err := srv.Start(); err != nil {
			// http.ErrServerClosed is expected on shutdown
		}
	}()

	fmt.Fprintf(os.Stderr, "\n  Dioptra dashboard: %s\n\n", srv.URL())
	fmt.Fprintf(os.Stderr, "  Press Ctrl+C to stop\n\n")

	if !connectFlags.noBrowser {
		browser.Open(srv.URL())
	}

	// Wait for Ctrl+C or SSH drop (tunnel cancels ctx on keepalive failure)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sig:
	case <-ctx.Done():
		fmt.Fprintf(os.Stderr, "\n→ SSH connection lost\n")
	}

	fmt.Fprintf(os.Stderr, "\n→ Shutting down...\n")
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutCancel()
	srv.Shutdown(shutCtx)
	cancel()

	return nil
}

type resolvedConn struct {
	SSHHost    string
	SSHUser    string
	SSHPort    int
	SSHKeyPath string
	sshPass    string
	DBHost     string
	DBPort     int
	DBName     string
	DBUser     string
	dbPass     string
}

func (r *resolvedConn) DBPassword() string {
	if r.dbPass != "" {
		return r.dbPass
	}
	if p := os.Getenv("PGPASSWORD"); p != "" {
		return p
	}
	fmt.Fprint(os.Stderr, "Database password: ")
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr) // newline after hidden input
	if err != nil {
		return ""
	}
	return string(pw)
}

// parseURIArg parses "user@host/db" or "user@host:port/db" into a resolvedConn.
func parseURIArg(arg string) (*resolvedConn, bool) {
	// Must contain @ to be treated as a URI (not a saved connection name)
	if !strings.Contains(arg, "@") {
		return nil, false
	}
	// Prefix with ssh:// so url.Parse handles it properly
	u, err := url.Parse("ssh://" + arg)
	if err != nil || u.Host == "" {
		return nil, false
	}
	dbname := strings.TrimPrefix(u.Path, "/")
	if dbname == "" {
		return nil, false
	}
	conn := &resolvedConn{
		SSHHost: u.Hostname(),
		SSHUser: u.User.Username(),
		DBName:  dbname,
	}
	if p := u.Port(); p != "" {
		fmt.Sscanf(p, "%d", &conn.SSHPort)
	}
	return conn, true
}

func resolveConnection(args []string) (*resolvedConn, error) {
	conn := &resolvedConn{
		SSHHost:    connectFlags.host,
		SSHUser:    connectFlags.sshUser,
		SSHPort:    connectFlags.sshPort,
		SSHKeyPath: connectFlags.sshKey,
		sshPass:    connectFlags.sshPass,
		DBHost:     connectFlags.dbHost,
		DBPort:     connectFlags.dbPort,
		DBName:     connectFlags.dbName,
		DBUser:     connectFlags.dbUser,
		dbPass:     connectFlags.dbPass,
	}

	// Load from config if name given
	if len(args) > 0 {
		// Check if it looks like user@host/db first
		if uri, ok := parseURIArg(args[0]); ok {
			// URI overrides flags for host/user/db, flags still override URI
			if conn.SSHHost == "" {
				conn.SSHHost = uri.SSHHost
			}
			if conn.SSHUser == "" {
				conn.SSHUser = uri.SSHUser
			}
			if conn.SSHPort == 0 {
				conn.SSHPort = uri.SSHPort
			}
			if conn.DBName == "" {
				conn.DBName = uri.DBName
			}
		} else {
			cfg, err := config.Load()
			if err != nil {
				return nil, fmt.Errorf("load config: %w", err)
			}
			saved, err := cfg.Get(args[0])
			if err != nil {
				return nil, err
			}

			// Config values as defaults, flags override
			if conn.SSHHost == "" {
				conn.SSHHost = saved.SSHHost
			}
			if conn.SSHUser == "" {
				conn.SSHUser = saved.SSHUser
			}
			if conn.SSHPort == 0 {
				conn.SSHPort = saved.SSHPort
			}
			if conn.SSHKeyPath == "" {
				conn.SSHKeyPath = saved.SSHKeyPath
			}
			if conn.DBHost == "" {
				conn.DBHost = saved.DBHost
			}
			if conn.DBPort == 0 {
				conn.DBPort = saved.DBPort
			}
			if conn.DBName == "" {
				conn.DBName = saved.DBName
			}
			if conn.DBUser == "" {
				conn.DBUser = saved.DBUser
			}
			if conn.dbPass == "" {
				conn.dbPass = saved.DBPassword
			}
			if conn.sshPass == "" {
				conn.sshPass = saved.SSHPassword
			}
		}
	}

	if conn.SSHHost == "" {
		return nil, fmt.Errorf("--host is required (or use a saved connection name, or user@host/db)")
	}
	if conn.DBName == "" {
		return nil, fmt.Errorf("--db is required")
	}
	if conn.DBUser == "" {
		conn.DBUser = "postgres"
	}

	return conn, nil
}
