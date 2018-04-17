package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
	"github.com/srizzling/shepherd/shepherd"
)

var version = "master"

var (
	token      string
	baseURL    string
	org        string
	dryRun     bool
	maintainer string
	pbranch    string

	vrsn bool
)

const (
	// BANNER is what is printed for help/info output.
	BANNER = `
____     _   _  U _____ u  ____    _   _  U _____ u   ____     ____
/ __"| u |'| |'| \| ___"|/U|  _"\ u|'| |'| \| ___"|/U |  _"\ u |  _"\
<\___ \/ /| |_| |\ |  _|"  \| |_) |/| |_| |\ |  _|"   \| |_) |//| | | |
u___) | U|  _  |u | |___   |  __/ U|  _  |u | |___    |  _ <  U| |_| |\
|____/>> |_| |_|  |_____|  |_|     |_| |_|  |_____|   |_| \_\  |____/ u
 )(  (__)//   \\  <<   >>  ||>>_   //   \\  <<   >>   //   \\_  |||_
(__)    (_") ("_)(__) (__)(__)__) (_") ("_)(__) (__) (__)  (__)(__)_)

ensures your GitHub repositories are herded like sheep
Version: %s
developed with <3 by Sriram Venkatesh

`
)

func init() {
	// parse flags
	flag.StringVar(&token, "token", os.Getenv("GITHUB_TOKEN"), "required: GitHub API token (or env var GITHUB_TOKEN)")
	flag.StringVar(&org, "org", "", "required: organization to look through")
	flag.StringVar(&pbranch, "branch", "master", "branch to protect")

	flag.StringVar(&baseURL, "url", "", "optional: GitHub Enterprise URL")
	flag.StringVar(&maintainer, "maintainer", "", "required: team to set as CODEOWNERS")
	flag.BoolVar(&dryRun, "dryrun", false, "optional: do not change branch settings just print the changes that would occur")

	flag.BoolVar(&vrsn, "version", false, "optional: print version and exit")

	// Exit safely when version is used
	if vrsn {
		fmt.Printf(BANNER, version)
		os.Exit(0)
	}

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(BANNER, version))
		flag.PrintDefaults()
	}
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
	b, err := bot.GetBranch(repo, pbranch)
	if err != nil {
		return err
	}

	coExist, prExist, err := bot.CheckCodeOwners(repo, b)
	if err != nil {
		return err
	}

	if !coExist {
		fmt.Printf("[UPDATE REQUIRED] %s: A codeowner file was not found, a PR should be created\n", *repo.FullName)

		if !dryRun {
			pr, err := bot.DoCreateCodeowners(repo, b)
			if err != nil {
				return err
			}
			fmt.Printf("[UPDATED] %s: A PR (%s) has been created to add CODEOWNERS file\n", *repo.FullName, pr.GetIssueURL())
		}

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

		if !dryRun {
			err = bot.DoTeamRepoManagement(repo)
			if err != nil {
				return err
			}
			fmt.Printf("[OK] %s: is now managed by %s\n", *repo.FullName, maintainer)
		}
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
	if !dryRun {
		err = bot.DoProtectBranch(repo, b)
		if err != nil {
			return err
		}

		fmt.Printf("[OK] %s: %s is now protected\n", *repo.FullName, b.GetName())
	}
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
