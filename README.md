# gctl

gctl is a command line utility to interact with GitLab on the command line built
with cobra.

## Authentication

Enter an API access token into an environment variable called `GITLAB_ACCESS_TOKEN`.

##  Subcommands

### Clone

The clone subcommand clones all repositories under a particular group.
Unfortunately, sets origins to be https for now.

### Tree

Tree prints a tree of all repos under a group

### Pipeline

The cherry on the cake, a terminal user interface for interacting with pipelines.

## Contributors

This repo was just open sourced so had to lose the git history a bit. Here's
our kudos for the contributors:

- [Dan Potepa](https://github.com/cuotos)
- [Zee](https://github.com/zeeraw)

