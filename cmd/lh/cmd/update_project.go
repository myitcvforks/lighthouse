package cmd

import (
	"github.com/nwidger/lighthouse/projects"
	"github.com/spf13/cobra"
)

type updateProjectsCmdOpts struct {
	archived  bool
	unarchive bool
	name      string
	public    bool
	private   bool
}

var updateProjectsCmdFlags updateProjectsCmdOpts

// projectCmd represents the project command
var updateProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Update a project",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		flags := updateProjectsCmdFlags
		p := projects.NewService(service)
		if len(args) == 0 {
			FatalUsage(cmd, "must supply project ID or name")
		}
		project, err := p.Get(args[0])
		if err != nil {
			FatalUsage(cmd, err)
		}
		if flags.archived {
			project.Archived = true
		}
		if flags.unarchive {
			project.Archived = false
		}
		if len(flags.name) > 0 {
			project.Name = flags.name
		}
		if flags.public {
			project.Public = true
		}
		if flags.private {
			project.Public = false
		}
		err = p.Update(project)
		if err != nil {
			FatalUsage(cmd, err)
		}
		project, err = p.GetByID(project.ID)
		if err != nil {
			FatalUsage(cmd, err)
		}
		JSON(project)
	},
}

func init() {
	updateCmd.AddCommand(updateProjectCmd)
	updateProjectCmd.Flags().BoolVar(&updateProjectsCmdFlags.archived, "archived", false, "Archive project")
	updateProjectCmd.Flags().BoolVar(&updateProjectsCmdFlags.unarchive, "unarchive", false, "Unarchive project")
	updateProjectCmd.Flags().StringVar(&updateProjectsCmdFlags.name, "name", "", "Change project name")
	updateProjectCmd.Flags().BoolVar(&updateProjectsCmdFlags.public, "public", false, "Make project public")
	updateProjectCmd.Flags().BoolVar(&updateProjectsCmdFlags.private, "private", false, "Make project private")
}
