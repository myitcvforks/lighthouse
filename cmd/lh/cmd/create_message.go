package cmd

import (
	"github.com/nwidger/lighthouse/messages"
	"github.com/spf13/cobra"
)

type createMessagesCmdOpts struct {
	title string
	body  string
}

var createMessagesCmdFlags createMessagesCmdOpts

// messageCmd represents the message command
var createMessageCmd = &cobra.Command{
	Use:   "message",
	Short: "Create a message (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		flags := createMessagesCmdFlags
		projectID := Project()
		m := messages.NewService(service, projectID)
		message := &messages.Message{
			Title: flags.title,
			Body:  flags.body,
		}
		if len(message.Title) == 0 {
			FatalUsage(cmd, "Please specify message title with --title")
		}
		if len(message.Body) == 0 {
			FatalUsage(cmd, "Please specify message body with --body")
		}
		nm, err := m.Create(message)
		if err != nil {
			FatalUsage(cmd, err)
		}
		JSON(nm)
	},
}

func init() {
	createCmd.AddCommand(createMessageCmd)
	createMessageCmd.Flags().StringVar(&createMessagesCmdFlags.title, "title", "", "Message title (required)")
	createMessageCmd.Flags().StringVar(&createMessagesCmdFlags.body, "body", "", "Message body (required)")
}
