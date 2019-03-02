package cmd

import (
	"github.com/nwidger/lighthouse/projects"
	"github.com/spf13/cobra"
)

type getProjectCmdOpts struct {
	memberships bool
}

var getProjectCmdFlags getProjectCmdOpts

// projectCmd represents the project command
var projectCmd = &cobra.Command{
	Use:   "project [id-or-name]",
	Short: "Get your Lighthouse project",
	Run: func(cmd *cobra.Command, args []string) {
		flags := getProjectCmdFlags
		p := projects.NewService(service)
		if len(args) == 0 {
			FatalUsage(cmd, "must supply project ID or name")
		}
		if flags.memberships {
			ms, err := p.Memberships(args[0])
			if err != nil {
				FatalUsage(cmd, err)
			}
			JSON(ms)
		} else {
			project, err := p.Get(args[0])
			if err != nil {
				FatalUsage(cmd, err)
			}
			JSON(project)
		}
	},
}

func init() {
	getCmd.AddCommand(projectCmd)
	projectCmd.Flags().BoolVar(&getProjectCmdFlags.memberships, "memberships", false, "List project's memberships")
}
