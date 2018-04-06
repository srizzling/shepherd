package shepard

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-github/github"
)

func (s *ShepardBot) retreiveTeams(orgName string) ([]*github.Team, error) {
	opt := &github.ListOptions{
		PerPage: 10,
	}
	var allTeams []*github.Team
	for {
		teams, resp, err := s.gClient.Organizations.ListTeams(s.ctx, orgName, opt)
		if err != nil {
			return nil, err
		}
		allTeams = append(allTeams, teams...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allTeams, nil
}

func (s *ShepardBot) setMaintainerTeam(maintainerTeamName string) error {
	teams, err := s.retreiveTeams(s.org.GetLogin())
	if err != nil {
		return err
	}

	for _, team := range teams {
		if strings.EqualFold(maintainerTeamName, s.org.GetLogin()+"/"+*team.Name) {
			s.maintainerTeam = team
			return nil
		}
	}

	errMsg := fmt.Sprintf("Team (%s) not found within org", maintainerTeamName)
	err = errors.New(errMsg)
	return err
}
