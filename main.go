package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"
)

const prefix = ".github/workflows/"

func realMain() error {
	token := flag.String("token", "", "token")
	flag.Parse()
	
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	runs, res, err := client.Actions.ListWorkflows(ctx, "containerd", "containerd", nil)
	if err != nil {
		return err
	}

	fmt.Printf("runs = %+v\nres = %+v\n", runs, res)

	for _, wf := range runs.Workflows {
		path := *wf.Path

		// Some stuff are internally treated as workflows.
		// Their paths are like "dynamic/pages/...".
		if ! strings.HasPrefix(path, ".github/workflows/") {
			continue
		}
		fmt.Printf("%s - %s\n", path[len(prefix):], *wf.Name)

		runs, _, err := client.Actions.ListWorkflowRunsByID(ctx, "containerd", "containerd", *wf.ID, &github.ListWorkflowRunsOptions{ExcludePullRequests:true})
		if err != nil {
			return err
		}
		for _, run := range runs.WorkflowRuns {
			if *run.Event == "pull_request" {
				continue
			}
			fmt.Printf("  %s %s %+v\n", *run.HeadBranch, *run.Conclusion, *run.Event)
			jobs, _, err := client.Actions.ListWorkflowJobs(ctx, "containerd", "containerd", *run.ID, nil)
			if err != nil {
				return err
			}
			for _, job := range jobs.Jobs {
				fmt.Printf("    %s %s\n", *job.Conclusion, *job.Name)
			}
		}
	}
	return nil
}


func main() {
	err := realMain()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghaa: %s\n", err)
		os.Exit(1)
	}
}
