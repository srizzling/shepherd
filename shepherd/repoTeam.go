package shepherd

import (
	"net/http"

	"github.com/google/go-github/github"
)

// CheckTeamRepoManagement verifies if the team is an admin of the project
func (s *ShepardBot) CheckTeamRepoManagement(repo *github.Repository) (bool, error) {

	_, response, err := s.gClient.Organizations.IsTeamRepo(
		s.ctx,
		*s.maintainerTeam.ID,
		*repo.Owner.Login,
		*repo.Name,
	)

	// not managed
	if response.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if response.StatusCode == http.StatusOK {
		return true, nil
	}

	if err != nil {
		return false, err
	}

	return false, nil
}

// DoTeamRepoManagement sets the team as an admin of the repo
func (s *ShepardBot) DoTeamRepoManagement(repo *github.Repository) error {
	opt := &github.OrganizationAddTeamRepoOptions{
		Permission: "admin",
	}

	_, err := s.gClient.Organizations.AddTeamRepo(s.ctx, *s.maintainerTeam.ID, *repo.Owner.Login, *repo.Name, opt)
	return err
}
