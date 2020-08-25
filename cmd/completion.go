package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(genBashCompletion)
}

// genBashCompletion represents the completion command
var genBashCompletion = &cobra.Command{
	Use:   "gen-bash-completion",
	Short: "Generates bash completion scripts",
	Long: `To load completion run

	. <(bitbucket completion)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(bitbucket completion)
`,
	Run: func(cmd *cobra.Command, args []string) {
		rootCmd.GenBashCompletion(os.Stdout)
	},
	Hidden: true,
}

func init() {
	rootCmd.AddCommand(getZshCompletion)
}

// getZshCompletion represents the completion command
var getZshCompletion = &cobra.Command{
	Use:   "gen-zsh-completion",
	Short: "Generates zsh completion scripts",
	Long: `To load completion add the completion script to a file loaded by your $fpath

	gctl gen-zsh-completion > ~/.zshfuncs/_gctl
`,
	Run: func(cmd *cobra.Command, args []string) {
		rootCmd.GenZshCompletion(os.Stdout)
	},
	Hidden: true,
}
