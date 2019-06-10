package cmd

import (
	"io"
	"os"

	"github.com/nwidger/lighthouse/users"
	"github.com/spf13/cobra"
)

type userCmdOpts struct {
	memberships bool
	avatar      bool
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
		} else if flags.avatar {
			user, err := u.Get(args[0])
			if err != nil {
				FatalUsage(cmd, err)
			}
			if len(user.AvatarURL) == 0 {
				FatalUsage(cmd, "user has no avatar")
			}
			r, _, err := u.GetAvatar(user)
			if err != nil {
				FatalUsage(cmd, err)
			}
			defer r.Close()
			io.Copy(os.Stdout, r)
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
	userCmd.Flags().BoolVar(&userCmdFlags.avatar, "avatar", false, "Download user avatar image (prints image to standard out)")
}
