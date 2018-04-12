 <p align="center">
    <a href="https://goreportcard.com/report/github.com/srizzling/shepherd"><img src="https://goreportcard.com/badge/github.com/srizzling/shepherd" /></a>
  </p>

## shepherd

## Table of Contents

    - [ToC](#toc)
    - [Introduction](#introduction)
    - [Features](#features)
    - [Usage](#usage)
    - [Deployment](#deployment)

## Introduction

`shepherd` is a useful cli for GitHub orgs who have multiple repositories, and want to ensure that a main "maintainer" team owns the repositories within that org. This project is heavily inspired by [pepper](https://github.com/genuinetools/pepper).


## Features

`shepherd` has the following features to herd your org repositories to be the same, like sheep:

- `shepherd` will check for and create a CODEOWNER file (by creating a PR) into your protected branch. The created CODEOWNER file depends on the "maintainer" team configuration.
- `shepherd` will set your specified branch (default: master) to be protected
- `shepherd` will ensure that your protected branch will need required reviews from CODEOWNERS configured by the PR mentioned above

It is useful to note that `shepherd` will not:

- configure status checks on your repository. This is because status checks are unique, or different per repo.
- overwrite an existing CODEOWNERS file. This is because shepherd gives you the flexibility to configure multiple CODEOWNERS on different code paths (without adding complexity to the tool)

## Usage

```
____     _   _  U _____ u  ____    _   _  U _____ u   ____     ____
/ __"| u |'| |'| \| ___"|/U|  _"\ u|'| |'| \| ___"|/U |  _"\ u |  _"\
<\___ \/ /| |_| |\ |  _|"  \| |_) |/| |_| |\ |  _|"   \| |_) |//| | | |
u___) | U|  _  |u | |___   |  __/ U|  _  |u | |___    |  _ <  U| |_| |\
|____/>> |_| |_|  |_____|  |_|     |_| |_|  |_____|   |_| \_\  |____/ u
 )(  (__)//   \\  <<   >>  ||>>_   //   \\  <<   >>   //   \\_  |||_
(__)    (_") ("_)(__) (__)(__)__) (_") ("_)(__) (__) (__)  (__)(__)_)

ensures your GitHub repositories are herded like sheep
Version: master
developed with <3 by Sriram Venkatesh

  -branch string
    	optional: branch to protect (default: master) (default "master")
  -debug
    	optional: run in debug mode
  -dryrun
    	optional: do not change branch settings just print the changes that would occur (default: false)
  -maintainer string
    	required: team to set as CODEOWNERS
  -org string
    	required: organization to look through
  -token string
    	required: GitHub API token (or env var GITHUB_TOKEN)
  -url string
    	optional: GitHub Enterprise URL (default: github.com)
  -version
    	optional: print version and exit
```


