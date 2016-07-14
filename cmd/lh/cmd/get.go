package cmd

import "github.com/spf13/cobra"

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get Lighthouse resources",
}

func init() {
	RootCmd.AddCommand(getCmd)
}
