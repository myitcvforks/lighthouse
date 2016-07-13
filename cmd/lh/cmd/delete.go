package cmd

import "github.com/spf13/cobra"

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete Lighthouse resources by ID",
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}
