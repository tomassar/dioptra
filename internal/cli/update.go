package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tomassar/dioptra/internal/update"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update dioptra to the latest release",
	RunE: func(cmd *cobra.Command, args []string) error {
		if update.Version == "dev" {
			fmt.Println("Running from source — nothing to update.")
			return nil
		}

		fmt.Println("Checking for updates...")
		latest, err := update.LatestVersion()
		if err != nil {
			return fmt.Errorf("could not check for updates: %w", err)
		}

		if !update.IsNewer(update.Version, latest) {
			fmt.Printf("You're already on the latest version (%s).\n", update.Version)
			return nil
		}

		fmt.Printf("New version available: %s → %s\n", update.Version, latest)
		return update.Do(latest)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
