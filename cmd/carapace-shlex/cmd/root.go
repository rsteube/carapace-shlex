package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/rsteube/carapace"
	"github.com/rsteube/carapace-bridge/pkg/actions/bridge"
	shlex "github.com/rsteube/carapace-shlex"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:  "carapace-spec",
	Long: "simple shell lexer",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tokens, err := shlex.Split(args[0])
		if err != nil {
			return err
		}
		m, err := json.MarshalIndent(tokens, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(m))
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	carapace.Gen(rootCmd).PositionalCompletion(
		bridge.ActionCarapaceBin().SplitP(),
	)
}
