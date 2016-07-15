package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/nwidger/lighthouse/tickets"
	"github.com/spf13/cobra"
)

type getTicketCmdOpts struct {
	title      string
	comment    string
	state      string
	assigned   string
	milestone  string
	tags       string
	attachment string
}

var getTicketCmdFlags getTicketCmdOpts

// ticketCmd represents the ticket command
var ticketCmd = &cobra.Command{
	Use:   "ticket [number]",
	Short: "Get a ticket (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		flags := getTicketCmdFlags
		projectID := Project()
		t := tickets.NewService(service, projectID)
		if len(args) == 0 {
			FatalUsage(cmd, "must supply ticket number")
		}
		number, err := TicketID(args[0])
		if err != nil {
			FatalUsage(cmd, err)
		}
		ticket, err := t.Get(number)
		if err != nil {
			FatalUsage(cmd, err)
		}
		if len(flags.attachment) == 0 {
			JSON(ticket)
		} else {
			var attachment *tickets.Attachment
			for _, a := range ticket.Attachments {
				if a.Attachment.Filename == flags.attachment {
					attachment = a.Attachment
					break
				}
			}
			if attachment == nil {
				FatalUsage(cmd, fmt.Sprintf("no such attachment with filename %q", flags.attachment))
			}
			r, err := t.GetAttachment(attachment)
			if err != nil {
				FatalUsage(cmd, err)
			}
			defer r.Close()
			io.Copy(os.Stdout, r)
		}
	},
}

func init() {
	getCmd.AddCommand(ticketCmd)
	ticketCmd.Flags().StringVar(&getTicketCmdFlags.attachment, "attachment", "", "Download ticket attachment by filename (prints attachment to standard out)")
}
