package cmd

import "github.com/spf13/cobra"

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Lighthouse resources",
}

func init() {
	RootCmd.AddCommand(updateCmd)
}
