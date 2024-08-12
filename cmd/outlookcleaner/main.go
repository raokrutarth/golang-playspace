package main

import (
	"context"
	"os"

	"github.com/raokrutarth/golang-playspace/pkg/logger"
	"github.com/spf13/cobra"
)

func main() {
	l := logger.GetLogger()
	ctx := logger.ContextWithLogger(context.Background(), l)
	if err := initDB(ctx); err != nil {
		l.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	var cmdAuthInit = &cobra.Command{
		Use:   "auth-init",
		Short: "Get encrypted versions of username and password to save in the secrets file.",
		Run: func(_ *cobra.Command, _ []string) {
			ReadAndInitCredentials(ctx)
		},
	}

	var authValidate = &cobra.Command{
		Use:   "auth-validate",
		Short: "Validate the credentials set in the secrets file and list the readable folders per account",
		Run: func(cmd *cobra.Command, args []string) {
			sl := l.With("cmd", cmd.Name())
			sl.Info("running account configuration validation", "args", args)
			AuthValidate(ctx)
		},
	}

	var cmdIngest = &cobra.Command{
		Use:   "ingest",
		Short: "Ingest the mailbox messages into local database",
		Run: func(cmd *cobra.Command, args []string) {
			sl := l.With("cmd", cmd.Name())
			sl.Info("running account configuration validation", "args", args)
			connections, err := NewMailAccountConnections(ctx)
			if err != nil {
				l.Error("failed to get account connection", "error", err)
				os.Exit(1)
			}
			defer func() {
				for _, c := range connections {
					if err = c.client.Logout(); err != nil {
						sl.Error("failed logout", "username", c.username, "error", err)
					}
				}
			}()
			err = Ingest(ctx, connections)
			if err != nil {
				sl.Error("failed to ingest", "error", err)
			}
		},
	}

	// var cmdPrune = &cobra.Command{
	// 	Use:   "prune",
	// 	Short: "Move stale and read emails from the 'prune' folders to a review folder.",
	// 	Run: func(cmd *cobra.Command, args []string) {
	// 		log.Info("Running prune with command args")
	// 		CustomPrune()
	// 	},
	// }
	// add single folder cli flag to use during debugging

	// TODO: add un-prune command

	var rootCmd = &cobra.Command{Use: "outlook-cleaner"}
	rootCmd.AddCommand(
		cmdAuthInit,
		authValidate,
		cmdIngest,
		// cmdPrune
	)
	if err := rootCmd.Execute(); err != nil {
		l.Error("failed to execute root command", "error", err)
		os.Exit(1)
	}
}
