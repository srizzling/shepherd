package shepherd

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type ShepardBot struct {
	gClient        *github.Client
	ctx            context.Context
	maintainerTeam *github.Team
	org            *github.Organization
}

// ShepardError is a generic error container for reporting errors/http status code errors from the Github API
type ShepardError struct {
	resp *github.Response
}

func (e *ShepardError) Error() string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(e.resp.Response.Body)
	responseBody := buf.String() // Does a complete copy of the bytes in the buffer.
	return fmt.Sprintf("sherpard has encountered an error: %s-%d: %s", e.resp.Request.RemoteAddr, e.resp.StatusCode, responseBody)
}

// NewBot creates a new ShepardBot based off the baseURL(provide empty string if you want to default to basic github)
func NewBot(baseURL string, token string, maintainerTeamName string, orgName string) (*ShepardBot, error) {
	// initialize a new github client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	ctx := context.Background()
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// Setup baseUrl for github enterprise, else default to Github.com.
	if baseURL != "" {
		var err error
		client.BaseURL, err = url.Parse(baseURL + "/api/v3/")
		if err != nil {
			return nil, err
		}
	}

	bot := &ShepardBot{
		gClient: client,
		ctx:     ctx,
	}

	// set github org to bot
	err := bot.setOrg(orgName)
	if err != nil {
		return nil, err
	}

	// set maintainer team to org
	err = bot.setMaintainerTeam(maintainerTeamName)
	if err != nil {
		return nil, err
	}

	return bot, nil
}

// RetreiveRepos returns a list of repos within the organization
func (s *ShepardBot) RetreiveRepos() ([]*github.Repository, error) {
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 10},
	}

	var allRepos []*github.Repository
	for {
		repos, resp, err := s.gClient.Repositories.ListByOrg(s.ctx, *s.org.Login, opt)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

// GetBranch function return a branch obj depending on the name provided
func (s *ShepardBot) GetBranch(repo *github.Repository, branchName string) (*github.Branch, error) {
	branch, _, err := s.gClient.Repositories.GetBranch(s.ctx, *repo.Owner.Login, *repo.Name, branchName)
	if err != nil {
		return nil, err
	}
	return branch, nil
}

func (s *ShepardBot) setOrg(orgName string) error {
	org, _, err := s.gClient.Organizations.Get(s.ctx, orgName)
	if err != nil {
		return err
	}
	s.org = org
	return nil
}
