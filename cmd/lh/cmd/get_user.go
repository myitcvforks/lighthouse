package cmd

import (
	"log"
	"strconv"

	"github.com/nwidger/lighthouse/users"
	"github.com/spf13/cobra"
)

type userCmdOpts struct {
	memberships bool
}

var userCmdFlags userCmdOpts

// userCmd represents the user command
var userCmd = &cobra.Command{
	Use:   "user [id]",
	Short: "Get information about a Lighthouse user",
	Run: func(cmd *cobra.Command, args []string) {
		flags := userCmdFlags
		u := users.NewService(service)
		if len(args) == 0 {
			log.Fatal("must supply user ID")
		}
		userID, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatal(err)
		}
		if flags.memberships {
			memberships, err := u.Memberships(userID)
			if err != nil {
				log.Fatal(err)
			}
			JSON(memberships)
		} else {
			user, err := u.Get(userID)
			if err != nil {
				log.Fatal(err)
			}
			JSON(user)
		}
	},
}

func init() {
	getCmd.AddCommand(userCmd)
	userCmd.Flags().BoolVar(&userCmdFlags.memberships, "memberships", false, "Show user's memberships")
}
