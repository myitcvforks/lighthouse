package cmd

import (
	"github.com/nwidger/lighthouse/changesets"
	"github.com/spf13/cobra"
)

type changesetsCmdOpts struct {
	page int
	all  bool
}

var changesetsCmdFlags changesetsCmdOpts

// changesetsCmd represents the changesets command
var changesetsCmd = &cobra.Command{
	Use:   "changesets",
	Short: "List changesets (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			err error
			cs  changesets.Changesets
		)
		flags := changesetsCmdFlags
		projectID := Project()
		c := changesets.NewService(service, projectID)
		opts := &changesets.ListOptions{
			Page: flags.page,
		}
		if flags.all {
			cs, err = c.ListAll(opts)
		} else {
			cs, err = c.List(opts)
		}
		if err != nil {
			FatalUsage(cmd, err)
		}
		JSON(cs)
	},
}

func init() {
	listCmd.AddCommand(changesetsCmd)
	changesetsCmd.Flags().IntVar(&changesetsCmdFlags.page, "page", 0, "Page to return")
	changesetsCmd.Flags().BoolVar(&changesetsCmdFlags.all, "all", false, "Return all changesets")
}
