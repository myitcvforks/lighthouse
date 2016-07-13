package cmd

import (
	"log"
	"strconv"

	"github.com/nwidger/lighthouse/bins"
	"github.com/spf13/cobra"
)

// binCmd represents the bin command
var deleteBinCmd = &cobra.Command{
	Use:   "bin [id]",
	Short: "Delete a bin (requires -p)",
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
		err = b.Delete(binID)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	deleteCmd.AddCommand(deleteBinCmd)
}
