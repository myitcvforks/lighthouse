package cmd

import (
	"github.com/nwidger/lighthouse/messages"
	"github.com/spf13/cobra"
)

// messageCmd represents the message command
var messageCmd = &cobra.Command{
	Use:   "message [id-or-title]",
	Short: "Get a message (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		projectID := Project()
		m := messages.NewService(service, projectID)
		if len(args) == 0 {
			FatalUsage(cmd, "must supply message ID or title")
		}
		msgID, err := MessageID(args[0])
		if err != nil {
			FatalUsage(cmd, err)
		}
		msg, err := m.Get(msgID)
		if err != nil {
			FatalUsage(cmd, err)
		}
		JSON(msg)
	},
}

func init() {
	getCmd.AddCommand(messageCmd)
}
