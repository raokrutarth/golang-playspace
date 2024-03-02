package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	InitConfig()
	InitializeDB()
}

func main() {

	var cmdAuthInit = &cobra.Command{
		Use:   "auth-init",
		Short: "Get encrypted versions of username and password to save in the config.",
		Run: func(cmd *cobra.Command, args []string) {
			ReadAndInitCredentials()
		},
	}

	var cmdPrune = &cobra.Command{
		Use:   "prune",
		Short: "Move old and read emails for selected folder to a review folder.",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info().Strs("cmdArgs", args).Msg("Running prune with command args")
			CustomPrune()
		},
	}

	var cmdIngest = &cobra.Command{
		Use:   "ingest",
		Short: "Ingest the mailbox messages into local database",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info().Strs("cmdArgs", args).Msg("Running ingest with command args")
			Ingest()
		},
	}

	var cmdBulkDelete = &cobra.Command{
		Use:   "bulk-delete",
		Short: "Bulk remove messages matching given conditions",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info().Strs("cmdArgs", args).Msg("Running prune with command args")
			BulkDelete([]string{})
		},
	}

	var rootCmd = &cobra.Command{Use: "irl-archive-cli"}
	rootCmd.AddCommand(cmdAuthInit, cmdPrune, cmdIngest, cmdBulkDelete)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}

}
