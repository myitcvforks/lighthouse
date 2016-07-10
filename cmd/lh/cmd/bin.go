package cmd

import (
	"log"
	"strconv"

	"github.com/nwidger/lighthouse/bins"
	"github.com/spf13/cobra"
)

// binCmd represents the bin command
var binCmd = &cobra.Command{
	Use:   "bin [id]",
	Short: "Get a ticket bin (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		projectID := Project()
		b := bins.NewService(service, projectID)
		if len(args) == 0 {
			log.Fatal("must supply bin ID")
		}
		binID, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatal(err)
		}
		bin, err := b.Get(binID)
		if err != nil {
			log.Fatal(err)
		}
		JSON(bin)
	},
}

func init() {
	getCmd.AddCommand(binCmd)
}
