package cmd

import (
	"fmt"

	shlex "github.com/rsteube/carapace-shlex"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:  "carapace-spec",
	Long: "TODO",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		splitted, err := shlex.Split(args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%#v", splitted.Strings())
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}
