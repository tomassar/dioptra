package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tomassar/dioptra/internal/config"
)

var saveCmd = &cobra.Command{
	Use:   "save <name>",
	Short: "Save a connection profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runSave,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved connections",
	RunE:  runConfigList,
}

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a saved connection",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigRemove,
}

var saveFlags struct {
	sshHost string
	sshUser string
	sshPort int
	sshKey  string
	sshPass string
	dbHost  string
	dbPort  int
	dbName  string
	dbUser  string
	dbPass  string
}

func init() {
	f := saveCmd.Flags()
	f.StringVar(&saveFlags.sshHost, "host", "", "SSH host (required)")
	f.StringVar(&saveFlags.sshUser, "ssh-user", "", "SSH user")
	f.IntVar(&saveFlags.sshPort, "ssh-port", 0, "SSH port")
	f.StringVar(&saveFlags.sshKey, "ssh-key", "", "Path to SSH private key")
	f.StringVar(&saveFlags.sshPass, "ssh-password", "", "SSH password")
	f.StringVar(&saveFlags.dbHost, "db-host", "", "Remote DB host")
	f.IntVar(&saveFlags.dbPort, "db-port", 0, "Remote DB port")
	f.StringVar(&saveFlags.dbName, "db", "", "Database name (required)")
	f.StringVar(&saveFlags.dbUser, "db-user", "", "Database user")
	f.StringVar(&saveFlags.dbPass, "db-password", "", "Database password")

	saveCmd.MarkFlagRequired("host")
	saveCmd.MarkFlagRequired("db")

	rootCmd.AddCommand(saveCmd, listCmd, removeCmd)
}

func runSave(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	conn := config.Connection{
		Name:        args[0],
		SSHHost:     saveFlags.sshHost,
		SSHUser:     saveFlags.sshUser,
		SSHPort:     saveFlags.sshPort,
		SSHKeyPath:  saveFlags.sshKey,
		SSHPassword: saveFlags.sshPass,
		DBHost:      saveFlags.dbHost,
		DBPort:      saveFlags.dbPort,
		DBName:      saveFlags.dbName,
		DBUser:      saveFlags.dbUser,
		DBPassword:  saveFlags.dbPass,
	}

	if err := cfg.Add(conn); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Saved connection %q\n", args[0])
	return nil
}

func runConfigList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if len(cfg.Connections) == 0 {
		fmt.Fprintln(os.Stderr, "No saved connections. Use `dioptra save <name>` to create one.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tHOST\tDB\tDB USER")
	for _, c := range cfg.Connections {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.Name, c.SSHHost, c.DBName, c.DBUser)
	}
	return w.Flush()
}

func runConfigRemove(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if err := cfg.Remove(args[0]); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Removed connection %q\n", args[0])
	return nil
}
