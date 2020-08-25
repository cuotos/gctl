package cmd

import (
	"fmt"
	"log"
	"os"

	"gctl/gitlab"

	"gctl/tui"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gctl",
	Short: "Interact with git hosting services",
	Long:  "Several tools for interacting with git hosting services like creating, listing and navigating repositories",
}

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Show a tree of projects",
	Long:  `Show a tree of projects on the remote hosting service`,
	Run: func(cmd *cobra.Command, args []string) {
		accessToken, err := cmd.Flags().GetString("gitlab_access_token")
		if err != nil {
			log.Fatalln(err)
		}
		remote := gitlab.New(accessToken)
		group, err := cmd.Flags().GetString("group")
		if err != nil {
			log.Fatalln(err)
		}
		projects, err := remote.ListGroupProjects(group)
		if err != nil {
			log.Fatalln(err)
		}
		for _, project := range projects {
			fmt.Println(project.PathWithNamespace)
		}
	},
}

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone a number of projects under a group",
	Long: `
	Clone the projects under a gitlab group into a specific
	directory structure organised by the groups/subgroups.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		accessToken, err := cmd.Flags().GetString("gitlab_access_token")
		if err != nil {
			log.Fatalln(err)
		}
		directory, err := cmd.Flags().GetString("directory")
		if err != nil {
			log.Fatalln(err)
		}
		remote := gitlab.New(accessToken)
		group, err := cmd.Flags().GetString("group")
		if err != nil {
			log.Fatalln(err)
		}
		projects, err := remote.ListGroupProjects(group)
		if err != nil {
			log.Fatalln(err)
		}

		_, errors := remote.Clone(directory, accessToken, projects)
		if errors != nil {
			log.Println("Errors when cloning repos")
		}
	},
}

var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "View and control pipelines for a project",
	Long: `
	Pipeline uses a tui to view and control pipelines
	`,
	Run: func(cmd *cobra.Command, args []string) {
		accessToken, err := cmd.Flags().GetString("gitlab_access_token")
		if err != nil {
			log.Fatalln(err)
		}
		remote := gitlab.New(accessToken)

		namespace, err := cmd.Flags().GetString("namespace")
		if err != nil {
			log.Fatalln(err)
		}

		project, err := remote.GetProjectID(namespace)
		if err != nil {
			log.Fatalln(err)
		}

		tui.New(remote, namespace, project.ID)
	},
}

func init() {
	// Root flags
	rootFlags := rootCmd.PersistentFlags()
	rootFlags.String("gitlab_access_token", os.Getenv("GITLAB_ACCESS_TOKEN"), "An API access token from GitLab")

	// Tree flags
	treeFlags := treeCmd.Flags()
	treeFlags.StringP("group", "g", os.Getenv("GITLAB_GROUP"), "The root group")
	treeFlags.BoolP("cached", "c", false, "Only fetch information from the cache")

	// Clone flags
	cloneFlags := cloneCmd.Flags()
	cloneFlags.StringP("group", "g", os.Getenv("GITLAB_GROUP"), "The group with all the projects")
	cloneFlags.StringP("directory", "d", "/tmp/gctl/", "The directory to clone the repos too")

	// Pipeline flags
	pipelineFlags := pipelineCmd.Flags()
	pipelineFlags.StringP("namespace", "n", "", "The namespace to view the pipelines of(e.g. username/repo")

	// Add commands
	rootCmd.AddCommand(treeCmd)
	rootCmd.AddCommand(cloneCmd)
	rootCmd.AddCommand(pipelineCmd)
}

// Execute will execute the command line tool
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
