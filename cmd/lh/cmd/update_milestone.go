package cmd

import (
	"time"

	"github.com/nwidger/lighthouse/milestones"
	"github.com/spf13/cobra"
)

type updateMilestonesCmdOpts struct {
	goals string
	title string
	due   string
	close bool
	open  bool
}

var updateMilestonesCmdFlags updateMilestonesCmdOpts

// milestoneCmd represents the milestone command
var updateMilestoneCmd = &cobra.Command{
	Use:   "milestone [id-or-title]",
	Short: "Update a milestone (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		flags := updateMilestonesCmdFlags
		projectID := Project()
		m := milestones.NewService(service, projectID)
		if len(args) == 0 {
			FatalUsage(cmd, "must supply milestone ID or title")
		}
		milestoneID, err := MilestoneID(args[0])
		if err != nil {
			FatalUsage(cmd, err)
		}
		milestone, err := m.Get(milestoneID)
		if err != nil {
			FatalUsage(cmd, err)
		}
		if len(flags.goals) > 0 {
			milestone.Goals = flags.goals
		}
		if len(flags.title) > 0 {
			milestone.Title = flags.title
		}
		if len(flags.due) > 0 {
			due, err := time.Parse("2006-01-02", flags.due)
			if err != nil {
				FatalUsage(cmd, err)
			}
			milestone.DueOn = &due
		}
		err = m.Update(milestone)
		if err != nil {
			FatalUsage(cmd, err)
		}
		if flags.close {
			err = m.Close(milestoneID)
			if err != nil {
				FatalUsage(cmd, err)
			}
		}
		if flags.open {
			err = m.Open(milestoneID)
			if err != nil {
				FatalUsage(cmd, err)
			}
		}
		milestone, err = m.Get(milestoneID)
		if err != nil {
			FatalUsage(cmd, err)
		}
		JSON(milestone)
	},
}

func init() {
	updateCmd.AddCommand(updateMilestoneCmd)
	updateMilestoneCmd.Flags().StringVar(&updateMilestonesCmdFlags.goals, "goals", "", "Change milestone goals")
	updateMilestoneCmd.Flags().StringVar(&updateMilestonesCmdFlags.title, "title", "", "Change milestone title")
	updateMilestoneCmd.Flags().StringVar(&updateMilestonesCmdFlags.due, "due", "", "Change milestone due date YYYY-MM-DD")
	updateMilestoneCmd.Flags().BoolVar(&updateMilestonesCmdFlags.close, "close", false, "Close milestone")
	updateMilestoneCmd.Flags().BoolVar(&updateMilestonesCmdFlags.open, "open", false, "Open milestone")
}
