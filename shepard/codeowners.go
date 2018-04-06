package shepard

import (
	"fmt"
	"net/http"

	"github.com/adam-hanna/randomstrings"
	"github.com/google/go-github/github"
)

func (s *ShepardBot) createBranch(repo *github.Repository, refObj *github.Reference) error {
	_, resp, err := s.gClient.Git.CreateRef(s.ctx, *repo.Owner.Login, *repo.Name, refObj)

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden {
		return &ShepardError{resp: resp}
	}

	return err
}

func (s *ShepardBot) commitFileToBranch(repo *github.Repository, branchName string) error {
	content := []byte(
		fmt.Sprintf("* @%s/%s", s.org.GetLogin(), s.maintainerTeam.GetName()),
	)

	_, _, err := s.gClient.Repositories.CreateFile(
		s.ctx,
		*repo.Owner.Login,
		*repo.Name,
		".github/CODEOWNERS",
		&github.RepositoryContentFileOptions{
			Branch:  github.String(branchName),
			Message: github.String("Adding CODEOWNERS file"),
			Content: content,
		},
	)

	return err
}

func (s *ShepardBot) createPR(repo *github.Repository, branchName string, branch *github.Branch) (*github.PullRequest, error) {
	prMessage := fmt.Sprintf("Hi there @%s!,\n\nI'm your helpful shepard and I've found that you are missing an important CODEOWNERS file which is mandated to be included for repos within this org (this ensures that the maintainers are pinged to review PR as they come in).\n\nThis PR is automatically created by [shepard](https://github.com/srizzling/shepard)\n\nThanks,\nShepard Bot", s.maintainerTeam.GetName())

	// Create a PR with the branch created
	newPR := &github.NewPullRequest{
		Title:               github.String("[AUTOMATED] Adding CODEOWNERS file"),
		MaintainerCanModify: github.Bool(true),
		Head:                github.String(branchName),
		Base:                github.String(branch.GetName()),
		Body:                github.String(prMessage),
	}

	pr, _, err := s.gClient.PullRequests.Create(s.ctx, *repo.Owner.Login, *repo.Name, newPR)
	return pr, err
}

// DoCreateCodeowners function will create a CODEOWNERS file in a branch, create a PR against the repo
// and set the reviewer (of the CODEOWNERS PR) as the maintainer team configured
func (s *ShepardBot) DoCreateCodeowners(repo *github.Repository, branch *github.Branch) (*github.PullRequest, error) {
	// Create a branch on the repo, from current master
	sRand, err := randomstrings.GenerateRandomString(5)
	if err != nil {
		return nil, err
	}

	branchName := fmt.Sprintf("add-codeowners-shepard-%s", sRand)
	newRef := github.Reference{
		Ref: github.String(fmt.Sprintf("refs/heads/%s", branchName)),
		Object: &github.GitObject{
			SHA: branch.Commit.SHA,
		},
	}

	err = s.createBranch(repo, &newRef)
	if err != nil {
		return nil, err
	}

	// Commit CODEOWNERS file to Branch
	err = s.commitFileToBranch(repo, branchName)
	if err != nil {
		return nil, err
	}

	// Create PR with newly created branch
	pr, err := s.createPR(repo, branchName, branch)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

// CheckCodeOwners verifies if the CODEOWNERS file exist in the repo, in the specfied branch
func (s *ShepardBot) CheckCodeOwners(repo *github.Repository, branch *github.Branch) (bool, *github.PullRequest, error) {
	// CODEOWNERS can be in .github or docs or in the root of the repo
	defaultCodeOwnersLoc := []string{"CODEOWNERS", ".github/CODEOWNERS", "docs/CODEOWNERS"}

	refName := github.String(fmt.Sprintf("refs/heads/%s", branch.GetName()))

	opt := &github.RepositoryContentGetOptions{
		Ref: *refName,
	}

	for _, coPath := range defaultCodeOwnersLoc {
		_, _, resp, err := s.gClient.Repositories.GetContents(
			s.ctx,
			*repo.Owner.Login,
			*repo.Name,
			coPath,
			opt,
		)

		if resp.StatusCode != http.StatusNotFound {
			return true, nil, nil // file not found
		} else if resp.StatusCode == http.StatusNotFound {
			continue // file not found skip to next iteration
		} else if err != nil {
			return false, nil, err
		}
	}

	// Check if there is a PR with the title [AUTOMATED] Adding CODEOWNERS file
	pulls, _, _ := s.gClient.PullRequests.List(s.ctx, *repo.Owner.Login, *repo.Name, nil)

	for _, pr := range pulls {
		if *pr.Title == "[AUTOMATED] Adding CODEOWNERS file" {
			return true, pr, nil
		}
	}

	// Looked everywhere the codeowners file couldn't be found
	return false, nil, nil
}
