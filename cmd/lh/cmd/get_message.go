package cmd

import (
	"log"
	"strconv"

	"github.com/nwidger/lighthouse/messages"
	"github.com/spf13/cobra"
)

// messageCmd represents the message command
var messageCmd = &cobra.Command{
	Use:   "message [id]",
	Short: "Get a message (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		projectID := Project()
		m := messages.NewService(service, projectID)
		if len(args) == 0 {
			log.Fatal("must supply message ID")
		}
		msgID, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatal(err)
		}
		msg, err := m.Get(msgID)
		if err != nil {
			log.Fatal(err)
		}
		JSON(msg)
	},
}

func init() {
	getCmd.AddCommand(messageCmd)
}
