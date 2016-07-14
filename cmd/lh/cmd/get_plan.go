package cmd

import "github.com/spf13/cobra"

// planCmd represents the plan command
var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Get your Lighthouse plan",
	Run: func(cmd *cobra.Command, args []string) {
		p, err := service.Plan()
		if err != nil {
			FatalUsage(cmd, err)
		}
		JSON(p)
	},
}

func init() {
	getCmd.AddCommand(planCmd)
}
