package shepherd

import (
	"github.com/google/go-github/github"
)

func (s *ShepardBot) DoProtectBranch(repo *github.Repository, branch *github.Branch) error {
	owner := *repo.Owner.Login
	repoName := *repo.Name

	protect := &github.ProtectionRequest{
		RequiredStatusChecks: &github.RequiredStatusChecks{
			Strict:   false,
			Contexts: []string{},
		},
	}

	_, _, err := s.gClient.Repositories.UpdateBranchProtection(s.ctx, owner, repoName, branch.GetName(), protect)

	if err != nil {
		return err
	}

	patch := &github.PullRequestReviewsEnforcementUpdate{
		RequireCodeOwnerReviews: true,
		DismissStaleReviews:     github.Bool(true),
	}

	_, _, err = s.gClient.Repositories.UpdatePullRequestReviewEnforcement(
		s.ctx,
		owner,
		repoName,
		branch.GetName(),
		patch,
	)

	return err
}

// CheckProtectionBranch verifies if the the branch is a protected branch and verifies if it has to be verified by CODEOWNERS
func (s *ShepardBot) CheckProtectionBranch(repo *github.Repository, branch *github.Branch) (bool, error) {
	// Check if branch is even protected if its not return instantly
	if !branch.GetProtected() {
		return false, nil
	}

	reviewEnforcement, _, err := s.gClient.Repositories.GetPullRequestReviewEnforcement(s.ctx,
		*repo.Owner.Login,
		*repo.Name,
		branch.GetName(),
	)

	if err != nil {
		return false, err
	}

	// Check if review enforcement is set for codeowners
	return reviewEnforcement.RequireCodeOwnerReviews, nil
}
