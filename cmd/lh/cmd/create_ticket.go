package cmd

import (
	"log"

	"github.com/nwidger/lighthouse/tickets"
	"github.com/spf13/cobra"
)

type createTicketsCmdOpts struct {
	title     string
	body      string
	state     string
	assigned  string
	milestone string
	tags      string
}

var createTicketsCmdFlags createTicketsCmdOpts

// ticketCmd represents the ticket command
var createTicketCmd = &cobra.Command{
	Use:   "ticket",
	Short: "Create a ticket (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		flags := createTicketsCmdFlags
		projectID := Project()
		t := tickets.NewService(service, projectID)
		tc := &tickets.Ticket{
			Title: flags.title,
			Body:  flags.body,
			State: flags.state,
			Tag:   flags.tags,
		}
		if len(tc.Title) == 0 {
			log.Fatal("Please specify ticket title with --title")
		}
		if len(flags.assigned) > 0 {
			tc.AssignedUserID, err = UserID(flags.assigned)
			if err != nil {
				log.Fatal(err)
			}
		}
		if len(flags.milestone) > 0 {
			tc.MilestoneID, err = MilestoneID(flags.milestone)
			if err != nil {
				log.Fatal(err)
			}
		}
		nt, err := t.Create(tc)
		if err != nil {
			log.Fatal(err)
		}
		JSON(nt)
	},
}

func init() {
	createCmd.AddCommand(createTicketCmd)
	createTicketCmd.Flags().StringVar(&createTicketsCmdFlags.title, "title", "", "Ticket title (required)")
	createTicketCmd.Flags().StringVar(&createTicketsCmdFlags.body, "body", "", "Ticket body (optional)")
	createTicketCmd.Flags().StringVar(&createTicketsCmdFlags.state, "state", "", "Ticket state (optional)")
	createTicketCmd.Flags().StringVar(&createTicketsCmdFlags.assigned, "assigned", "", "Assign ticket to a user (optional)")
	createTicketCmd.Flags().StringVar(&createTicketsCmdFlags.milestone, "milestone", "", "Assign ticket to a milestone (optional)")
	createTicketCmd.Flags().StringVar(&createTicketsCmdFlags.tags, "tags", "", "Comma-separated tags (optional)")
}
