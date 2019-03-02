package cmd

import (
	"github.com/nwidger/lighthouse/projects"
	"github.com/spf13/cobra"
)

// projectCmd represents the project command
var deleteProjectCmd = &cobra.Command{
	Use:   "project [id-or-name]",
	Short: "Delete a project (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		p := projects.NewService(service)
		if len(args) == 0 {
			FatalUsage(cmd, "must supply project ID or name")
		}
		err := p.Delete(args[0])
		if err != nil {
			FatalUsage(cmd, err)
		}
	},
}

func init() {
	deleteCmd.AddCommand(deleteProjectCmd)
}
