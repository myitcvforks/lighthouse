package cmd

import (
	"github.com/nwidger/lighthouse/bins"
	"github.com/spf13/cobra"
)

// binCmd represents the bin command
var binCmd = &cobra.Command{
	Use:   "bin [id-or-name]",
	Short: "Get a ticket bin (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		projectID := Project()
		b := bins.NewService(service, projectID)
		if len(args) == 0 {
			FatalUsage(cmd, "must supply bin ID or name")
		}
		bin, err := b.Get(args[0])
		if err != nil {
			FatalUsage(cmd, err)
		}
		JSON(bin)
	},
}

func init() {
	getCmd.AddCommand(binCmd)
}
