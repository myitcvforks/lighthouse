package cmd

import (
	"log"

	"github.com/nwidger/lighthouse/changesets"
	"github.com/spf13/cobra"
)

// changesetCmd represents the changeset command
var changesetCmd = &cobra.Command{
	Use:   "changeset [revision]",
	Short: "Get a changeset (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		projectID := Project()
		c := changesets.NewService(service, projectID)
		if len(args) == 0 {
			log.Fatal("must supply revision")
		}
		revision := args[0]
		changeset, err := c.Get(revision)
		if err != nil {
			log.Fatal(err)
		}
		JSON(changeset)
	},
}

func init() {
	getCmd.AddCommand(changesetCmd)
}
