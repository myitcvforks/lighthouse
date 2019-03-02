package cmd

import (
	"github.com/nwidger/lighthouse/messages"
	"github.com/spf13/cobra"
)

// messageCmd represents the message command
var deleteMessageCmd = &cobra.Command{
	Use:   "message [id-or-title]",
	Short: "Delete a message (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		projectID := Project()
		m := messages.NewService(service, projectID)
		if len(args) == 0 {
			FatalUsage(cmd, "must supply message ID or title")
		}
		err := m.Delete(args[0])
		if err != nil {
			FatalUsage(cmd, err)
		}
	},
}

func init() {
	deleteCmd.AddCommand(deleteMessageCmd)
}
