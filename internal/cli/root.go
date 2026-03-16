package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tomassar/dioptra/internal/update"
)

var rootCmd = &cobra.Command{
	Use:   "dioptra [user@host/db | connection-name]",
	Short: "SSH into your VPS and browse Postgres from a local dashboard",
	Long: `Dioptra spins up an SSH tunnel to your VPS, connects to PostgreSQL,
and opens a local web dashboard to browse schemas, tables, and run queries.

No exposed ports. No configuration GUIs. Just one command.

Examples:
  dioptra root@vps.example.com/mydb
  dioptra myconn                      # saved connection
  dioptra connect --host vps.example.com --db mydb`,
	Args: cobra.MaximumNArgs(1),
	// If a positional arg is passed at the root level, delegate to connectCmd
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			// Must look like user@host/db or a saved name — hand off to connect
			return connectCmd.RunE(connectCmd, args)
		}
		// No args: show help
		return cmd.Help()
	},
	// Silence usage on RunE errors (keeps error output clean)
	SilenceUsage: true,
}

func init() {
	// Copy connect flags onto root so they work at the top level too
	// e.g. dioptra root@host/db --write
	rootCmd.Flags().AddFlagSet(connectCmd.Flags())
}

// hasURIArg returns true if the args look like a URI shorthand (not a subcommand).
func hasURIArg(args []string) bool {
	return len(args) == 1 && strings.Contains(args[0], "@")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// --version flag
	rootCmd.Version = update.Version
	rootCmd.SetVersionTemplate("dioptra {{.Version}}\n")

	// Background update nudge: check after every command, non-blocking
	cobra.OnFinalize(func() {
		if update.Version == "dev" {
			return
		}
		latest, err := update.LatestVersion()
		if err != nil {
			return // silently skip if offline
		}
		if update.IsNewer(update.Version, latest) {
			fmt.Fprintf(os.Stderr, "\n💡 dioptra %s is available (you have %s). Run 'dioptra update' to upgrade.\n", latest, update.Version)
		}
	})
}
