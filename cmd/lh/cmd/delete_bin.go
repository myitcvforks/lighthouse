package cmd

import (
	"github.com/nwidger/lighthouse/bins"
	"github.com/spf13/cobra"
)

// binCmd represents the bin command
var deleteBinCmd = &cobra.Command{
	Use:   "bin [id-or-name]",
	Short: "Delete a bin (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		projectID := Project()
		b := bins.NewService(service, projectID)
		if len(args) == 0 {
			FatalUsage(cmd, "must supply bin ID or name")
		}
		err := b.Delete(args[0])
		if err != nil {
			FatalUsage(cmd, err)
		}
	},
}

func init() {
	deleteCmd.AddCommand(deleteBinCmd)
}
