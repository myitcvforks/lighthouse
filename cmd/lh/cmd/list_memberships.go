package cmd

import (
	"log"

	"github.com/nwidger/lighthouse/projects"
	"github.com/spf13/cobra"
)

// membershipsCmd represents the memberships command
var membershipsCmd = &cobra.Command{
	Use:   "memberships [project-id-or-name]",
	Short: "List a project's memberships",
	Run: func(cmd *cobra.Command, args []string) {
		p := projects.NewService(service)
		if len(args) == 0 {
			log.Fatal("Please specify project ID via -p, --project, LH_PROJECT or config file")
		}
		projectID, err := ProjectID(args[0])
		if err != nil {
			log.Fatal(err)
		}
		ms, err := p.Memberships(projectID)
		if err != nil {
			log.Fatal(err)
		}
		JSON(ms)
	},
}

func init() {
	listCmd.AddCommand(membershipsCmd)
}
