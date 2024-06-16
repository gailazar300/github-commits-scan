package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
)

// Pagination number of result per page
const (
	LastCommitDateEnvVar = "ENTRO_LAST_COMMIT_DATE"
	perPage              = 30
)

// Patterns to match AWS access keys and secret keys
var awsAccessKeyPattern = regexp.MustCompile(`AKIA[0-9A-Z]{16}`)

// TODO gives false positives
var awsSecretKeyPattern = regexp.MustCompile(`[0-9a-zA-Z/+]{40}`)

func main() {
	ctx := context.Background()
	owner, repo, token := parseCommandLineFlags()

	client := newGithubClient(ctx, token)
	opts := getCommitListOptions()
	for {
		commits, _, err := client.Repositories.ListCommits(ctx, owner, repo, opts)
		if err != nil {
			log.Fatalf("Error fetching commits: %v", err)
		}
		if len(commits) == 0 {
			break
		}

		// Iterate through each commit
		for _, commit := range commits {
			commitSHA := commit.GetSHA()
			fmt.Printf("Commit: %s\n", commitSHA)
			fmt.Printf("Committer: %s\n", commit.GetCommitter().GetLogin())

			// Get commit details
			commitDetail, _, err := client.Repositories.GetCommit(ctx, owner, repo, commitSHA, &github.ListOptions{PerPage: perPage})
			if err != nil {
				log.Fatalf("Error fetching commit details: %v", err)
			}
			var accessKeysRes []string
			var secretKeysRes []string
			for _, file := range commitDetail.Files {
				if file.GetFilename() == "" {
					continue
				}
				// Get file content
				fileContent, _, _, err := client.Repositories.GetContents(ctx, owner, repo, file.GetFilename(), &github.RepositoryContentGetOptions{Ref: commitSHA})
				if err != nil {
					log.Printf("Error fetching file content: %v", err)
				}

				var content string
				if fileContent != nil {
					// Check if the content is directly available as plain text
					content, err = fileContent.GetContent()
					if err != nil {
						log.Fatalf("Error getting content: %v", err)
					}
					if content == "" {
						log.Printf("Unsupported content encoding: %s", fileContent.GetEncoding())
						continue
					}
				}

				// Scan for AWS exposures and save results.
				accessKeys, secretKeys := scanForAWSSecrets(content)
				accessKeysRes = append(accessKeysRes, accessKeys...)
				secretKeysRes = append(secretKeysRes, secretKeys...)

				// In case of interruptions, save the last commit date
				err = os.Setenv(LastCommitDateEnvVar, commitDetail.Commit.Author.GetDate().Format(time.RFC3339))
				if err != nil {
					log.Printf("Failed to set last commit date: %v", err)
				}
			}
			fmt.Printf("accessKeys: %v\n", accessKeysRes)
			fmt.Printf("secretKeys: %s\n", secretKeysRes)
		}
	}
}

func getCommitListOptions() *github.CommitsListOptions {
	opts := &github.CommitsListOptions{
		ListOptions: github.ListOptions{PerPage: perPage},
	}
	lastCommitDate := os.Getenv(LastCommitDateEnvVar)
	if lastCommitDate != "" {
		parsedTime, err := time.Parse(time.RFC3339, lastCommitDate)
		if err != nil {
			log.Fatalf("Error parsing time: %v", err)
		}
		opts.Since = parsedTime
	}
	return opts
}

// New GitHub client initiation.
func newGithubClient(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

func parseCommandLineFlags() (string, string, string) {
	// Parse command-line flags
	var owner, token, repo string
	flag.StringVar(&owner, "owner", "", "GitHub repository owner")
	flag.StringVar(&token, "token", "", "GitHub access token")
	flag.StringVar(&repo, "repo", "", "GitHub repository name")
	flag.Parse()

	// Check if required flags are provided
	if owner == "" || token == "" || repo == "" {
		flag.Usage()
		os.Exit(1)
	}
	return owner, repo, token
}

// Scan content for AWS secrets
func scanForAWSSecrets(content string) (accessKeys, secretKeys []string) {
	accessKeys = awsAccessKeyPattern.FindAllString(content, -1)
	secretKeys = awsSecretKeyPattern.FindAllString(content, -1)
	return
}
