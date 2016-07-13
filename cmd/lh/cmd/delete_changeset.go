package cmd

import (
	"log"

	"github.com/nwidger/lighthouse/changesets"
	"github.com/spf13/cobra"
)

// changesetCmd represents the changeset command
var deleteChangesetCmd = &cobra.Command{
	Use:   "changeset [revision]",
	Short: "Delete a changeset (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		projectID := Project()
		c := changesets.NewService(service, projectID)
		if len(args) == 0 {
			log.Fatal("must supply revision")
		}
		revision := args[0]
		err := c.Delete(revision)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	deleteCmd.AddCommand(deleteChangesetCmd)
}
