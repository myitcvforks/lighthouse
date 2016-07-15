package cmd

import (
	"github.com/nwidger/lighthouse/messages"
	"github.com/spf13/cobra"
)

type updateMessagesCmdOpts struct {
	title string
	body  string
}

var updateMessagesCmdFlags updateMessagesCmdOpts

// messageCmd represents the message command
var updateMessageCmd = &cobra.Command{
	Use:   "message [id-or-title]",
	Short: "Update a message (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		flags := updateMessagesCmdFlags
		projectID := Project()
		m := messages.NewService(service, projectID)
		if len(args) == 0 {
			FatalUsage(cmd, "must supply message ID or title")
		}
		messageID, err := MessageID(args[0])
		if err != nil {
			FatalUsage(cmd, err)
		}
		message, err := m.Get(messageID)
		if err != nil {
			FatalUsage(cmd, err)
		}
		if len(flags.title) > 0 {
			message.Title = flags.title
		}
		if len(flags.body) > 0 {
			message.Body = flags.body
		}
		err = m.Update(message)
		if err != nil {
			FatalUsage(cmd, err)
		}
		message, err = m.Get(messageID)
		if err != nil {
			FatalUsage(cmd, err)
		}
		JSON(message)
	},
}

func init() {
	updateCmd.AddCommand(updateMessageCmd)
	updateMessageCmd.Flags().StringVar(&updateMessagesCmdFlags.title, "title", "", "Change message title")
	updateMessageCmd.Flags().StringVar(&updateMessagesCmdFlags.body, "body", "", "Change message body")
}
