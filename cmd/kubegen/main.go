package main // import "github.com/errordeveloper/kubegen/cmd/kubegen"

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	stdout bool
	format string
)

func main() {
	var rootCmd = &cobra.Command{
		Use: "kubegen [command]",
	}

	rootCmd.PersistentFlags().BoolVarP(&stdout, "stdout", "s", false,
		"Output to stdout instead of creating files")
	rootCmd.PersistentFlags().StringVarP(&format, "output", "o", "yaml",
		"Output format [\"yaml\" or \"json\"]")

	rootCmd.AddCommand(bundleCmd)
	rootCmd.AddCommand(moduleCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
