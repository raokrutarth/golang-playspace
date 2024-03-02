package main

import (
	"fmt"
	"os"

	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// docs: https://github.com/spf13/cobra/blob/main/site/content/user_guide.md

func Reverse(input string) (result string) {
	for _, c := range input {
		result = string(c) + result
	}
	return result
}

func Inspect(input string, digits bool) (count int, kind string) {
	if !digits {
		return len(input), "char"
	}
	return inspectNumbers(input), "digit"
}

func inspectNumbers(input string) (count int) {
	for _, c := range input {
		_, err := strconv.Atoi(string(c))
		if err == nil {
			count++
		}
	}
	return count
}

func main() {
	var cfgFile string
	initConfig := func() {
		if cfgFile != "" {
			viper.SetConfigFile(cfgFile)
		} else {
			viper.AddConfigPath("$HOME")
			viper.SetConfigType("yaml")
			viper.SetConfigName(".cipher_cli")
		}

		viper.AutomaticEnv()
		if err := viper.ReadInConfig(); err == nil {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
	cobra.OnInitialize(initConfig)

	var version = "0.0.1"
	var rootCmd = &cobra.Command{
		Use:     "stringer",
		Short:   "stringer - a simple CLI to transform and inspect strings",
		Version: version,
		Long:    `stringer is a super fancy CLI (kidding)`,
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	var Region, u, pw string
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	rootCmd.PersistentFlags().StringP("author", "a", "YOUR NAME", "author name for copyright attribution")
	rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")
	rootCmd.MarkFlagRequired("author")
	rootCmd.PersistentFlags().StringVarP(&Region, "region", "r", "", "AWS region (required)")
	rootCmd.MarkPersistentFlagRequired("region")

	rootCmd.Flags().StringVarP(&u, "username", "u", "", "Username (required if password is set)")
	rootCmd.Flags().StringVarP(&pw, "password", "p", "", "Password (required if username is set)")
	rootCmd.MarkFlagsRequiredTogether("username", "password")

	var ofJson, ofYaml bool
	rootCmd.Flags().BoolVar(&ofJson, "json", false, "Output in JSON")
	rootCmd.Flags().BoolVar(&ofYaml, "yaml", false, "Output in YAML")
	rootCmd.MarkFlagsOneRequired("json", "yaml")
	rootCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
	viper.SetDefault("author", "NAME HERE <EMAIL ADDRESS>")
	viper.SetDefault("license", "apache")

	var reverseCmd = &cobra.Command{
		Use:     "reverse",
		Aliases: []string{"rev"},
		Short:   "Reverses a string",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			res := Reverse(args[0])
			fmt.Println(res)
		},
	}

	// go run main.go inspect A1B2C3 --digits
	var onlyDigits bool
	var inspectCmd = &cobra.Command{
		Use:     "inspect",
		Aliases: []string{"insp"},
		Short:   "Inspects a string",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			i := args[0]
			res, kind := Inspect(i, onlyDigits)

			pluralS := "s"
			if res == 1 {
				pluralS = ""
			}
			fmt.Printf("'%s' has %d %s%s.\n", i, res, kind, pluralS)
		},
	}
	inspectCmd.Flags().BoolVarP(&onlyDigits, "digits", "d", false, "Count only digits")

	rootCmd.AddCommand(reverseCmd)
	rootCmd.AddCommand(inspectCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
