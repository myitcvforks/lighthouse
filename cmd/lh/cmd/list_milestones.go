package cmd

import (
	"log"

	"github.com/nwidger/lighthouse/milestones"
	"github.com/spf13/cobra"
)

type milestonesCmdOpts struct {
	page int
	all  bool
}

var milestonesCmdFlags milestonesCmdOpts

// milestonesCmd represents the milestones command
var milestonesCmd = &cobra.Command{
	Use:   "milestones",
	Short: "List milestones (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			err error
			ms  milestones.Milestones
		)
		projectID := Project()
		m := milestones.NewService(service, projectID)
		opts := &milestones.ListOptions{
			Page: milestonesCmdFlags.page,
		}
		if milestonesCmdFlags.all {
			ms, err = m.ListAll(opts)
		} else {
			ms, err = m.List(opts)
		}
		if err != nil {
			log.Fatal(err)
		}
		JSON(ms)
	},
}

func init() {
	listCmd.AddCommand(milestonesCmd)
	milestonesCmd.Flags().IntVar(&milestonesCmdFlags.page, "page", 0, "Page to return")
	milestonesCmd.Flags().BoolVar(&milestonesCmdFlags.all, "all", false, "Return all milestones")
}
