package cmd

import (
	"github.com/nwidger/lighthouse/users"
	"github.com/spf13/cobra"
)

type userCmdOpts struct {
	memberships bool
}

var userCmdFlags userCmdOpts

// userCmd represents the user command
var userCmd = &cobra.Command{
	Use:   "user [id-or-name]",
	Short: "Get information about a Lighthouse user",
	Run: func(cmd *cobra.Command, args []string) {
		flags := userCmdFlags
		u := users.NewService(service)
		if len(args) == 0 {
			FatalUsage(cmd, "must supply user ID or name")
		}
		if flags.memberships {
			memberships, err := u.Memberships(args[0])
			if err != nil {
				FatalUsage(cmd, err)
			}
			JSON(memberships)
		} else {
			user, err := u.Get(args[0])
			if err != nil {
				FatalUsage(cmd, err)
			}
			JSON(user)
		}
	},
}

func init() {
	getCmd.AddCommand(userCmd)
	userCmd.Flags().BoolVar(&userCmdFlags.memberships, "memberships", false, "Show user's memberships")
}
