package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	gogit "github.com/go-git/go-git/v5"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
)

const (
	exampleRepoUrl  = "git@github.com:parkermonson/meetupexamples.git"
	exampleFilePath = "gitexampledoc.txt"
	githubAPIDomain = "https://api.github.com"
	githubAPIPath   = "/repos/%s/%s/pulls"
)

//This command is a humorous example of how to change a file and check it into a repository, pushing it into a pull request, without

//This function clones a repo into memory, checks out a new branch, updates a file, pushes the branch and creates a pull request.
func updateAndPush(updateField, updateText string) {
	//load github credentials. This uses your id_rsa key for mac and linux users
	keys, err := loadPublicKeys()
	handleGitCmdError(err)

	//clone the repo
	repo, err := cloneRepoInMemory(keys, exampleRepoUrl)
	handleGitCmdError(err)

	//get the working tree
	worktree, err := repo.Worktree()
	handleGitCmdError(err)

	//create a new branch using epoch time to generate a name
	branchName := "working-update-" + fmt.Sprint(time.Now().Unix())
	err = checkoutNewBranch(worktree, branchName)
	handleGitCmdError(err)

	//update our in memory file
	err = updateWorkingtreeFile(worktree, exampleFilePath, updateField, updateText) //TODO fix this
	handleGitCmdError(err)

	//commit and push to git
	err = commitAndPushChanges(repo, worktree, keys)
	handleGitCmdError(err)

}

func loadPublicKeys() (*ssh.PublicKeys, error) {
	var publicKey *ssh.PublicKeys

	sshPath := os.Getenv("HOME") + "/.ssh/id_rsa"

	sshKey, _ := ioutil.ReadFile(sshPath)

	publicKey, err := ssh.NewPublicKeys("git", []byte(sshKey), "")
	if err != nil {
		return nil, err
	}

	return publicKey, nil
}

func cloneRepoInMemory(publicKey *ssh.PublicKeys, gitURL string) (*gogit.Repository, error) {
	fs := memfs.New()

	storer := memory.NewStorage()

	r, err := gogit.Clone(storer, fs, &gogit.CloneOptions{
		URL:  gitURL,
		Auth: publicKey,
	})
	if err != nil {
		log.Fatalf("error cloning repo into memory: %s\n", err.Error())
		return nil, err
	}

	return r, nil
}

func checkoutNewBranch(wtree *gogit.Worktree, branchName string) error {

	b := fmt.Sprintf("refs/heads/%s", branchName)
	bRef := plumbing.ReferenceName(b)

	err := wtree.Checkout(&gogit.CheckoutOptions{
		Create: false,
		Force:  false,
		Branch: bRef,
	})
	if err != nil {
		err := wtree.Checkout(&gogit.CheckoutOptions{
			Create: true,
			Force:  false,
			Branch: bRef,
		})
		if err != nil {
			fmt.Println("Checkout Really not working")
			return err
		}
	}
	return nil

}

func updateWorkingtreeFile(worktree *gogit.Worktree, filepath, updateField, updateText string) error {
	//get the existing file lines
	existinglines, err := scanFileFromWorkTree(worktree, filepath)
	if err != nil {
		return err
	}

	//update the file lines with our new message
	fileContent := ""
	for _, line := range existinglines {
		//if the line we are looking at is the field we want to update, replace it
		if strings.Contains(line, updateField) {
			line = updateField + "->" + updateText
		}
		//add the line  to the fileContent
		fileContent += line + "\n"
	}

	//overwrite the file in memory
	//make sure file is removed
	err = worktree.Filesystem.Remove(filepath)
	if err != nil {
		return err
	}

	//recreate the file
	newFile, err := worktree.Filesystem.Create(filepath)
	if err != nil {
		return err
	}

	//write the content to the file
	_, err = newFile.Write([]byte(fileContent))
	if err != nil {
		return err
	}

	//add the new file back into the worktree
	_, err = worktree.Add(filepath)
	if err != nil {
		return err
	}

	return nil
}

func scanFileFromWorkTree(worktree *gogit.Worktree, filePath string) ([]string, error) {
	f, err := worktree.Filesystem.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func commitAndPushChanges(repo *gogit.Repository, worktree *gogit.Worktree, key *ssh.PublicKeys) error {
	commitMessage := "updating doomsday clock"

	_, err := worktree.Commit(commitMessage, &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Parker Monson",
			Email: "fishfillet103@gmail.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	err = repo.Push(&gogit.PushOptions{
		Auth: key,
	})
	if err != nil {
		return err
	}

	return nil
}

type pullRequestBody struct {
	Title string `json:"title"`
	Head  string `json:"head"`
	Base  string `json:"base"`
	Body  string `json:"body"`
}

func handleGitCmdError(err error) {
	if err != nil {
		log.Panicf("Error executing git example cmd: %s\n", err.Error())
	}
}
