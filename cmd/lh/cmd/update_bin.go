package cmd

import (
	"github.com/nwidger/lighthouse/bins"
	"github.com/spf13/cobra"
)

type updateBinsCmdOpts struct {
	defaultBin   bool
	noDefaultBin bool
	name         string
	query        string
}

var updateBinsCmdFlags updateBinsCmdOpts

// binCmd represents the bin command
var updateBinCmd = &cobra.Command{
	Use:   "bin [id-or-name]",
	Short: "Update a bin (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		flags := updateBinsCmdFlags
		projectID := Project()
		b := bins.NewService(service, projectID)
		if len(args) == 0 {
			FatalUsage(cmd, "must supply bin ID or name")
		}
		binID, err := BinID(args[0])
		if err != nil {
			FatalUsage(cmd, err)
		}
		bin, err := b.Get(binID)
		if err != nil {
			FatalUsage(cmd, err)
		}
		if flags.defaultBin {
			bin.Default = true
		}
		if flags.noDefaultBin {
			bin.Default = false
		}
		if len(flags.name) > 0 {
			bin.Name = flags.name
		}
		if len(flags.query) > 0 {
			bin.Query = flags.query
		}
		err = b.Update(bin)
		if err != nil {
			FatalUsage(cmd, err)
		}
		bin, err = b.Get(binID)
		if err != nil {
			FatalUsage(cmd, err)
		}
		JSON(bin)
	},
}

func init() {
	updateCmd.AddCommand(updateBinCmd)
	updateBinCmd.Flags().BoolVar(&updateBinsCmdFlags.defaultBin, "default", false, "Make bin your default filter")
	updateBinCmd.Flags().BoolVar(&updateBinsCmdFlags.noDefaultBin, "no-default", false, "Remove bin as default filter")
	updateBinCmd.Flags().StringVar(&updateBinsCmdFlags.name, "name", "", "Change bin name")
	updateBinCmd.Flags().StringVar(&updateBinsCmdFlags.query, "query", "", "Change bin query")
}
