package main

import (
	"bytes"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mholt/archiver"
	"github.com/nwidger/lighthouse"
	"github.com/nwidger/lighthouse/bins"
	"github.com/nwidger/lighthouse/changesets"
	"github.com/nwidger/lighthouse/messages"
	"github.com/nwidger/lighthouse/milestones"
	"github.com/nwidger/lighthouse/profiles"
	"github.com/nwidger/lighthouse/projects"
	"github.com/nwidger/lighthouse/tickets"
	"github.com/nwidger/lighthouse/users"
	gitlab "github.com/xanzy/go-gitlab"
)

var (
	usersMap     = map[int]*gitlab.User{}
	usersNameMap = map[string]*gitlab.User{}

	projectsMap   = map[int]*gitlab.Project{}
	milestonesMap = map[int]*gitlab.Milestone{}
	issuesMap     = map[int]*gitlab.Issue{}
)

func main() {
	export := ""
	token := ""
	baseURL := ""
	usersPath := ""
	password := "changeme"
	project := ""
	milestone := ""
	number := 0
	delete := false
	insecure := false

	flag.StringVar(&token, "token", token, "GitLab API token to use")
	flag.StringVar(&baseURL, "base-url", baseURL, "GitLab base URL to use (i.e., https://gitlab.example.com/)")
	flag.StringVar(&usersPath, "users", usersPath, "Path to JSON file mapping Lighthouse user ID's to GitLab users")
	flag.StringVar(&password, "password", password, "Password to use when creating GitLab users")
	flag.StringVar(&project, "project", project, "Only migrate projects with the given name (useful for testing)")
	flag.StringVar(&milestone, "milestone", milestone, "Only migrate milestones with the given title (useful for testing)")
	flag.IntVar(&number, "number", number, "Only migrate tickets with the given number (useful for testing)")
	flag.BoolVar(&delete, "delete", delete, "Delete all GitLab projects and users (except user owning API token -token) before importing")
	flag.BoolVar(&insecure, "insecure", insecure, "Allow insecure HTTPS connections to GitLab API")

	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Fprintf(os.Stderr, "Must specify path to Lighthouse export file\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if len(baseURL) == 0 {
		fmt.Fprintf(os.Stderr, "Must specify GitLab base URL via -base-url\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if len(usersPath) == 0 {
		fmt.Fprintf(os.Stderr, "Must specify path to Lighthouse users map file via -users\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if len(token) == 0 {
		fmt.Fprintf(os.Stderr, "Must specify GitLab API token via -token\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if len(password) == 0 {
		fmt.Fprintf(os.Stderr, "Must specify password for creating GitLab users via -password\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	export = flag.Arg(0)

	exp, err := readLHExport(export)
	if err != nil {
		log.Fatal(err)
	}

	var client *http.Client
	if insecure {
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	}

	git := gitlab.NewClient(client, token)
	err = git.SetBaseURL(baseURL)
	if err != nil {
		log.Fatal(err)
	}

	me, _, err := git.Users.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}

	if delete {
		us, _, err := git.Users.ListUsers(&gitlab.ListUsersOptions{})
		if err != nil {
			log.Fatal(err)
		}
		for _, u := range us {
			if u.Username == me.Username {
				continue
			}
			git.Users.DeleteUser(u.ID)
			if err != nil {
				log.Fatal(err)
			}
		}

		ps, _, err := git.Projects.ListProjects(&gitlab.ListProjectsOptions{})
		if err != nil {
			log.Fatal(err)
		}
		for _, p := range ps {
			_, err = git.Projects.DeleteProject(p.ID)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	f, err := os.Open(usersPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	err = dec.Decode(&usersMap)
	if err != nil {
		log.Fatal(err)
	}

	for _, lhUser := range exp.users.list {
		userOpt, options, ok := lhUserToCreateUser(lhUser, password)
		if !ok {
			continue
		}
		fmt.Println("creating user", *userOpt.Username)
		u, _, err := git.Users.CreateUser(userOpt, options...)
		if err != nil {
			fmt.Fprintln(os.Stderr, "unable to create user", lhUser.Name, err)
			continue
		}
		usersMap[lhUser.ID] = u
		usersNameMap[lhUser.Name] = u
	}

	us, _, err := git.Users.ListUsers(&gitlab.ListUsersOptions{})
	for _, u := range us {
		for _, lhUser := range exp.users.list {
			if u.Name == lhUser.Name {
				usersMap[lhUser.ID] = u
				usersNameMap[lhUser.Name] = u
				break
			}
		}
	}

	for _, lhProject := range exp.projects.list {
		if len(project) > 0 && !strings.EqualFold(lhProject.Name, project) {
			continue
		}
		projectOpt, options, ok := lhProjectToCreateProject(lhProject)
		if !ok {
			continue
		}
		fmt.Println("creating project", *projectOpt.Name)
		p, _, err := git.Projects.CreateProject(projectOpt, options...)
		if err != nil {
			fmt.Fprintln(os.Stderr, "unable to create project", lhProject.Name, err)
			continue
		}
		projectsMap[lhProject.ID] = p

		for _, lhMembership := range lhProject.memberships {
			memberOpt, options, ok := lhMembershipToAddProjectMember(lhMembership)
			if !ok {
				continue
			}
			_, _, err = git.ProjectMembers.AddProjectMember(p.ID, memberOpt, options...)
			if err != nil {
				fmt.Fprintln(os.Stderr, "unable to add", lhMembership.User.Name, "to project", lhProject.Name, err)
			}
		}

		for _, lhMilestone := range lhProject.milestones.list {
			if len(milestone) > 0 && !strings.EqualFold(lhMilestone.Title, milestone) {
				continue
			}
			createMilestoneOpt, options, ok := lhMilestoneToCreateMilestone(lhMilestone)
			if !ok {
				continue
			}
			fmt.Println("creating milestone", *createMilestoneOpt.Title)
			m, _, err := git.Milestones.CreateMilestone(p.ID, createMilestoneOpt, options...)
			if err != nil {
				fmt.Fprintln(os.Stderr, "unable to create milestone", lhMilestone.Title, "to project", lhProject.Name, err)
				continue
			}
			milestonesMap[lhMilestone.ID] = m

			updateMilestoneOpt, options, ok := lhMilestoneToUpdateMilestone(lhMilestone)
			if ok {
				_, _, err = git.Milestones.UpdateMilestone(p.ID, m.ID, updateMilestoneOpt, options...)
				if err != nil {
					fmt.Fprintln(os.Stderr, "unable to update milestone", lhMilestone.Title, "in project", lhProject.Name, err)
				}
			}
		}

		for _, lhTicket := range lhProject.tickets.list {
			if number > 0 && lhTicket.Number != number {
				continue
			}
			issueOpt, options, ok := lhTicketToCreateIssue(lhTicket)
			if !ok {
				continue
			}
			fmt.Println("creating issue", *issueOpt.IID)
			i, _, err := git.Issues.CreateIssue(p.ID, issueOpt, options...)
			if err != nil {
				fmt.Fprintln(os.Stderr, "unable to create issue", lhTicket.Number, "in project", lhProject.Name, err)
				continue
			}
			issuesMap[lhTicket.Number] = i

			for _, watcherID := range lhTicket.WatchersIDs {
				options := withSudoByUserID(watcherID)
				_, _, err = git.Issues.SubscribeToIssue(p.ID, i.IID, options...)
				if err != nil && err != io.EOF {
					fmt.Fprintln(os.Stderr, "unable to subscribe user", watcherID, "to issue", i.IID, "in project", lhProject.Name, err)
				}
			}

			for _, lhVersion := range lhTicket.Versions {
				issueOpt, options, ok := lhTicketVersionToUpdateIssue(lhVersion)
				if ok {
					_, _, err = git.Issues.UpdateIssue(p.ID, i.IID, issueOpt, options...)
					if err != nil {
						fmt.Fprintln(os.Stderr, "unable to update issue", i.IID, "in project", lhProject.Name, err)
					}
				}
				noteOpt, options, ok := lhTicketVersionToCreateIssueNote(lhVersion, lhVersion.CreatedAt.Equal(*lhTicket.CreatedAt))
				if ok {
					_, _, err = git.Notes.CreateIssueNote(p.ID, i.IID, noteOpt, options...)
					if err != nil {
						fmt.Fprintln(os.Stderr, "unable to create issue note for issue", i.IID, "in project", lhProject.Name, err)
					}
				}
			}

			for _, lhAttachment := range lhTicket.attachments.list {
				dir, file, options, ok := lhAttachmentToUploadFile(lhAttachment)
				if !ok {
					continue
				}
				pf, _, err := git.Projects.UploadFile(p.ID, file, options...)
				os.RemoveAll(dir)
				if err != nil {
					fmt.Fprintln(os.Stderr, "unable to upload file", file, "for issue", i.IID, "in project", lhProject.Name, err)
					continue
				}
				noteOpt, options, ok := lhAttachmentToCreateIssueNote(lhAttachment, pf)
				if !ok {
					continue
				}
				_, _, err = git.Notes.CreateIssueNote(p.ID, i.IID, noteOpt, options...)
				if err != nil {
					fmt.Fprintln(os.Stderr, "unable to create attachment issue note for issue", i.IID, "in project", lhProject.Name, err)
				}
			}
		}
	}
}

func projectByID(id int) (*gitlab.Project, bool) {
	if id == 0 {
		return nil, false
	}
	p, ok := projectsMap[id]
	if !ok || p == nil {
		return nil, false
	}
	return p, true
}

func milestoneByID(id int) (*gitlab.Milestone, bool) {
	if id == 0 {
		return nil, false
	}
	m, ok := milestonesMap[id]
	if !ok || m == nil {
		return nil, false
	}
	return m, true
}

func issueByNumber(number int) (*gitlab.Issue, bool) {
	if number == 0 {
		return nil, false
	}
	i, ok := issuesMap[number]
	if !ok || i == nil {
		return nil, false
	}
	return i, true
}

func userByID(id int) (*gitlab.User, bool) {
	if id == 0 {
		return nil, false
	}
	u, ok := usersMap[id]
	if !ok || u == nil {
		return nil, false
	}
	return u, true
}

func userByUsername(username string) (*gitlab.User, bool) {
	if len(username) == 0 {
		return nil, false
	}
	u, ok := usersNameMap[username]
	if !ok || u == nil {
		return nil, false
	}
	return u, true
}

func withSudoByUserID(id int) []gitlab.OptionFunc {
	var options []gitlab.OptionFunc
	u, ok := userByID(id)
	if ok {
		options = append(options, gitlab.WithSudo(u.ID))
	}
	return options
}

func withSudoByUsername(username string) []gitlab.OptionFunc {
	var options []gitlab.OptionFunc
	u, ok := userByUsername(username)
	if ok {
		options = append(options, gitlab.WithSudo(u.ID))
	}
	return options
}

func lhUserToCreateUser(lhUser *lhUser, password string) (*gitlab.CreateUserOptions, []gitlab.OptionFunc, bool) {
	var options []gitlab.OptionFunc
	u, ok := userByID(lhUser.ID)
	if !ok {
		return nil, nil, false
	}
	opt := &gitlab.CreateUserOptions{
		Email:            gitlab.String(u.Email),
		Password:         gitlab.String(password),
		Username:         gitlab.String(u.Username),
		Name:             gitlab.String(u.Name),
		Admin:            gitlab.Bool(true),
		CanCreateGroup:   gitlab.Bool(true),
		SkipConfirmation: gitlab.Bool(true),
	}
	return opt, options, true
}

func lhProjectToCreateProject(lhProject *lhProject) (*gitlab.CreateProjectOptions, []gitlab.OptionFunc, bool) {
	var options []gitlab.OptionFunc
	opt := &gitlab.CreateProjectOptions{
		Name:        gitlab.String(strings.ReplaceAll(lhProject.Name, `'`, ``)),
		Description: gitlab.String(lhtoGitLabMarkdown(lhProject.Description)),
	}
	return opt, options, true
}

func lhMembershipToAddProjectMember(lhMembership *projects.Membership) (*gitlab.AddProjectMemberOptions, []gitlab.OptionFunc, bool) {
	var options []gitlab.OptionFunc
	u, ok := userByID(lhMembership.UserID)
	if !ok {
		return nil, nil, false
	}
	opt := &gitlab.AddProjectMemberOptions{
		UserID:      gitlab.Int(u.ID),
		AccessLevel: gitlab.AccessLevel(gitlab.MaintainerPermissions),
	}
	return opt, options, true
}

func lhMilestoneToCreateMilestone(lhMilestone *milestones.Milestone) (*gitlab.CreateMilestoneOptions, []gitlab.OptionFunc, bool) {
	options := withSudoByUsername(lhMilestone.UserName)
	var startDate, dueDate *gitlab.ISOTime
	if lhMilestone.CreatedAt != nil {
		d := gitlab.ISOTime(*lhMilestone.CreatedAt)
		startDate = &d
	}
	if lhMilestone.DueOn != nil &&
		(lhMilestone.CreatedAt == nil || lhMilestone.DueOn.After(*lhMilestone.CreatedAt)) {
		d := gitlab.ISOTime(*lhMilestone.DueOn)
		dueDate = &d
	}
	opt := &gitlab.CreateMilestoneOptions{
		Title:       gitlab.String(lhMilestone.Title),
		Description: gitlab.String(lhtoGitLabMarkdown(lhMilestone.Goals)),
		StartDate:   startDate,
		DueDate:     dueDate,
	}
	return opt, options, true
}

func lhMilestoneToUpdateMilestone(lhMilestone *milestones.Milestone) (*gitlab.UpdateMilestoneOptions, []gitlab.OptionFunc, bool) {
	options := withSudoByUsername(lhMilestone.UserName)
	var stateEvent *string
	if lhMilestone.CompletedAt != nil {
		stateEvent = gitlab.String("close")
	} else {
		stateEvent = gitlab.String("activate")
	}
	if stateEvent == nil {
		return nil, nil, false
	}
	opt := &gitlab.UpdateMilestoneOptions{
		StateEvent: stateEvent,
	}
	return opt, options, true
}

func lhTicketToCreateIssue(lhTicket *lhTicket) (*gitlab.CreateIssueOptions, []gitlab.OptionFunc, bool) {
	options := withSudoByUserID(lhTicket.CreatorID)

	var title *string
	title = gitlab.String(lhTicket.Title)
	var description *string
	description = gitlab.String(lhtoGitLabMarkdown(lhTicket.Body))
	var assigneeIDs []int
	if lhTicket.AssignedUserID == 0 {
		assigneeIDs = append(assigneeIDs, 0)
	} else {
		u, ok := userByID(lhTicket.AssignedUserID)
		if ok {
			assigneeIDs = append(assigneeIDs, u.ID)
		}
	}
	var milestoneID *int
	if lhTicket.MilestoneID == 0 {
		milestoneID = gitlab.Int(0)
	} else {
		m, ok := milestoneByID(lhTicket.MilestoneID)
		if ok {
			milestoneID = gitlab.Int(m.ID)
		}
	}
	var labels gitlab.Labels
	labels = lhTicketToLabels(lhTicket)
	var createdAt *time.Time
	if lhTicket.CreatedAt != nil {
		createdAt = lhTicket.CreatedAt
	}

	if len(lhTicket.Versions) > 0 {
		lhVersion := lhTicket.Versions[0]
		updateOpt, _, ok := lhTicketVersionToUpdateIssue(lhVersion)
		if ok {
			assigneeIDs = updateOpt.AssigneeIDs
			milestoneID = updateOpt.MilestoneID
			labels = updateOpt.Labels
		}
	}

	opt := &gitlab.CreateIssueOptions{
		IID:         gitlab.Int(lhTicket.Number),
		Title:       title,
		Description: description,
		AssigneeIDs: assigneeIDs,
		MilestoneID: milestoneID,
		Labels:      labels,
		CreatedAt:   createdAt,
	}
	return opt, options, true
}

func lhTicketVersionToUpdateIssue(lhVersion *tickets.TicketVersion) (*gitlab.UpdateIssueOptions, []gitlab.OptionFunc, bool) {
	options := withSudoByUserID(lhVersion.UserID)
	var title *string
	title = gitlab.String(lhVersion.Title)
	var assigneeIDs []int
	if lhVersion.AssignedUserID == 0 {
		assigneeIDs = append(assigneeIDs, 0)
	} else {
		u, ok := userByID(lhVersion.AssignedUserID)
		if ok {
			assigneeIDs = append(assigneeIDs, u.ID)
		}
	}
	var milestoneID *int
	if lhVersion.MilestoneID == 0 {
		milestoneID = gitlab.Int(0)
	} else {
		m, ok := milestoneByID(lhVersion.MilestoneID)
		if ok {
			milestoneID = gitlab.Int(m.ID)
		}
	}
	labels := lhTicketVersionToLabels(lhVersion)
	var stateEvent *string
	if lhVersion.Closed {
		stateEvent = gitlab.String("close")
	} else {
		stateEvent = gitlab.String("reopen")
	}
	var updatedAt *time.Time
	if lhVersion.UpdatedAt != nil {
		updatedAt = lhVersion.UpdatedAt
	}
	opt := &gitlab.UpdateIssueOptions{
		Title:       title,
		AssigneeIDs: assigneeIDs,
		MilestoneID: milestoneID,
		Labels:      labels,
		StateEvent:  stateEvent,
		UpdatedAt:   updatedAt,
	}
	return opt, options, true
}

func lhTicketVersionToCreateIssueNote(lhVersion *tickets.TicketVersion, currentVersion bool) (*gitlab.CreateIssueNoteOptions, []gitlab.OptionFunc, bool) {
	options := withSudoByUserID(lhVersion.UserID)
	var createdAt *time.Time
	if lhVersion.CreatedAt != nil {
		createdAt = lhVersion.CreatedAt
	}
	var body string
	if len(lhVersion.DiffableAttributes.State) > 0 {
		body += fmt.Sprintf("**State changed from `\"%s\"` to `\"%s\"`**\n\n",
			lhVersion.DiffableAttributes.State, lhVersion.State)
	}
	if !currentVersion {
		body += lhtoGitLabMarkdown(lhVersion.Body)
	}
	if len(strings.TrimSpace(body)) == 0 {
		return nil, nil, false
	}
	opt := &gitlab.CreateIssueNoteOptions{
		Body:      gitlab.String(body),
		CreatedAt: createdAt,
	}
	return opt, options, true
}

func lhAttachmentToUploadFile(lhAttachment *lhAttachment) (dir, file string, options []gitlab.OptionFunc, ok bool) {
	var err error
	options = withSudoByUserID(lhAttachment.UploaderID)
	dir, err = ioutil.TempDir("", "lhtogitlab-ticket-attachment")
	if err != nil {
		return "", "", nil, false
	}
	defer func() {
		if !ok && len(dir) > 0 {
			os.RemoveAll(dir)
		}
	}()
	file = filepath.Join(dir, lhAttachment.Filename)
	f, err := os.Create(file)
	if err != nil {
		return "", "", nil, false
	}
	defer f.Close()
	io.Copy(f, lhAttachment.r)
	return dir, file, options, true
}

func lhAttachmentToCreateIssueNote(lhAttachment *lhAttachment, pf *gitlab.ProjectFile) (*gitlab.CreateIssueNoteOptions, []gitlab.OptionFunc, bool) {
	options := withSudoByUserID(lhAttachment.UploaderID)
	var createdAt *time.Time
	if lhAttachment.CreatedAt != nil {
		createdAt = lhAttachment.CreatedAt
	}
	body := pf.Markdown
	if len(body) == 0 {
		return nil, nil, false
	}
	opt := &gitlab.CreateIssueNoteOptions{
		Body:      gitlab.String(body),
		CreatedAt: createdAt,
	}
	return opt, options, true
}

func lhTicketToLabels(lhTicket *lhTicket) gitlab.Labels {
	var labels gitlab.Labels
	for _, tag := range lhTicket.Tags {
		labels = append(labels, tag.Tag.Name)
	}
	return labels
}

func lhTicketVersionToLabels(lhVersion *tickets.TicketVersion) gitlab.Labels {
	var labels gitlab.Labels
	r := strings.NewReader(lhVersion.Tag)
	cr := csv.NewReader(r)
	cr.Comma = ' '
	record, err := cr.Read()
	if err != nil {
		record = strings.Fields(lhVersion.Tag)
	}
	for _, r := range record {
		if len(r) == 0 {
			continue
		}
		labels = append(labels, r)
	}
	return labels
}

func lhtoGitLabMarkdown(text string) string {
	if len(strings.TrimSpace(text)) == 0 {
		return text
	}
	converted := strings.ReplaceAll(text, `@@@`, "```")
	if converted == text {
		return text
	}
	return fmt.Sprintf(`
%s

<details>
<summary>Original Lighthouse text</summary>

%s
%s
%s
</details>
`, converted, "```", text, "```")
}

type lhExport struct {
	plan     *lighthouse.Plan
	profile  *profiles.User
	projects *lhProjects
	users    *lhUsers
}

type lhProjects struct {
	byID map[int]*lhProject
	list []*lhProject
}

type lhProject struct {
	*projects.Project

	bins        lhBins
	changesets  lhChangesets
	memberships projects.Memberships
	messages    messages.Messages
	milestones  lhMilestones
	tickets     lhTickets
}

type lhBins struct {
	byID map[int]*bins.Bin
	list []*bins.Bin
}

type lhChangesets struct {
	byRevision map[string]*changesets.Changeset
	list       []*changesets.Changeset
}

type lhChangeset struct {
	*changesets.Changeset
}

type lhMilestones struct {
	byID map[int]*milestones.Milestone
	list []*milestones.Milestone
}

type lhTickets struct {
	byNumber map[int]*lhTicket
	list     []*lhTicket
}

type lhTicket struct {
	*tickets.Ticket

	attachments lhAttachments
}

type lhUsers struct {
	byID map[int]*lhUser
	list []*lhUser
}

type lhUser struct {
	*users.User

	avatar      *lhFile
	memberships users.Memberships
}

type lhAttachments struct {
	byFilename map[string]*lhAttachment
	list       []*lhAttachment
}

type lhAttachment struct {
	*tickets.Attachment

	r io.Reader
}

type lhFile struct {
	filename string
	r        io.Reader
}

func readLHExport(path string) (*lhExport, error) {
	tempDir, err := ioutil.TempDir("", "lhtogitlab")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer signal.Reset(os.Interrupt)

	go func(c chan os.Signal) {
		<-c
		if len(tempDir) > 0 {
			os.RemoveAll(tempDir)
		}
	}(c)

	e := &lhExport{
		projects: &lhProjects{
			byID: map[int]*lhProject{},
			list: []*lhProject{},
		},
		users: &lhUsers{
			byID: map[int]*lhUser{},
			list: []*lhUser{},
		},
	}

	tgz := archiver.NewTarGz()
	tgz.Tar.OverwriteExisting = true

	err = tgz.Unarchive(path, tempDir)
	if err != nil {
		return nil, err
	}

	userDirs, err := filepath.Glob(filepath.Join(tempDir, "*", "users", "*"))
	if err != nil {
		return nil, err
	}

	for _, dir := range userDirs {
		uf, err := os.Open(filepath.Join(dir, "user.json"))
		if err != nil {
			return nil, err
		}
		defer uf.Close()
		dec := json.NewDecoder(uf)
		u := &lhUser{
			User:        &users.User{},
			memberships: users.Memberships{},
		}
		err = dec.Decode(u.User)
		if err != nil {
			return nil, err
		}
		uf.Close()
		mf, err := os.Open(filepath.Join(dir, "memberships.json"))
		if err == nil {
			defer mf.Close()
			dec = json.NewDecoder(mf)
			err = dec.Decode(&u.memberships)
			if err != nil {
				return nil, err
			}
			mf.Close()
		}
		avatarPaths, err := filepath.Glob(filepath.Join(dir, "avatar.*"))
		if err != nil {
			return nil, err
		}
		if len(avatarPaths) != 0 {
			u.avatar = &lhFile{
				filename: filepath.Base(avatarPaths[0]),
			}
			buf, err := ioutil.ReadFile(avatarPaths[0])
			if err != nil {
				return nil, err
			}
			u.avatar.r = bytes.NewReader(buf)
		}
		e.users.byID[u.ID] = u
		e.users.list = append(e.users.list, u)
	}
	sort.Slice(e.users.list, func(i, j int) bool { return e.users.list[i].ID < e.users.list[j].ID })

	projectDirs, err := filepath.Glob(filepath.Join(tempDir, "*", "projects", "*"))
	if err != nil {
		return nil, err
	}

	for _, dir := range projectDirs {
		pf, err := os.Open(filepath.Join(dir, "project.json"))
		if err != nil {
			return nil, err
		}
		defer pf.Close()
		dec := json.NewDecoder(pf)
		p := &lhProject{
			Project: &projects.Project{},
			bins: lhBins{
				byID: map[int]*bins.Bin{},
				list: []*bins.Bin{},
			},
			changesets: lhChangesets{
				byRevision: map[string]*changesets.Changeset{},
				list:       []*changesets.Changeset{},
			},
			memberships: projects.Memberships{},
			messages:    messages.Messages{},
			milestones: lhMilestones{
				byID: map[int]*milestones.Milestone{},
				list: []*milestones.Milestone{},
			},
			tickets: lhTickets{
				byNumber: map[int]*lhTicket{},
				list:     []*lhTicket{},
			},
		}
		err = dec.Decode(p.Project)
		if err != nil {
			return nil, err
		}
		pf.Close()
		mf, err := os.Open(filepath.Join(dir, "memberships.json"))
		if err == nil {
			defer mf.Close()
			var memberships projects.Memberships
			dec = json.NewDecoder(mf)
			err = dec.Decode(&memberships)
			if err != nil {
				return nil, err
			}
			mf.Close()
			var unique projects.Memberships
			seen := map[int]struct{}{}
			for _, membership := range memberships {
				if _, ok := seen[membership.UserID]; ok {
					continue
				}
				unique = append(unique, membership)
				seen[membership.UserID] = struct{}{}
			}
			p.memberships = unique
		}

		binPaths, err := filepath.Glob(filepath.Join(dir, "bins", "*.json"))
		if err != nil {
			return nil, err
		}
		for _, binPath := range binPaths {
			bf, err := os.Open(binPath)
			if err != nil {
				return nil, err
			}
			defer bf.Close()
			dec = json.NewDecoder(bf)
			b := &bins.Bin{}
			err = dec.Decode(b)
			if err != nil {
				return nil, err
			}
			bf.Close()
			p.bins.byID[b.ID] = b
			p.bins.list = append(p.bins.list, b)
		}
		sort.Slice(p.bins.list, func(i, j int) bool { return p.bins.list[i].ID < p.bins.list[j].ID })

		changesetPaths, err := filepath.Glob(filepath.Join(dir, "changesets", "*.json"))
		if err != nil {
			return nil, err
		}
		for _, changesetPath := range changesetPaths {
			cf, err := os.Open(changesetPath)
			if err != nil {
				return nil, err
			}
			defer cf.Close()
			dec = json.NewDecoder(cf)
			c := &changesets.Changeset{}
			err = dec.Decode(c)
			if err != nil {
				return nil, err
			}
			cf.Close()
			p.changesets.byRevision[strings.ToLower(c.Revision)] = c
			p.changesets.list = append(p.changesets.list, c)
		}
		sort.Slice(p.changesets.list, func(i, j int) bool {
			if p.changesets.list[i].ChangedAt != nil &&
				p.changesets.list[j].ChangedAt != nil {
				return p.changesets.list[i].ChangedAt.Before(*p.changesets.list[j].ChangedAt)
			}
			return p.changesets.list[i].Revision < p.changesets.list[j].Revision
		})

		messagePaths, err := filepath.Glob(filepath.Join(dir, "messages", "*.json"))
		if err != nil {
			return nil, err
		}
		for _, messagePath := range messagePaths {
			mf, err := os.Open(messagePath)
			if err != nil {
				return nil, err
			}
			defer mf.Close()
			dec = json.NewDecoder(mf)
			m := &messages.Message{}
			err = dec.Decode(m)
			if err != nil {
				return nil, err
			}
			mf.Close()
			p.messages = append(p.messages, m)
		}
		sort.Slice(p.messages, func(i, j int) bool { return p.messages[i].ID < p.messages[j].ID })

		milestonePaths, err := filepath.Glob(filepath.Join(dir, "milestones", "*.json"))
		if err != nil {
			return nil, err
		}
		for _, milestonePath := range milestonePaths {
			mf, err := os.Open(milestonePath)
			if err != nil {
				return nil, err
			}
			defer mf.Close()
			dec = json.NewDecoder(mf)
			m := &milestones.Milestone{}
			err = dec.Decode(m)
			if err != nil {
				return nil, err
			}
			mf.Close()
			p.milestones.byID[m.ID] = m
			p.milestones.list = append(p.milestones.list, m)
		}
		sort.Slice(p.milestones.list, func(i, j int) bool { return p.milestones.list[i].ID < p.milestones.list[j].ID })

		ticketDirs, err := filepath.Glob(filepath.Join(dir, "tickets", "*"))
		if err != nil {
			return nil, err
		}
		for _, ticketDir := range ticketDirs {
			tf, err := os.Open(filepath.Join(ticketDir, "ticket.json"))
			if err != nil {
				return nil, err
			}
			defer tf.Close()
			dec := json.NewDecoder(tf)
			t := &lhTicket{
				Ticket: &tickets.Ticket{},
				attachments: lhAttachments{
					byFilename: map[string]*lhAttachment{},
					list:       []*lhAttachment{},
				},
			}
			err = dec.Decode(t.Ticket)
			if err != nil {
				return nil, err
			}
			tf.Close()
			filenameMap := map[string]*tickets.Attachment{}
			for _, a := range t.Attachments {
				filenameMap[a.Attachment.Filename] = a.Attachment
			}
			attachmentPaths, err := filepath.Glob(filepath.Join(ticketDir, "*"))
			if err != nil {
				return nil, err
			}
			for _, attachmentPath := range attachmentPaths {
				if filepath.Base(attachmentPath) == "ticket.json" {
					continue
				}
				buf, err := ioutil.ReadFile(attachmentPath)
				if err != nil {
					return nil, err
				}
				a, ok := filenameMap[filepath.Base(attachmentPath)]
				if !ok {
					continue
				}
				attachment := &lhAttachment{
					Attachment: a,
					r:          bytes.NewReader(buf),
				}
				t.attachments.byFilename[attachment.Filename] = attachment
				t.attachments.list = append(t.attachments.list, attachment)
			}
			p.tickets.byNumber[t.Number] = t
			p.tickets.list = append(p.tickets.list, t)
		}
		sort.Slice(p.tickets.list, func(i, j int) bool { return p.tickets.list[i].Number < p.tickets.list[j].Number })

		e.projects.byID[p.ID] = p
		e.projects.list = append(e.projects.list, p)
	}
	sort.Slice(e.projects.list, func(i, j int) bool { return e.projects.list[i].ID < e.projects.list[j].ID })

	return e, nil
}
