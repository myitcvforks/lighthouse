package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/nwidger/lighthouse"
	"github.com/nwidger/lighthouse/changesets"
)

func getAccountAndProject(repoPath string) (string, int, error) {
	buf, err := ioutil.ReadFile(filepath.Join(repoPath, ".lhproj"))
	if err != nil {
		return "", 0, err
	}

	url, err := url.Parse(strings.TrimSpace(string(buf)))
	if err != nil {
		return "", 0, err
	}

	idx := strings.Index(url.Host, ".lighthouseapp.com")
	if idx == -1 {
		return "", 0, fmt.Errorf("unable to determine account name %q", url.String())
	}

	account := url.Host[:idx]
	if len(account) == 0 {
		return "", 0, fmt.Errorf("empty account name %q", url.String())
	}

	idx = strings.LastIndex(url.Path, "/")
	if idx == -1 {
		return "", 0, fmt.Errorf("unable to determine project ID %q", url.Path)
	}

	projectStr := url.Path[idx+1:]
	projectID, err := strconv.Atoi(projectStr)
	if err != nil {
		return "", 0, fmt.Errorf("unable to parse project ID %q", projectStr)
	}

	return account, projectID, nil
}

func getToken(repoPath, commitAuthor string) (string, error) {
	buf, err := ioutil.ReadFile(filepath.Join(repoPath, ".lhkeys"))
	if err != nil {
		return "", err
	}

	token := ""

	for _, line := range strings.Split(string(buf), "\n") {
		f := strings.Fields(strings.TrimSpace(line))
		if len(f) != 2 {
			return "", fmt.Errorf("invalid line %q", line)
		}

		lineAuthor, lineToken := f[0], f[1]
		if lineAuthor == commitAuthor {
			token = lineToken
			break
		}
	}

	if len(token) == 0 {
		return "", fmt.Errorf("unable to find token for %q", commitAuthor)
	}

	return token, nil
}

func runSVNLook(args ...string) string {
	cmd := exec.Command("svnlook", args...)

	output, err := cmd.Output()
	if err != nil {
		log.Fatal("svnlook", args, ":", err)
	}

	return string(output)
}

func createChangeset(repoPath, revision, commitAuthor string) (*changesets.Changeset, error) {
	commitLog := runSVNLook("log", repoPath, "-r", revision)
	commitDate := runSVNLook("date", repoPath, "-r", revision)
	commitChanged := runSVNLook("changed", repoPath, "-r", revision)

	ss := strings.Split(commitDate, "(")
	if len(ss) == 0 {
		return nil, fmt.Errorf("unable to determine commit date %q", commitDate)
	}
	commitTime, err := time.Parse("2006-01-02 15:04:05 -0700", strings.TrimSpace(ss[0]))
	if err != nil {
		return nil, fmt.Errorf("unable to parse commit date %q", commitDate)
	}

	base := strings.TrimSpace(filepath.Base(repoPath))
	if len(base) == 0 {
		return nil, fmt.Errorf("base of %s is empty", repoPath)
	}

	title := fmt.Sprintf("%s committed changeset [%s]", commitAuthor, revision)
	body := fmt.Sprintf(`Commit log:

%s

[ViewVC](http://example.com/viewvc/%s?view=rev&amp;revision=%s)`, commitLog, base, revision)

	c := &changesets.Changeset{
		Title:     title,
		Body:      body,
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

	return c, nil
}

func gatherAndPost(repoPath, revision string) error {
	account, projectID, err := getAccountAndProject(repoPath)
	if err != nil {
		return err
	}

	commitAuthor := strings.TrimSpace(runSVNLook("author", repoPath, "-r", revision))

	token, err := getToken(repoPath, commitAuthor)
	if err != nil {
		return err
	}

	c, err := createChangeset(repoPath, revision, commitAuthor)
	if err != nil {
		return err
	}

	client := &http.Client{
		Transport: &lighthouse.Transport{
			Token:            token,
			TokenAsBasicAuth: true,
		},
	}

	s := lighthouse.NewService(account, client)
	cs := changesets.NewService(s, projectID)

	c, err = cs.Create(c)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("usage: <repo-path> <revision>\n")
		os.Exit(1)
	}
	repoPath, revision := os.Args[1], os.Args[2]

	f, err := os.OpenFile("/tmp/svn-hooks.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	err = gatherAndPost(repoPath, revision)
	if err != nil {
		log.Fatalf("%s: %s: %s\n", repoPath, revision, err.Error())
	}
}
