package cmd

import (
	"log"

	"github.com/nwidger/lighthouse/projects"
	"github.com/spf13/cobra"
)

// projectsCmd represents the projects command
var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List projects",
	Run: func(cmd *cobra.Command, args []string) {
		p := projects.NewService(service)
		ps, err := p.List()
		if err != nil {
			log.Fatal(err)
		}
		JSON(ps)
	},
}

func init() {
	listCmd.AddCommand(projectsCmd)
}
