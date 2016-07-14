package cmd

import (
	"github.com/nwidger/lighthouse/messages"
	"github.com/spf13/cobra"
)

// messagesCmd represents the messages command
var messagesCmd = &cobra.Command{
	Use:   "messages",
	Short: "List messages (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		projectID := Project()
		m := messages.NewService(service, projectID)
		ms, err := m.List()
		if err != nil {
			FatalUsage(cmd, err)
		}
		JSON(ms)
	},
}

func init() {
	listCmd.AddCommand(messagesCmd)
}
