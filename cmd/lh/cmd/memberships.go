package cmd

import (
	"log"

	"github.com/nwidger/lighthouse/projects"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// membershipsCmd represents the memberships command
var membershipsCmd = &cobra.Command{
	Use:   "memberships [project-id]",
	Short: "List a project's memberships",
	Run: func(cmd *cobra.Command, args []string) {
		p := projects.NewService(service)
		projectID := viper.GetInt("project")
		if len(args) == 0 && projectID == 0 {
			log.Fatal("Please specify project ID via -p, --project, LH_PROJECT or config file")
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
