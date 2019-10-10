package main

import (
	"context"
	"errors"
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
	"strings"
	"time"
)

var args PestoArgs

//var helmRepository = "https://github.dns.ad.zopa.com/zopaUK/helm-state.git"
//var helmRepository = "https://github.dns.ad.zopa.com/zopaUK/pesto.git"
var helmRepository = "https://github.com/mustaine/go-pesto.git"
var temp = "tmp"

type PestoArgs struct {
	Token        string
	Environment  string
	Namespace    string
	Applications []string
	Version      string
}

func (args *PestoArgs) files() []string {
	var f []string
	for _, v := range args.Applications {
		f = append(f, fmt.Sprintf("%s/%s/%s", args.Environment, args.Namespace, v))

	}
	return f
}

func (args *PestoArgs) branchName() string {
	return fmt.Sprintf("helm-change-%s-%s-%s", args.Environment, args.Namespace, args.Version)
}

func (args *PestoArgs) commitMessage() string {
	return fmt.Sprintf("helm change %s-%s-%s", args.Environment, args.Namespace, args.Version)
}

func (args *PestoArgs) pullRequestTitle() string {
	return fmt.Sprintf("Automated PR %s-%s-%s", args.Environment, args.Namespace, args.Version)
}

func (args *PestoArgs) jiraTicket() string {
	ticket := strings.ToUpper(fmt.Sprintf("PESTO-%s-%s-%s", args.Environment, args.Namespace, args.Version))
	return fmt.Sprintf("https://jira.zopa/browse/%s", ticket)
}

func main() {
	pestoArgs, err := validateArgs(os.Args)
	CheckIfError(err)
	args = *pestoArgs

	repository := cloneRepository(helmRepository, temp)
	updateVersion()
	commit(repository)
	push(repository)
	pr, err := makePullRequest()
	mergePullRequest(pr)
}

func validateArgs(args []string) (*PestoArgs, error) {
	numArgs := 6
	if len(args) != numArgs {
		return nil, errors.New(fmt.Sprintf("Arguments %d different than %d", len(args), numArgs))
	}

	token := args[1]
	env := args[2]
	namespace := args[3]
	apps := strings.Split(args[4], ",")
	version := args[5]
	return &PestoArgs{token, env, namespace, apps, version}, nil
}

func updateVersion() {
	for _, file := range args.files() {
		b, err := ioutil.ReadFile(temp + "/" + file)
		CheckIfError(err)

		chart := map[string]interface{}{
			"deployment": map[interface{}]interface{}{},
		}

		err = yaml.Unmarshal(b, &chart)
		CheckIfError(err)

		deployment := chart["deployment"].(map[interface{}]interface{})

		deployment["version"] = args.Version
		chart["deployment"] = deployment
		newFile, err := yaml.Marshal(chart)
		CheckIfError(err)

		err = ioutil.WriteFile(temp+"/"+file, newFile, 0)
		CheckIfError(err)
	}
}

func cloneRepository(url string, directory string) *git.Repository {
	// Clone the given repository to the given directory
	Info("git clone %s %s --recursive", url, directory)

	os.RemoveAll(temp)
	repository, err := git.PlainClone(directory, false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: "whatever",
			Password: args.Token,
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
	for _, file := range args.files() {
		Info("git add " + file)
		_, err = w.Add(file)
		CheckIfError(err)
	}

	// We can verify the current status of the worktree using the method Status.
	Info("git status --porcelain")
	status, err := w.Status()
	CheckIfError(err)
	fmt.Println(status)

	// Commits the current staging area to the repository, with the new file
	// just created. We should provide the object.Signature of Author of the
	// commit.
	Info("git commit -m \"" + args.commitMessage() + "\"")
	commit, err := w.Commit(args.commitMessage(), &git.CommitOptions{
		Author: &object.Signature{
			Name:  "zopadev",
			Email: "dev2@zopa.com",
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
	downstreamReference := plumbing.ReferenceName("refs/heads/" + args.branchName())
	referenceList := append([]config.RefSpec{},
		config.RefSpec(upstreamReference+":"+downstreamReference))

	err := repository.Push(&git.PushOptions{
		RefSpecs: referenceList,
		Auth: &http.BasicAuth{
			Username: "-",
			Password: args.Token,
		},
	})
	CheckIfError(err)
}

func githubClient() *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: args.Token},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	//client, err := github.NewEnterpriseClient("https://github.dns.ad.zopa.com/api/v3", "zopaUK", tc)
	//CheckIfError(err)
	client := github.NewClient(tc)
	return client
}

func makePullRequest() (*github.PullRequest, error) {
	Info("make pull request")

	client := githubClient()

	newPR := &github.NewPullRequest{
		Title:               github.String(args.pullRequestTitle()),
		Head:                github.String(args.branchName()),
		Base:                github.String("master"),
		Body:                github.String(args.jiraTicket()),
		MaintainerCanModify: github.Bool(true),
	}

	//pr, _, err := client.PullRequests.Create(context.Background(), "zopaUK", "helm-state", newPR)
	pr, _, err := client.PullRequests.Create(context.Background(), "mustaine", "go-pesto", newPR)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	Info("PR created: %s\n", pr.GetHTMLURL())
	return pr, nil
}

func getPullRequest(pullRequest *github.PullRequest) {
	client := githubClient()
	pr, _, err := client.PullRequests.Get(context.Background(), "zopaUK", "pesto", pullRequest.GetNumber())
	if err != nil {
		fmt.Println(err)
		return
	}

	Info("PR GetMergeableState: %s\n", pr.GetMergeableState())
	Info("PR GetMergeable: %s\n", pr.GetMergeable())

}
func mergePullRequest(pr *github.PullRequest) {
	Info("merge pull request")

Loop:
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		select {
		//check exitMessage to see whether to break out or not
		case <-ctx.Done():
			break Loop
		case <-time.After(10 * time.Second):
			break Loop
		//do something repeatedly very fast in the for loop
		default:
			time.Sleep(time.Second * 5)
			getPullRequest(pr)
			// stuff
		}
	}
	//getPullRequest(pr)

	//client := githubClient()
	//
	//options := &github.PullRequestOptions{
	//	MergeMethod: "merge",
	//}
	//
	//result, _, err := client.PullRequests.Merge(context.Background(), "zopaUK", "helm-state", pr.GetNumber(), "", options)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	//Info("PR merged: %s\n", result.GetSHA())
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
