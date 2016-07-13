package cmd

import (
	"log"

	"github.com/nwidger/lighthouse/bins"
	"github.com/spf13/cobra"
)

// binsCmd represents the bins command
var binsCmd = &cobra.Command{
	Use:   "bins",
	Short: "List ticket bins (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		projectID := Project()
		b := bins.NewService(service, projectID)
		bs, err := b.List()
		if err != nil {
			log.Fatal(err)
		}
		JSON(bs)
	},
}

func init() {
	listCmd.AddCommand(binsCmd)
}
