<p align="center">

  <p align="center">
  [![Go Report Card](https://goreportcard.com/badge/github.com/srizzling/shepherd)](https://goreportcard.com/report/github.com/srizzling/shepherd)

  </p>
</p>

# shepherd

## ToC

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


