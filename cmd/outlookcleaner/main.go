package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	InitializeDB()
}

func main() {
	var cmdAuthInit = &cobra.Command{
		Use:   "auth-init",
		Short: "Get encrypted versions of username and password to save in the secrets file.",
		Run: func(cmd *cobra.Command, args []string) {
			ReadAndInitCredentials()
		},
	}

	var authValidate = &cobra.Command{
		Use:   "auth-validate",
		Short: "Validate the credentials set in the secrets file and list the readable folders per account",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("TODO")
			// check credentials and list mailboxes
		},
	}

	var cmdIngest = &cobra.Command{
		Use:   "ingest",
		Short: "Ingest the mailbox messages into local database",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info().Strs("cmdArgs", args).Msg("Running ingest with command args")
			// Ingest()
		},
	}

	var cmdPrune = &cobra.Command{
		Use:   "prune",
		Short: "Move stale and read emails from the 'prune' folders to a review folder.",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info().Strs("cmdArgs", args).Msg("Running prune with command args")
			CustomPrune()
		},
	}
	// add single folder cli flag to use during debugging

	var rootCmd = &cobra.Command{Use: "outlook-cleaner"}
	rootCmd.AddCommand(cmdAuthInit, authValidate, cmdPrune, cmdIngest)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
