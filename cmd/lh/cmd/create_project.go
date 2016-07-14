package cmd

import (
	"log"

	"github.com/nwidger/lighthouse/projects"
	"github.com/spf13/cobra"
)

type createProjectsCmdOpts struct {
	archived bool
	name     string
	public   bool
}

var createProjectsCmdFlags createProjectsCmdOpts

// projectCmd represents the project command
var createProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Create a project",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		flags := createProjectsCmdFlags
		p := projects.NewService(service)
		project := &projects.Project{
			Archived: flags.archived,
			Name:     flags.name,
			Public:   flags.public,
		}
		if len(project.Name) == 0 {
			log.Fatal("Please specify project name with --name")
		}
		np, err := p.Create(project)
		if err != nil {
			log.Fatal(err)
		}
		JSON(np)
	},
}

func init() {
	createCmd.AddCommand(createProjectCmd)
	createProjectCmd.Flags().BoolVar(&createProjectsCmdFlags.archived, "archived", false, "Create archived project")
	createProjectCmd.Flags().StringVar(&createProjectsCmdFlags.name, "name", "", "Project name (required)")
	createProjectCmd.Flags().BoolVar(&createProjectsCmdFlags.public, "public", false, "Create public project")
}
