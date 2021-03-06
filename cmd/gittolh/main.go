package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/nwidger/lighthouse"
	"github.com/nwidger/lighthouse/changesets"
)

func getAccountAndProject() (string, int, error) {
	account, _ := runGit("config", "--get", "lighthouse.account")
	account = strings.TrimSpace(account)

	projectStr, _ := runGit("config", "--get", "lighthouse.project")
	projectStr = strings.TrimSpace(projectStr)

	if len(account) == 0 {
		log.Printf("gittolh: unable to find Lighthouse account name, please run 'git config lighthouse.account <account-name>' on remote repository")
		return "", 0, fmt.Errorf("empty account name %q", account)
	}

	projectID, err := strconv.Atoi(projectStr)
	if err != nil {
		log.Printf("gittolh: unable to find Lighthouse project ID, please run 'git config lighthouse.project <project-id>' on remote repository")
		return "", 0, fmt.Errorf("unable to parse project ID %q", projectStr)
	}

	return account, projectID, nil
}

func getToken(commitEmail string) (string, error) {
	// use name portion of email as git config key
	name := commitEmail
	if idx := strings.Index(commitEmail, "@"); idx != -1 {
		name = commitEmail[:idx]
	}

	token, _ := runGit("config", "--get", fmt.Sprintf("lighthouse.keys.%s", name))
	token = strings.TrimSpace(token)
	if len(token) == 0 {
		log.Printf("gittolh: unable to find Lighthouse token for %s, please run 'git config lighthouse.keys.%s <token>' on remote repository", name, name)
		return "", fmt.Errorf("unable to find token for %q", name)
	}

	return token, nil
}

func getFooter() string {
	footer, err := runGit("config", "--get", "lighthouse.footer")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(footer)
}

func mustRunGit(args ...string) string {
	output, err := runGit(args...)
	if err != nil {
		log.Fatal("git", args, ":", err)
	}
	return output
}

func runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	return string(output), err
}

func createChangesets(oldrev, newrev, refname string) ([]*changesets.Changeset, error) {
	var change, revType, refType, refShortName, revSpec string

	oldrev = strings.TrimSpace(mustRunGit("rev-parse", oldrev))
	newrev = strings.TrimSpace(mustRunGit("rev-parse", newrev))

	switch {
	case strings.Count(oldrev, "0") == len(oldrev):
		change = "created"
	case strings.Count(oldrev, "0") == len(newrev):
		change = "deleted"
	default:
		change = "updated"
	}

	switch change {
	case "created", "updated":
		revType = strings.TrimSpace(mustRunGit("cat-file", "-t", newrev))
	case "deleted":
		revType = strings.TrimSpace(mustRunGit("cat-file", "-t", oldrev))
	}

	if strings.HasPrefix(refname, "refs/tags/") && revType == "commit" {
		refType = "tag"
		refShortName = strings.TrimPrefix(refname, "refs/tags/")
	} else if strings.HasPrefix(refname, "refs/tags/") && revType == "tag" {
		refType = "annotated tag"
		refShortName = strings.TrimPrefix(refname, "refs/tags/")
	} else if strings.HasPrefix(refname, "refs/heads/") && revType == "commit" {
		refType = "branch"
		refShortName = strings.TrimPrefix(refname, "refs/heads/")
	} else if strings.HasPrefix(refname, "refs/remotes/") && revType == "commit" {
		refType = "tracking branch"
		refShortName = strings.TrimPrefix(refname, "refs/remotes/")
	} else {
		log.Printf("unknown update type %q (%q)", refname, revType)
		return nil, nil
	}

	if change == "deleted" || refType != "branch" {
		return nil, nil
	}

	if change == "created" {
		revSpec = fmt.Sprintf("HEAD..%s", newrev)
	} else {
		revSpec = fmt.Sprintf("%s..%s", oldrev, newrev)
	}

	footer := getFooter()

	cc := []*changesets.Changeset{}

	commits := strings.TrimSpace(mustRunGit("log", "-s", "--format=%H", revSpec))

	for _, revision := range strings.Split(commits, "\n") {
		commitAuthor := strings.TrimSpace(mustRunGit("show", "-s", "--format=%an", revision))
		commitEmail := strings.TrimSpace(mustRunGit("show", "-s", "--format=%ae", revision))
		commitLog := mustRunGit("show", "-s", "--format=%s%n%n%b", revision)
		commitDiffStat := mustRunGit("diff", "--stat", fmt.Sprintf("%s^..%s", revision, revision))
		commitDate := mustRunGit("show", "-s", "--format=%at", revision)
		commitChanged := mustRunGit("diff-tree", "-r", "--name-status", "--no-commit-id", revision)

		sec, err := strconv.ParseInt(strings.TrimSpace(commitDate), 10, 64)
		if err != nil {
			return nil, err
		}
		commitTime := time.Unix(sec, 0)

		title := fmt.Sprintf("%s committed changeset [%s] which %s %s %s", commitAuthor, revision, change, refType, refShortName)
		body := fmt.Sprintf(`%s %s %s:

%s

@@@
%s
@@@`, strings.Title(change), refType, refShortName, commitLog, commitDiffStat)
		if len(footer) > 0 {
			ftr := footer
			if strings.Contains(ftr, "%s") {
				ftr = strings.Replace(ftr, "%s", revision, 1)
			}
			body += "\n\n" + ftr
		}

		c := &changesets.Changeset{
			Title:     title,
			Body:      body,
			Committer: commitEmail,
			Revision:  revision,
			ChangedAt: &commitTime,
			Changes:   changesets.Changes{},
		}

		for _, line := range strings.Split(commitChanged, "\n") {
			if len(line) == 0 {
				continue
			}
			idx := strings.IndexAny(line, " \t\f")
			if idx == -1 {
				continue
			}
			op, field := strings.TrimSpace(line[:idx]),
				strings.TrimSpace(line[idx:])
			if len(op) == 0 || len(field) == 0 {
				continue
			}
			c.Changes = append(c.Changes, &changesets.Change{
				Operation: op,
				Path:      field,
			})
		}

		cc = append(cc, c)
	}

	return cc, nil
}

func gatherAndPost(oldrev, newrev, refname string) error {
	var ok bool

	account, projectID, err := getAccountAndProject()
	if err != nil {
		return err
	}

	cc, err := createChangesets(oldrev, newrev, refname)
	if err != nil {
		return err
	}
	if len(cc) == 0 {
		return nil
	}

	lt := &lighthouse.Transport{
		TokenAsBasicAuth: true,
	}

	client := &http.Client{
		Transport: lt,
	}

	s := lighthouse.NewService(account, client)
	cs := changesets.NewService(s, projectID)

	tokens := map[string]string{}

	for _, c := range cc {
		if _, ok = tokens[c.Committer]; ok {
			continue
		}
		token, err := getToken(c.Committer)
		if err != nil {
			continue
		}
		tokens[c.Committer] = token
	}

	err = nil

	// iterate list in reverse so older changes are posted first
	for i := len(cc) - 1; i >= 0; i = i - 1 {
		c := cc[i]
		lt.Token = tokens[c.Committer]
		if len(lt.Token) == 0 {
			continue
		}
		c, err = cs.Create(c)
	}

	return err
}

func main() {
	usage := "<oldrev> <newrev> <refname>"

	if len(os.Args) != 1 && len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
		os.Exit(1)
	}

	f, err := os.OpenFile("/tmp/git-hooks.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	if len(os.Args) == 4 {
		oldrev, newrev, refname := os.Args[1], os.Args[2], os.Args[3]

		fmt.Fprintln(f, oldrev, newrev, refname)
		err = gatherAndPost(oldrev, newrev, refname)
		if err != nil {
			fmt.Fprintf(f, "%s %s %s: %s\n", oldrev, newrev, refname, err.Error())
		}
	} else {
		oldrev, newrev, refname := "", "", ""

		for {
			n, err := fmt.Fscanf(os.Stdin, "%s %s %s", &oldrev, &newrev, &refname)
			if err != nil || n != 3 {
				break
			}

			fmt.Fprintln(f, oldrev, newrev, refname)
			err = gatherAndPost(oldrev, newrev, refname)
			if err != nil {
				fmt.Fprintf(f, "%s %s %s: %s\n", oldrev, newrev, refname, err.Error())
			}
		}
	}

}
