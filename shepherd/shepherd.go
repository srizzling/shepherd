package shepherd

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// ShepardBot is the main bot object that gets created
type ShepardBot struct {
	gClient *github.Client
	ctx     context.Context
	Repos   map[*github.Repository]RepoConfig
}

// Config struct holds the configuration data to do things
type Config struct {
	IncludeUserRepo bool
	BaseURL         string
	DryRun          bool
	Debug           bool
	Organizations   []OrganizationsConfig
	Repos           []RepoConfig
	GithubToken     string `mapstructure:"GITHUB_TOKEN" yaml:"github_token, omitempty"`
}

// OrganizationsConfig is the config section for configuring organization
type OrganizationsConfig struct {
	OrgName         string `mapstructure:"OrgName" yaml:"orgName, omitempty"`
	Maintainer      string
	Labels          []Label
	Templates       map[string]string
	ProtectedBranch string `mapstructure:"protected_branch" yaml:"protected_branch"`
}

// ReposConfig is the config section for configuring organization
type RepoConfig struct {
	Name            string
	Maintainer      string
	Labels          []Label
	ProtectedBranch string
	Templates       map[string]string
	GHMaintainer    *github.Team
	GHLabels        []*github.Label
}

type Label struct {
	Name  string
	Color string
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
func NewBot(config Config) (*ShepardBot, error) {
	// initialize a new github client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.GithubToken},
	)
	ctx := context.Background()
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// Setup baseUrl for github enterprise, else default to Github.com.
	if config.BaseURL != "" {
		var err error
		client.BaseURL, err = url.Parse(config.BaseURL + "/api/v3/")
		if err != nil {
			return nil, err
		}
	}

	bot := &ShepardBot{
		gClient: client,
		ctx:     ctx,
	}

	repoMap := make(map[*github.Repository]RepoConfig)
	for _, org := range config.Organizations {
		orgMap, err := bot.retreiveRepoOnOrg(org)
		if err != nil {
			return nil, err
		}
		repoMap = mapUnion(repoMap, orgMap)
	}
	bot.Repos = repoMap
	return bot, nil
}

func mapUnion(m1, m2 map[*github.Repository]RepoConfig) map[*github.Repository]RepoConfig {
	for ia, va := range m1 {
		m2[ia] = va
	}
	return m2
}

// RetreiveReposOnOrg returns a list of repos within the organization
func (s *ShepardBot) retreiveRepoOnOrg(org OrganizationsConfig) (map[*github.Repository]RepoConfig, error) {
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 10},
	}

	orgObj, _, err := s.gClient.Organizations.Get(s.ctx, org.OrgName)
	if err != nil {
		return nil, err
	}

	var allRepos []*github.Repository
	for {
		repos, resp, err := s.gClient.Repositories.ListByOrg(s.ctx, *orgObj.Login, opt)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	repoMap := make(map[*github.Repository]RepoConfig)

	for _, r := range allRepos {
		mTeam, err := s.getMaintainerTeam(orgObj, org.Maintainer)
		if err != nil {
			return nil, err
		}

		repoMap[r] = RepoConfig{
			Name:            r.GetFullName(),
			Maintainer:      org.Maintainer,
			Labels:          org.Labels,
			GHMaintainer:    mTeam,
			GHLabels:        nil,
			ProtectedBranch: org.ProtectedBranch,
		}
	}
	return repoMap, nil
}

// GetBranch function return a branch obj depending on the name provided
func (s *ShepardBot) GetBranch(repo *github.Repository, branchName string) (*github.Branch, error) {
	branch, _, err := s.gClient.Repositories.GetBranch(s.ctx, *repo.Owner.Login, *repo.Name, branchName)
	if err != nil {
		return nil, err
	}
	return branch, nil
}
