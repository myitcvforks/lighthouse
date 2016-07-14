package cmd

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/nwidger/lighthouse/tickets"
	"github.com/spf13/cobra"
)

type updateTicketsCmdOpts struct {
	title      string
	comment    string
	state      string
	assigned   string
	milestone  string
	tags       string
	attachment string
}

var updateTicketsCmdFlags updateTicketsCmdOpts

// ticketCmd represents the ticket command
var updateTicketCmd = &cobra.Command{
	Use:   "ticket [number]",
	Short: "Update a ticket (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		flags := updateTicketsCmdFlags
		projectID := Project()
		t := tickets.NewService(service, projectID)
		if len(args) == 0 {
			log.Fatal("must supply ticket number")
		}
		number, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatal(err)
		}
		tkt, err := t.Get(number)
		if err != nil {
			log.Fatal(err)
		}
		if len(flags.attachment) > 0 {
			f, err := os.Open(flags.attachment)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			err = t.AddAttachment(tkt, filepath.Base(flags.attachment), f)
			if err != nil {
				log.Fatal(err)
			}
		}
		if len(flags.title) > 0 {
			tkt.Title = flags.title
		}
		if len(flags.comment) > 0 {
			tkt.Body = flags.comment
		}
		if len(flags.state) > 0 {
			tkt.State = flags.state
		}
		if len(flags.assigned) > 0 {
			tkt.AssignedUserID, err = UserID(flags.assigned)
			if err != nil {
				log.Fatal(err)
			}
		}
		if len(flags.milestone) > 0 {
			tkt.MilestoneID, err = MilestoneID(flags.milestone)
			if err != nil {
				log.Fatal(err)
			}
		}
		if len(flags.tags) > 0 {
			tkt.Tag = flags.tags
		}
		err = t.Update(tkt)
		if err != nil {
			log.Fatal(err)
		}
		JSON(tkt)
	},
}

func init() {
	updateCmd.AddCommand(updateTicketCmd)
	updateTicketCmd.Flags().StringVar(&updateTicketsCmdFlags.title, "title", "", "Change ticket title (optional)")
	updateTicketCmd.Flags().StringVar(&updateTicketsCmdFlags.comment, "comment", "", "Add a ticket comment (optional)")
	updateTicketCmd.Flags().StringVar(&updateTicketsCmdFlags.state, "state", "", "Change ticket state (optional)")
	updateTicketCmd.Flags().StringVar(&updateTicketsCmdFlags.assigned, "assigned", "", "Change user assigned to ticket (optional)")
	updateTicketCmd.Flags().StringVar(&updateTicketsCmdFlags.milestone, "milestone", "", "Assign ticket to a milestone (optional)")
	updateTicketCmd.Flags().StringVar(&updateTicketsCmdFlags.tags, "tags", "", "Comma-separated tags (optional)")
	updateTicketCmd.Flags().StringVar(&updateTicketsCmdFlags.attachment, "attachment", "", "Add file as attachment to ticket (optional)")
}
