package cmd

import (
	"github.com/nwidger/lighthouse/tickets"
	"github.com/spf13/cobra"
)

// ticketCmd represents the ticket command
var deleteTicketCmd = &cobra.Command{
	Use:   "ticket [number]",
	Short: "Delete a ticket (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		projectID := Project()
		t := tickets.NewService(service, projectID)
		if len(args) == 0 {
			FatalUsage(cmd, "must supply ticket number")
		}
		err := t.Delete(args[0])
		if err != nil {
			FatalUsage(cmd, err)
		}
	},
}

func init() {
	deleteCmd.AddCommand(deleteTicketCmd)
}
