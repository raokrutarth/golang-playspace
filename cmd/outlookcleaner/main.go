package main

import (
	"fmt"
	"os"

	"github.com/raokrutarth/golang-playspace/pkg/logger"
	"github.com/spf13/cobra"
)

func init() {
	InitializeDB()
}

func main() {
	log := logger.GetLogger()
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
			log = log.With("cmd", cmd.Name())
			log.Info("running account configuration validation", "args", args)
			AuthValidate(logger.ContextWithLogger(cmd.Context(), log))
		},
	}

	// TODO: add random sampler with folder name cli args

	var cmdIngest = &cobra.Command{
		Use:   "ingest",
		Short: "Ingest the mailbox messages into local database",
		Run: func(cmd *cobra.Command, args []string) {
			log = log.With("cmd", cmd.Name())
			log.Info("running account configuration validation", "args", args)
			connections, err := NewMailAccountConnections(logger.ContextWithLogger(cmd.Context(), log))
			if err != nil {
				log.Error("failed to get account connection", "error", err)
				return
			}
			defer func() {
				for _, c := range connections {
					if err = c.client.Logout(); err != nil {
						log.Error("failed logout", "error", err)
					}
				}
			}()

			err = Ingest(logger.ContextWithLogger(cmd.Context(), log), connections)
			if err != nil {
				log.Error("failed to ingest", "error", err)
			}
		},
	}

	var cmdPrune = &cobra.Command{
		Use:   "prune",
		Short: "Move stale and read emails from the 'prune' folders to a review folder.",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info("Running prune with command args")
			CustomPrune()
		},
	}
	// add single folder cli flag to use during debugging

	// TODO: add un-prune command

	var rootCmd = &cobra.Command{Use: "outlook-cleaner"}
	rootCmd.AddCommand(cmdAuthInit, authValidate, cmdPrune, cmdIngest)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
