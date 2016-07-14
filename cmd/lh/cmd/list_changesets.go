package cmd

import (
	"github.com/nwidger/lighthouse/changesets"
	"github.com/spf13/cobra"
)

// changesetsCmd represents the changesets command
var changesetsCmd = &cobra.Command{
	Use:   "changesets",
	Short: "List changesets (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		projectID := Project()
		c := changesets.NewService(service, projectID)
		cs, err := c.List()
		if err != nil {
			FatalUsage(cmd, err)
		}
		JSON(cs)
	},
}

func init() {
	listCmd.AddCommand(changesetsCmd)
}
