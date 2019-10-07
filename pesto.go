package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"time"
)

var ghToken = "token"
var repository = "https://github.dns.ad.zopa.com/zopaUK/pesto.git"
var updatedVersion = "1.4.5"
var inputFile = "tmp/streamster.yaml"
var outputFile = "tmp/streamster-output.yaml"

func main() {
	repository := cloneRepository(repository, "tmp")
	updateVersion()
	commit(repository)
	push(repository)
	makePullRequest()
}

func updateVersion() {
	b, err := ioutil.ReadFile(inputFile)
	CheckIfError(err)

	chart := map[string]interface{}{
		"deployment": map[interface{}]interface{}{},
	}

	err = yaml.Unmarshal(b, &chart)
	CheckIfError(err)

	deployment := chart["deployment"].(map[interface{}]interface{})

	deployment["version"] = updatedVersion
	chart["deployment"] = deployment
	newFile, err := yaml.Marshal(chart)
	CheckIfError(err)

	//err = ioutil.WriteFile(file, newFile, 0)
	err = ioutil.WriteFile(outputFile, newFile, 0666)
	CheckIfError(err)
}

func cloneRepository(url string, directory string) *git.Repository {
	// Clone the given repository to the given directory
	Info("git clone %s %s --recursive", url, directory)

	os.RemoveAll("tmp")
	repository, err := git.PlainClone(directory, false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: "whatever",
			Password: ghToken,
		},
		URL:      url,
		Progress: os.Stdout,
	})

	CheckIfError(err)

	return repository
}

func commit(repository *git.Repository) {
	w, err := repository.Worktree()
	CheckIfError(err)

	// Adds the new file to the staging area.
	Info("git add streamster-output.yaml")
	_, err = w.Add("streamster-output.yaml")
	CheckIfError(err)

	// We can verify the current status of the worktree using the method Status.
	Info("git status --porcelain")
	status, err := w.Status()
	CheckIfError(err)

	fmt.Println(status)

	// Commits the current staging area to the repository, with the new file
	// just created. We should provide the object.Signature of Author of the
	// commit.
	Info("git commit -m \"example go-git commit\"")
	commit, err := w.Commit("example go-git commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
			When:  time.Now(),
		},
	})
	// Prints the current HEAD to verify that all worked well.
	Info("git show -s")
	obj, err := repository.CommitObject(commit)
	CheckIfError(err)

	fmt.Println(obj)
}

func push(repository *git.Repository) {
	Info("git push")
	upstreamReference := plumbing.ReferenceName("+refs/heads/master")
	downstreamReference := plumbing.ReferenceName("refs/heads/helm-change")
	referenceList := append([]config.RefSpec{},
		config.RefSpec(upstreamReference+":"+downstreamReference))

	err := repository.Push(&git.PushOptions{
		RefSpecs: referenceList,
		Auth: &http.BasicAuth{
			Username: "-",
			Password: ghToken,
		},
	})
	CheckIfError(err)
}

func githubClient() *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client, err := github.NewEnterpriseClient("https://github.dns.ad.zopa.com", "zopaUK", tc)
	CheckIfError(err)
	return client
}

func makePullRequest() {
	client := githubClient()

	newPR := &github.NewPullRequest{
		Title:               github.String("My awesome pull request"),
		Head:                github.String("helm-change"),
		Base:                github.String("master"),
		Body:                github.String("This is the description of the PR created with the package `github.com/google/go-github/github`"),
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := client.PullRequests.Create(context.Background(), "zopaUK", "pesto", newPR)
	if err != nil {
		fmt.Println(err)
		return
	}

	Info("PR created: %s\n", pr.GetHTMLURL())
}

func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}
