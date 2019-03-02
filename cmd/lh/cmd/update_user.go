package cmd

import (
	"github.com/nwidger/lighthouse/users"
	"github.com/spf13/cobra"
)

type updateUserCmdOpts struct {
	job     string
	name    string
	website string
}

var updateUserCmdFlags updateUserCmdOpts

// updateUserCmd represents the user command
var updateUserCmd = &cobra.Command{
	Use:   "user [id-or-name]",
	Short: "Update information about a Lighthouse user",
	Run: func(cmd *cobra.Command, args []string) {
		flags := updateUserCmdFlags
		u := users.NewService(service)
		if len(args) == 0 {
			FatalUsage(cmd, "must supply user ID or name")
		}
		user, err := u.Get(args[0])
		if err != nil {
			FatalUsage(cmd, err)
		}
		if len(flags.job) > 0 {
			user.Job = flags.job
		}
		if len(flags.name) > 0 {
			user.Name = flags.name
		}
		if len(flags.website) > 0 {
			user.Website = flags.website
		}
		err = u.Update(user)
		if err != nil {
			FatalUsage(cmd, err)
		}
		user, err = u.GetByID(user.ID)
		if err != nil {
			FatalUsage(cmd, err)
		}
		JSON(user)
	},
}

func init() {
	updateCmd.AddCommand(updateUserCmd)
	updateUserCmd.Flags().StringVar(&updateUserCmdFlags.job, "job", "", "User job")
	updateUserCmd.Flags().StringVar(&updateUserCmdFlags.name, "name", "", "User name")
	updateUserCmd.Flags().StringVar(&updateUserCmdFlags.website, "website", "", "User website")
}
