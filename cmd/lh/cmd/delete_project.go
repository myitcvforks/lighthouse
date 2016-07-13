package cmd

import (
	"log"
	"strconv"

	"github.com/nwidger/lighthouse/projects"
	"github.com/spf13/cobra"
)

// projectCmd represents the project command
var deleteProjectCmd = &cobra.Command{
	Use:   "project [id]",
	Short: "Delete a project (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		p := projects.NewService(service)
		if len(args) == 0 {
			log.Fatal("must supply project ID")
		}
		projectID, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatal(err)
		}
		err = p.Delete(projectID)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	deleteCmd.AddCommand(deleteProjectCmd)
}
