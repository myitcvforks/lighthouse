package cmd

import (
	"log"
	"strconv"

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
			log.Fatal("must supply ticket number")
		}
		number, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatal(err)
		}
		err = t.Delete(number)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	deleteCmd.AddCommand(deleteTicketCmd)
}
