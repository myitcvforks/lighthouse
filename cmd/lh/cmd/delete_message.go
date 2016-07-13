package cmd

import (
	"log"
	"strconv"

	"github.com/nwidger/lighthouse/messages"
	"github.com/spf13/cobra"
)

// messageCmd represents the message command
var deleteMessageCmd = &cobra.Command{
	Use:   "message [id]",
	Short: "Delete a message (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		projectID := Project()
		m := messages.NewService(service, projectID)
		if len(args) == 0 {
			log.Fatal("must supply message ID")
		}
		messageID, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatal(err)
		}
		err = m.Delete(messageID)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	deleteCmd.AddCommand(deleteMessageCmd)
}
