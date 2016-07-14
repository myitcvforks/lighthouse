package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/nwidger/lighthouse/changesets"
	"github.com/spf13/cobra"
)

type createChangesetsCmdOpts struct {
	body     string
	time     string
	changes  string
	revision string
	title    string
	user     string
}

var createChangesetsCmdFlags createChangesetsCmdOpts

// changesetCmd represents the changeset command
var createChangesetCmd = &cobra.Command{
	Use:   "changeset",
	Short: "Create a changeset (requires -p)",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		flags := createChangesetsCmdFlags
		projectID := Project()
		m := changesets.NewService(service, projectID)
		changeset := &changesets.Changeset{
			Body:     flags.body,
			Revision: flags.revision,
			Title:    flags.title,
		}
		if len(flags.changes) == 0 {
			FatalUsage(cmd, "Please specify changeset changes with --changes")
		}
		changes := changesets.Changes{}
		for _, cng := range strings.Split(flags.changes, ",") {
			cng = strings.TrimSpace(cng)
			idx := strings.Index(cng, " ")
			if idx == -1 {
				FatalUsage(cmd, fmt.Sprintf("unable to parse change %q", cng))
			}
			op, path := strings.TrimSpace(cng[:idx]), strings.TrimSpace(cng[idx:])
			changes = append(changes, &changesets.Change{
				Operation: op,
				Path:      path,
			})
		}
		changeset.Changes = changes
		if len(changeset.Revision) == 0 {
			FatalUsage(cmd, "Please specify changeset revision with --revision")
		}
		if len(changeset.Title) == 0 {
			FatalUsage(cmd, "Please specify changeset title with --title")
		}
		if len(flags.time) > 0 {
			changedAt, err := time.Parse("2006-01-02 15:04:05", flags.time)
			if err != nil {
				FatalUsage(cmd, err)
			}
			changeset.ChangedAt = &changedAt
		}
		if len(flags.user) > 0 {
			userID, err := UserID(flags.user)
			if err != nil {
				FatalUsage(cmd, err)
			}
			changeset.UserID = userID
		}
		nm, err := m.Create(changeset)
		if err != nil {
			FatalUsage(cmd, err)
		}
		JSON(nm)
	},
}

func init() {
	createCmd.AddCommand(createChangesetCmd)
	createChangesetCmd.Flags().StringVar(&createChangesetsCmdFlags.body, "body", "", "Changeset body (optional)")
	createChangesetCmd.Flags().StringVar(&createChangesetsCmdFlags.time, "time", "", "Changeset 24-hour timestamp YYYY-MM-DD HH:mm:ss (optional)")
	createChangesetCmd.Flags().StringVar(&createChangesetsCmdFlags.changes, "changes", "", "Comma-separated changes 'OP PATH, OP PATH, OP PATH' (required)")
	createChangesetCmd.Flags().StringVar(&createChangesetsCmdFlags.revision, "revision", "", "Changeset revision (required)")
	createChangesetCmd.Flags().StringVar(&createChangesetsCmdFlags.title, "title", "", "Changeset title (required)")
	createChangesetCmd.Flags().StringVar(&createChangesetsCmdFlags.user, "user", "", "Assign changeset to user (optional)")
}
