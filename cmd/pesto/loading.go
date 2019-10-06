package main

import (
	"bytes"
	"fmt"
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

}

func cloneRepository(url string, directory string) {
	// Clone the given repository to the given directory
	examples.Info("git clone %s %s --recursive", url, directory)

	os.RemoveAll("tmp")
	r, err := git.PlainClone(directory, false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: "whatever",
			Password: "e1dacd86db1283801ed1fc845bc618bc12d246a7",
		},
		URL:      url,
		Progress: os.Stdout,
	})

	//r, err := git.PlainClone(directory, false, &git.CloneOptions{
	//	URL:               url,
	//	RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	//})

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
	//examples.Info("git add example-git-file")
	//_, err = w.Add("example-git-file")
	//examples.CheckIfError(err)

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
