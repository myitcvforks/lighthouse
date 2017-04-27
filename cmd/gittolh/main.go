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
	account := strings.TrimSpace(mustRunGit("config", "--get", "lighthouse.account"))
	projectStr := strings.TrimSpace(mustRunGit("config", "--get", "lighthouse.project"))

	if len(account) == 0 {
		return "", 0, fmt.Errorf("empty account name %q", account)
	}

	projectID, err := strconv.Atoi(projectStr)
	if err != nil {
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

	token := strings.TrimSpace(mustRunGit("config", "--get", fmt.Sprintf("lighthouse.keys.%s", name)))

	if len(token) == 0 {
		return "", fmt.Errorf("unable to find token for %q", name)
	}

	return token, nil
}

func getGitwebURL() string {
	gitwebURL, _ := runGit("config", "--get", "lighthouse.gitweb-url")
	gitwebURL = strings.TrimSpace(gitwebURL)
	return gitwebURL
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
	gitwebURL := getGitwebURL()

	cc := []*changesets.Changeset{}

	commits := strings.TrimSpace(mustRunGit("log", "-s", "--format=%H", fmt.Sprintf("%s..%s", oldrev, newrev)))

	for _, revision := range strings.Split(commits, "\n") {
		commitAuthor := strings.TrimSpace(mustRunGit("show", "-s", "--format=%an", revision))
		commitEmail := strings.TrimSpace(mustRunGit("show", "-s", "--format=%ae", revision))
		commitLog := mustRunGit("show", "-s", "--format=%s", newrev)
		commitDiffStat := mustRunGit("diff", "--stat", fmt.Sprintf("%s^..%s", revision, revision))
		commitDate := mustRunGit("show", "-s", "--format=%at", newrev)
		commitChanged := mustRunGit("diff-tree", "-r", "--name-status", "--no-commit-id", revision)

		sec, err := strconv.ParseInt(strings.TrimSpace(commitDate), 10, 64)
		if err != nil {
			return nil, err
		}
		commitTime := time.Unix(sec, 0)

		title := fmt.Sprintf("%s committed changeset [%s]", commitAuthor, revision)
		body := fmt.Sprintf(`
%s

%s`, commitLog, commitDiffStat)
		if len(gitwebURL) > 0 {
			body += fmt.Sprintf(`
[gitweb](%s;a=commit;h=%s)`, gitwebURL, revision)
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
			return err
		}
		tokens[c.Committer] = token
	}

	for _, c := range cc {
		lt.Token = tokens[c.Committer]
		c, err = cs.Create(c)
		if err != nil {
			return err
		}
		<-time.After(500 * time.Millisecond)
	}

	return nil
}

func main() {
	usage := "<oldrev> <newrev> <refname> via stdin"

	if len(os.Args) != 1 {
		fmt.Fprintf(os.Stderr, usage)
		os.Exit(1)
	}

	f, err := os.OpenFile("/tmp/git-hooks.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	oldrev, newrev, refname := "", "", ""
	n, err := fmt.Fscanf(os.Stdin, "%s %s %s", &oldrev, &newrev, &refname)
	if err != nil {
		log.Fatal(err)
	}
	if n != 3 {
		log.Fatal(usage)
	}

	err = gatherAndPost(oldrev, newrev, refname)
	if err != nil {
		log.Fatalf("%s %s %s: %s\n", oldrev, newrev, refname, err.Error())
	}
}
