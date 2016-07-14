package cmd

import (
	"log"

	"github.com/nwidger/lighthouse/milestones"
	"github.com/spf13/cobra"
)

// milestoneCmd represents the milestone command
var deleteMilestoneCmd = &cobra.Command{
	Use:   "milestone [id-or-title]",
	Short: "Delete a milestone (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		projectID := Project()
		m := milestones.NewService(service, projectID)
		if len(args) == 0 {
			log.Fatal("must supply milestone ID or title")
		}
		milestoneID, err := MilestoneID(args[0])
		if err != nil {
			log.Fatal(err)
		}
		err = m.Delete(milestoneID)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	deleteCmd.AddCommand(deleteMilestoneCmd)
}
