package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
	"github.com/srizzling/shepherd/shepherd"
)

var (
	token      string
	baseURL    string
	org        string
	dryRun     bool
	maintainer string
)

func init() {
	// parse flags
	flag.StringVar(&token, "token", os.Getenv("GITHUB_TOKEN"), "required: GitHub API token (or env var GITHUB_TOKEN)")
	flag.StringVar(&baseURL, "url", "", "optional: GitHub Enterprise URL (default: github.com)")
	flag.StringVar(&org, "org", "", "required: organization to include")
	flag.StringVar(&maintainer, "maintainer", "", "required: maintainer to set as the maintainer of the org")
	flag.BoolVar(&dryRun, "dry-run", false, "optional: do not change branch settings just print the changes that would occur (default: false)")
	flag.Parse()

	if token == "" {
		usageAndExit("GitHub token cannot be empty.", 1)
	}

	if org == "" {
		usageAndExit("no organization provided", 1)
	}

	if maintainer == "" {
		usageAndExit("no organization provided", 1)
	}

}

func main() {
	// intialize bot
	bot, err := shepherd.NewBot(baseURL, token, maintainer, org)
	if err != nil {
		logrus.Fatal(err)
		panic(err)
	}

	//Retreive repos that are owned by the org
	repos, err := bot.RetreiveRepos()
	if err != nil {
		logrus.Fatal(err)
		panic(err)
	}

	for _, repo := range repos {
		err = handleRepo(bot, repo)
		if err != nil {
			logrus.Fatal(err)
			panic(err)
		}
	}
}

// a function that will be applied to each repo on an org
func handleRepo(bot *shepherd.ShepardBot, repo *github.Repository) error {
	branchName := "master"

	b, err := bot.GetBranch(repo, branchName)
	if err != nil {
		return err
	}

	coExist, prExist, err := bot.CheckCodeOwners(repo, b)
	if err != nil {
		return err
	}

	if !coExist {
		fmt.Printf("[UPDATE REQUIRED] %s: A codeowner file was not found, a PR should be created\n", *repo.FullName)
		if dryRun {
			return nil
		}

		pr, err := bot.DoCreateCodeowners(repo, b)
		if err != nil {
			return err
		}

		fmt.Printf("[UPDATED] %s: A PR (%s) has been created to add CODEOWNERS file\n", *repo.FullName, pr.GetIssueURL())
		return nil // shouldn't go further at this point, since the PR has to be merged
	} else if prExist != nil {
		fmt.Printf("[MERGE REQUIRED] %s: CODEOWNERS file exists in a PR, please merge this before continuing\n", *repo.FullName)
		return nil // also shoudn't do anything since the PR hasn't been merged yet
	}
	fmt.Printf("[OK] %s: CODEOWNERS file already exists in repo\n", *repo.FullName)

	//Need to assign team to the repo even its in the org to be a "maintainer"
	repoManagement, err := bot.CheckTeamRepoManagement(repo)

	if err != nil {
		return err
	}

	if repoManagement {
		fmt.Printf("[OK] %s: is already managed by %s\n", *repo.FullName, maintainer)
	} else {
		fmt.Printf("[UPDATE REQUIRED] %s: needs to updated to be managed by %s\n", *repo.FullName, maintainer)
		err = bot.DoTeamRepoManagement(repo)
		if err != nil {
			return err
		}
		fmt.Printf("[OK] %s: is now managed by %s\n", *repo.FullName, maintainer)
	}

	// BRANCH PROTECTION + REQUIRED STATUS CHECKS
	branchProtect, err := bot.CheckProtectionBranch(repo, b)
	if err != nil {
		return err
	}

	if branchProtect {
		fmt.Printf("[OK] %s: %s is already protected\n", *repo.FullName, b.GetName())
		return nil
	}

	fmt.Printf("[UPDATE REQUIRED] %s: %s requires branch protection\n", *repo.FullName, b.GetName())

	// protect branch above
	err = bot.DoProtectBranch(repo, b)
	if err != nil {
		return err
	}

	fmt.Printf("[OK] %s: %s is now protected\n", *repo.FullName, b.GetName())
	return nil
}

func usageAndExit(message string, exitCode int) {
	if message != "" {
		fmt.Fprintf(os.Stderr, message)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(exitCode)
}
