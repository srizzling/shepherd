package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/srizzling/shepherd/shepherd"
)

var version = "master"

var (
	token         string
	vrsnFlag      bool
	configuration shepherd.Config
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

	// Intialize Viper
	viper.SetConfigName(".shepherd")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("shepherd")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match
	viper.BindEnv("GITHUB_TOKEN")

	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatalf("Error reading config file, %s", err)
		panic(err)
	}

	if err := viper.Unmarshal(&configuration); err != nil {
		logrus.Fatalf("Error marshalling config file, %s", err)
		panic(err)
	}

	token = configuration.GithubToken
	if token == "" {
		usageAndExit("Error! Github Token is required", 1)
	}
}

func main() {
	// intialize bot
	bot, err := shepherd.NewBot(configuration)
	if err != nil {
		logrus.Fatal(err)
		panic(err)
	}

	for repo, repoConfig := range bot.Repos {
		err = handleRepo(bot, repo, repoConfig)
		if err != nil {
			logrus.Fatal(err)
			panic(err)
		}
	}
}

// a function that will be applied to each repo on an org
func handleRepo(bot *shepherd.ShepardBot, repo *github.Repository, repoConfig shepherd.RepoConfig) error {
	b, err := bot.GetBranch(repo, repoConfig.ProtectedBranch)
	dryRun := configuration.DryRun
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
			pr, err := bot.DoCreateCodeowners(repo, b, repoConfig.GHMaintainer)
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
	repoManagement, err := bot.CheckTeamRepoManagement(repo, repoConfig.GHMaintainer)

	if err != nil {
		return err
	}

	if repoManagement {
		fmt.Printf("[OK] %s: is already managed by %s\n", *repo.FullName, repoConfig.Maintainer)
	} else {
		fmt.Printf("[UPDATE REQUIRED] %s: needs to updated to be managed by %s\n", *repo.FullName, repoConfig.Maintainer)

		if !dryRun {
			err = bot.DoTeamRepoManagement(repo, repoConfig.GHMaintainer)
			if err != nil {
				return err
			}
			fmt.Printf("[OK] %s: is now managed by %s\n", *repo.FullName, repoConfig.Maintainer)
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
