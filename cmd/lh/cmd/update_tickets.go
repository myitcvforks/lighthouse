package cmd

import (
	"fmt"

	"github.com/nwidger/lighthouse/tickets"
	"github.com/spf13/cobra"
)

type updateTicketssCmdOpts struct {
	all            bool
	query          string
	command        string
	migrationToken string
}

var updateTicketssCmdFlags updateTicketssCmdOpts

// ticketCmd represents the ticket command
var updateTicketsCmd = &cobra.Command{
	Use:   "tickets",
	Short: "Bulk update tickets (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		flags := updateTicketssCmdFlags
		projectID := Project()
		t := tickets.NewService(service, projectID)
		opts := &tickets.BulkEditOptions{
			Query:          flags.query,
			Command:        flags.command,
			MigrationToken: flags.migrationToken,
		}
		if flags.all {
			if len(flags.query) > 0 {
				FatalUsage(cmd, fmt.Errorf("cannot use --all with --query"))
			}
			opts.Query = "all"
		}
		if len(flags.query) == 0 {
			FatalUsage(cmd, fmt.Errorf("must supply query"))
		}
		if len(flags.command) == 0 {
			FatalUsage(cmd, fmt.Errorf("must supply command"))
		}
		err = t.BulkEdit(opts)
		if err != nil {
			FatalUsage(cmd, err)
		}
	},
}

func init() {
	updateCmd.AddCommand(updateTicketsCmd)
	updateTicketsCmd.Flags().BoolVar(&updateTicketssCmdFlags.all, "all", false, "Bulk update all tickets (cannot be used with --query)")
	updateTicketsCmd.Flags().StringVar(&updateTicketssCmdFlags.query, "query", "", "Search query, see http://help.lighthouseapp.com/faqs/getting-started/how-do-i-search-for-tickets (required)")
	updateTicketsCmd.Flags().StringVar(&updateTicketssCmdFlags.command, "command", "", "Command keywords, see https://lighthouse.tenderapp.com/kb/ticket-workflow/how-do-i-update-tickets-with-keywords (required unless using --all)")
	updateTicketsCmd.Flags().StringVar(&updateTicketssCmdFlags.migrationToken, "migration-token", "", "If 'project' or 'account' keywords are used in --command, this must be an Lighthouse API token with access to the new project")
}
