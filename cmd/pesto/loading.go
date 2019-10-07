package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/pelletier/go-toml"
	"gopkg.in/src-d/go-git.v4"
	examples "gopkg.in/src-d/go-git.v4/_examples"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func main() {
	inputFile := "tmp/cc.toml"
	outputFile := "tmp/output.toml"

	cloneRepository("https://github.com/mustaine/go-pesto.git", "tmp")

	data, err := ioutil.ReadFile(inputFile)
	conf, err := toml.Load(string(data))
	examples.CheckIfError(err)

	fmt.Println("Version: ", conf.Get("apps.card-zetryer.version"))
	conf.Set("apps.card-zetryer.version", "1.4.3")
	fmt.Println("Version: ", conf.Get("apps.card-zetryer.version"))

	var buf bytes.Buffer
	if _, err := conf.WriteTo(&buf); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Result: ", buf.String())
	file, err := os.OpenFile(
		outputFile,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666,
	)

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file.Write(buf.Bytes())

	commit()

	//push()
}

func cloneRepository(url string, directory string) {
	// Clone the given repository to the given directory
	examples.Info("git clone %s %s --recursive", url, directory)

	os.RemoveAll("tmp")
	r, err := git.PlainClone(directory, false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: "whatever",
			Password: "USE_TOKEN",
		},
		URL:      url,
		Progress: os.Stdout,
	})

	examples.CheckIfError(err)

	// ... retrieving the branch being pointed by HEAD
	ref, err := r.Head()
	examples.CheckIfError(err)
	// ... retrieving the commit object
	commit, err := r.CommitObject(ref.Hash())
	examples.CheckIfError(err)

	fmt.Println(commit)
}

func commit() {
	r, err := git.PlainOpen("tmp")
	examples.CheckIfError(err)

	w, err := r.Worktree()
	examples.CheckIfError(err)

	// Adds the new file to the staging area.
	examples.Info("git add output.toml")
	_, err = w.Add("output.toml")
	examples.CheckIfError(err)

	// We can verify the current status of the worktree using the method Status.
	examples.Info("git status --porcelain")
	status, err := w.Status()
	examples.CheckIfError(err)

	fmt.Println(status)

	// Commits the current staging area to the repository, with the new file
	// just created. We should provide the object.Signature of Author of the
	// commit.
	examples.Info("git commit -m \"example go-git commit\"")
	commit, err := w.Commit("example go-git commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
			When:  time.Now(),
		},
	})
	// Prints the current HEAD to verify that all worked well.
	examples.Info("git show -s")
	obj, err := r.CommitObject(commit)
	examples.CheckIfError(err)

	fmt.Println(obj)
}

func push() {
	r, err := git.PlainOpen("tmp")
	examples.CheckIfError(err)

	examples.Info("git push")
	// push using default options
	err = r.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: "whatever",
			Password: "e1dacd86db1283801ed1fc845bc618bc12d246a7",
		},
	})
	examples.CheckIfError(err)
}

func makePullRequest() {
	client := github.NewClient(nil)

	newPR := &github.NewPullRequest{
		Title:               github.String("My awesome pull request"),
		Head:                github.String("helm-change"),
		Base:                github.String("master"),
		Body:                github.String("This is the description of the PR created with the package `github.com/google/go-github/github`"),
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := client.PullRequests.Create(context.Background(), "myOrganization", "myRepository", newPR)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("PR created: %s\n", pr.GetHTMLURL())
}
