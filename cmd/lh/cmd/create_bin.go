package cmd

import (
	"log"

	"github.com/nwidger/lighthouse/bins"
	"github.com/spf13/cobra"
)

type createBinsCmdOpts struct {
	defaultBin bool
	name       string
	query      string
}

var createBinsCmdFlags createBinsCmdOpts

// binCmd represents the bin command
var createBinCmd = &cobra.Command{
	Use:   "bin",
	Short: "Create a bin (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		flags := createBinsCmdFlags
		projectID := Project()
		b := bins.NewService(service, projectID)
		bin := &bins.Bin{
			Default: flags.defaultBin,
			Name:    flags.name,
			Query:   flags.query,
		}
		if len(bin.Name) == 0 {
			log.Fatal("Please specify bin name with --name")
		}
		if len(bin.Query) == 0 {
			log.Fatal("Please specify bin query with --query")
		}
		nb, err := b.Create(bin)
		if err != nil {
			log.Fatal(err)
		}
		JSON(nb)
	},
}

func init() {
	createCmd.AddCommand(createBinCmd)
	createBinCmd.Flags().BoolVar(&createBinsCmdFlags.defaultBin, "default", false, "Make bin your default filter (optional)")
	createBinCmd.Flags().StringVar(&createBinsCmdFlags.name, "name", "", "Bin name (required)")
	createBinCmd.Flags().StringVar(&createBinsCmdFlags.query, "query", "", "Bin query (required)")
}
