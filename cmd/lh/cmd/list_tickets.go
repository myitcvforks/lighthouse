package cmd

import (
	"log"

	"github.com/nwidger/lighthouse/tickets"
	"github.com/spf13/cobra"
)

type ticketsCmdOpts struct {
	query string
	limit int
	page  int
	all   bool
}

var ticketsCmdFlags ticketsCmdOpts

// ticketsCmd represents the tickets command
var ticketsCmd = &cobra.Command{
	Use:   "tickets",
	Short: "List tickets (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			err error
			ts  tickets.Tickets
		)
		flags := ticketsCmdFlags
		projectID := Project()
		t := tickets.NewService(service, projectID)
		opts := &tickets.ListOptions{
			Query: flags.query,
			Limit: flags.limit,
			Page:  flags.page,
		}
		if flags.all {
			ts, err = t.ListAll(opts)
		} else {
			ts, err = t.List(opts)
		}
		if err != nil {
			log.Fatal(err)
		}
		JSON(ts)
	},
}

func init() {
	listCmd.AddCommand(ticketsCmd)
	ticketsCmd.Flags().StringVar(&ticketsCmdFlags.query, "query", "", "Search query, see http://help.lighthouseapp.com/faqs/getting-started/how-do-i-search-for-tickets")
	ticketsCmd.Flags().IntVar(&ticketsCmdFlags.limit, "limit", 0, "The number of tickets per page to return")
	ticketsCmd.Flags().IntVar(&ticketsCmdFlags.page, "page", 0, "Page to return")
	ticketsCmd.Flags().BoolVar(&ticketsCmdFlags.all, "all", false, "Return all tickets")
}
