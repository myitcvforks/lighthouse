package cmd

import (
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
		flags := milestonesCmdFlags
		projectID := Project()
		m := milestones.NewService(service, projectID)
		opts := &milestones.ListOptions{
			Page: flags.page,
		}
		if flags.all {
			ms, err = m.ListAll(opts)
		} else {
			ms, err = m.List(opts)
		}
		if err != nil {
			FatalUsage(cmd, err)
		}
		JSON(ms)
	},
}

func init() {
	listCmd.AddCommand(milestonesCmd)
	milestonesCmd.Flags().IntVar(&milestonesCmdFlags.page, "page", 0, "Page to return")
	milestonesCmd.Flags().BoolVar(&milestonesCmdFlags.all, "all", false, "Return all milestones")
}
