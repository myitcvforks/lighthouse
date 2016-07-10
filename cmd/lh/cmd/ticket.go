package cmd

import (
	"log"
	"strconv"

	"github.com/nwidger/lighthouse/tickets"
	"github.com/spf13/cobra"
)

// ticketCmd represents the ticket command
var ticketCmd = &cobra.Command{
	Use:   "ticket [number]",
	Short: "Get a ticket (requires -p)",
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
		ticket, err := t.Get(number)
		if err != nil {
			log.Fatal(err)
		}
		JSON(ticket)
	},
}

func init() {
	getCmd.AddCommand(ticketCmd)
}
