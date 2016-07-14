package cmd

import (
	"log"

	"github.com/nwidger/lighthouse/projects"
	"github.com/spf13/cobra"
)

// projectCmd represents the project command
var projectCmd = &cobra.Command{
	Use:   "project [id]",
	Short: "Get your Lighthouse project",
	Run: func(cmd *cobra.Command, args []string) {
		p := projects.NewService(service)
		if len(args) == 0 {
			log.Fatal("must supply project ID")
		}
		projectID, err := ProjectID(args[0])
		if err != nil {
			log.Fatal(err)
		}
		project, err := p.Get(projectID)
		if err != nil {
			log.Fatal(err)
		}
		JSON(project)
	},
}

func init() {
	getCmd.AddCommand(projectCmd)
}
