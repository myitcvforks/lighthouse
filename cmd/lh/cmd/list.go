package cmd

import "github.com/spf13/cobra"

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Lighthouse resources",
}

func init() {
	RootCmd.AddCommand(listCmd)
}
