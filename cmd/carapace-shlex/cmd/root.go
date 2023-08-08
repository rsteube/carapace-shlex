package cmd

import (
	"encoding/json"

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

		if cmd.Flag("current").Changed {
			tokens = tokens.CurrentPipeline()
		}
		if cmd.Flag("args").Changed {
			tokens = tokens.FilterRedirects()
		}
		if cmd.Flag("words").Changed {
			tokens = tokens.Words()
		}

		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		return encoder.Encode(tokens)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().Bool("current", false, "show current pipeline")
	rootCmd.Flags().Bool("words", false, "show words")
	rootCmd.Flags().Bool("args", false, "show words")

	carapace.Gen(rootCmd).PositionalCompletion(
		bridge.ActionCarapaceBin().SplitP(),
	)
}
