package cmd

import (
	"github.com/nwidger/lighthouse/profiles"
	"github.com/spf13/cobra"
)

// profileCmd represents the profile command
var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Get your Lighthouse profile",
	Run: func(cmd *cobra.Command, args []string) {
		p := profiles.NewService(service)
		u, err := p.Get()
		if err != nil {
			FatalUsage(cmd, err)
		}
		JSON(u)
	},
}

func init() {
	getCmd.AddCommand(profileCmd)
}
