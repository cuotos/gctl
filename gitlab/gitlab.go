package gitlab

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	gitlab "github.com/xanzy/go-gitlab"
	"golang.org/x/sync/semaphore"
	git "gopkg.in/src-d/go-git.v4"
	http "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

// Client represents a gitlab client
type Client struct{ gl *gitlab.Client }

// New returns a new instance of a client
func New(token string) *Client {
	return &Client{gitlab.NewClient(nil, token)}
}

// ListSubGroups returns a list of groups under the given group name
func (c *Client) ListSubGroups(group string) []*gitlab.Group {
	groups, _, err := c.gl.Groups.ListSubgroups(group, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	return groups
}

// ListProjects lists all projects you're member of
func (c *Client) ListProjects() ([]*gitlab.Project, error) {
	projectOps := gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 20,
			Page:    1,
		},
	}

	// Only pull back Projects you're a member of
	membership := true
	projectOps.Membership = &membership

	var projects []*gitlab.Project
	for {
		ps, resp, err := c.gl.Projects.ListProjects(&projectOps)
		if err != nil {
			log.Fatal(err)
		}
		// List all the projects we've found so far.
		for _, p := range ps {
			projects = append(projects, p)
		}

		// Exit the loop when we've seen all pages.
		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		// Update the page number to get the next page.
		projectOps.Page = resp.NextPage
	}

	return projects, nil
}

// ListGroupProjects will list all projects for a group id
func (c *Client) ListGroupProjects(group string) ([]*gitlab.Project, error) {
	projects, err := c.listGroupProjects(group)
	if err != nil {
		return projects, err
	}
	groups := c.ListSubGroups(group)

	wg := &sync.WaitGroup{}
	wg.Add(len(groups))
	wC := make(chan struct{})

	projectsC := make(chan []*gitlab.Project, 1)
	defer close(projectsC)
	errC := make(chan error, 1)
	defer close(errC)

	for _, g := range groups {
		go func(g *gitlab.Group, wg *sync.WaitGroup, c *Client) {
			ps, err := c.ListGroupProjects(g.FullPath)
			if err != nil {
				errC <- err
			} else {
				projectsC <- ps
			}
			wg.Done()
		}(g, wg, c)
	}

	go func() {
		wg.Wait()
		close(wC)
	}()

	for {
		select {
		case ps := <-projectsC:
			projects = append(projects, ps...)
		case err = <-errC:
		case <-wC:
			return projects, err
		}
	}
}

// Clone a list of projects from Gitlab
func (c *Client) Clone(directory string, accessToken string, projects []*gitlab.Project) ([]*git.Repository, []error) {
	ctx := context.Background()
	ctxWithCancel, cancel := context.WithCancel(ctx)

	defer cancel()

	var repos []*git.Repository
	var errors []error

	wg := &sync.WaitGroup{}
	wg.Add(len(projects))
	wC := make(chan struct{})

	var sem = semaphore.NewWeighted(int64(10))
	projectsC := make(chan *git.Repository)
	defer close(projectsC)
	errC := make(chan error, 1)
	defer close(errC)

	for _, project := range projects {
		go func(p *gitlab.Project, wg *sync.WaitGroup, c *Client) {
			sem.Acquire(ctxWithCancel, 1)
			defer sem.Release(1)
			path := filepath.Clean(directory) + "/" + p.PathWithNamespace
			log.Println("Cloning: " + path)
			r, err := git.PlainClone(path, false, &git.CloneOptions{
				Auth: &http.BasicAuth{
					Username: "git", // this can be anything except an empty string
					Password: accessToken,
				},
				URL:               "https://gitlab.com/" + p.PathWithNamespace,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			})
			if err != nil {
				errC <- err
				log.Printf("%s: %s", err, path)
			} else {
				projectsC <- r
			}
			wg.Done()
		}(project, wg, c)
	}

	go func() {
		wg.Wait()
		close(wC)
	}()

	for {
		select {
		case repo := <-projectsC:
			repos = append(repos, repo)
		case err := <-errC:
			errors = append(errors, err)
		case <-wC:
			return repos, errors
		}
	}
}

func (c *Client) listGroupProjects(group string) ([]*gitlab.Project, error) {
	projectOps := gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 20,
			Page:    1,
		},
	}

	// Only pull back Projects you're a member of
	membership := true
	projectOps.Membership = &membership

	var projects []*gitlab.Project
	for {
		ps, resp, err := c.gl.Groups.ListGroupProjects(group, &projectOps)
		if err != nil {
			return projects, err
		}
		// List all the projects we've found so far.
		for _, p := range ps {
			projects = append(projects, p)
		}

		// Exit the loop when we've seen all pages.
		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		// Update the page number to get the next page.
		projectOps.Page = resp.NextPage
	}

	return projects, nil
}

// GetCurrentRepo the repo from the current directory
func (c *Client) GetCurrentRepo() (*git.Repository, error) {
	currDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	r, err := git.PlainOpen(currDir)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// GetProjectID returns a project ID from a term
func (c *Client) GetProjectID(path string) (*gitlab.Project, error) {
	project, _, err := c.gl.Projects.GetProject(path)
	if err != nil {
		return nil, err
	}

	return project, nil
}

// GetPipelines gets pipelines
func (c *Client) GetPipelines(id int) ([]*gitlab.Pipeline, error) {
	pipelineOpts := &gitlab.ListProjectPipelinesOptions{}
	pipelineList, _, err := c.gl.Pipelines.ListProjectPipelines(id, pipelineOpts)
	if err != nil {
		return nil, err
	}

	var pipelines []*gitlab.Pipeline
	for _, v := range pipelineList {
		pipeline, _, err := c.gl.Pipelines.GetPipeline(id, v.ID)
		if err != nil {
			return nil, err
		}
		pipelines = append(pipelines, pipeline)
	}

	return pipelines, nil
}

// GetPipelineJobs get's  all the jobs for the requested pipeline
func (c *Client) GetPipelineJobs(projectID, pipelineID int) ([]*gitlab.Job, error) {
	listJobsOpts := &gitlab.ListJobsOptions{}
	jobs, _, err := c.gl.Jobs.ListPipelineJobs(projectID, pipelineID, listJobsOpts)
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

// GetJobTrace gets the trace for a single job. Trace is the GitLab name for a log, obviously
func (c *Client) GetJobTrace(projectID, jobID int) (io.Reader, error) {
	trace, _, err := c.gl.Jobs.GetTraceFile(projectID, jobID)
	if err != nil {
		return nil, err
	}

	return trace, nil
}

// RunJob runs a job that needs to be manually triggered
func (c *Client) RunJob(projectID, jobID int) (*gitlab.Job, error) {
	job, _, err := c.gl.Jobs.PlayJob(projectID, jobID)
	if err != nil {
		return nil, err
	}

	return job, nil
}

// CancelJob cnacels a job
func (c *Client) CancelJob(projectID, jobID int) (*gitlab.Job, error) {
	job, _, err := c.gl.Jobs.CancelJob(projectID, jobID)
	if err != nil {
		return nil, err
	}

	return job, nil
}

// RetryJob retries a job
func (c *Client) RetryJob(projectID, jobID int) (*gitlab.Job, error) {
	job, _, err := c.gl.Jobs.RetryJob(projectID, jobID)
	if err != nil {
		return nil, err
	}

	return job, nil
}
