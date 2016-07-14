package cmd

import (
	"log"
	"time"

	"github.com/nwidger/lighthouse/milestones"
	"github.com/spf13/cobra"
)

type createMilestonesCmdOpts struct {
	goals string
	title string
	due   string
}

var createMilestonesCmdFlags createMilestonesCmdOpts

// milestoneCmd represents the milestone command
var createMilestoneCmd = &cobra.Command{
	Use:   "milestone",
	Short: "Create a milestone (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		flags := createMilestonesCmdFlags
		projectID := Project()
		m := milestones.NewService(service, projectID)
		milestone := &milestones.Milestone{
			Goals: flags.goals,
			Title: flags.title,
		}
		if len(milestone.Title) == 0 {
			log.Fatal("Please specify milestone title with --title")
		}
		if len(flags.due) > 0 {
			due, err := time.Parse("2006-01-02", flags.due)
			if err != nil {
				log.Fatal(err)
			}
			milestone.DueOn = &due
		}
		nm, err := m.Create(milestone)
		if err != nil {
			log.Fatal(err)
		}
		JSON(nm)
	},
}

func init() {
	createCmd.AddCommand(createMilestoneCmd)
	createMilestoneCmd.Flags().StringVar(&createMilestonesCmdFlags.goals, "goals", "", "Milestone goals (optional)")
	createMilestoneCmd.Flags().StringVar(&createMilestonesCmdFlags.title, "title", "", "Milestone title (required)")
	createMilestoneCmd.Flags().StringVar(&createMilestonesCmdFlags.due, "due", "", "Milestone due date YYYY-MM-DD (optional)")
}
