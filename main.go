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

func listAllWorkflows(ctx context.Context, client *github.Client, owner, repos string) error {
	runs, _, err := client.Actions.ListWorkflows(ctx, owner, repos, nil)
	if err != nil {
		return err
	}

	for _, wf := range runs.Workflows {
		path := *wf.Path

		// Some stuff are internally treated as workflows.
		// Their paths are like "dynamic/pages/...".
		if !strings.HasPrefix(path, ".github/workflows/") {
			continue
		}

		runs, _, err := client.Actions.ListWorkflowRunsByID(ctx, owner, repos, *wf.ID,
			&github.ListWorkflowRunsOptions{ExcludePullRequests: true},
		)
		if err != nil {
			return err
		}

		var success, failure int
		for _, run := range runs.WorkflowRuns {
			if *run.Event == "pull_request" {
				continue
			}
			if run.GetConclusion() == "success" {
				success += 1
			} else {
				failure += 1
			}
		}
		if success + failure > 0 {
			fmt.Printf("%6.2f%% %s\n", (float64(success) / float64(success + failure)) * 100.0, wf.GetName())
		}
	}
	return nil
}

func realMain() error {
	token := flag.String("token", "", "token")
	flag.Parse()

	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return listAllWorkflows(ctx, client, "containerd", "containerd")
}

func main() {
	err := realMain()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghaa: %s\n", err)
		os.Exit(1)
	}
}
