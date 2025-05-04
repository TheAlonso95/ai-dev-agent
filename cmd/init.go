package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/TheAlonso95/ai-dev-agent/internal/github"
	"github.com/TheAlonso95/ai-dev-agent/internal/openai"
)

var initCmd = &cobra.Command{
	Use:   "init [idea]",
	Short: "Initialize a new project and create GitHub tasks",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		idea := args[0]

		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		token := os.Getenv("GITHUB_TOKEN")
		openaiKey := os.Getenv("OPENAI_API_KEY")
		owner := os.Getenv("GITHUB_USERNAME")
		repoName := "ai-" + sanitizeRepoName(idea)

		err = github.CreateRepo(repoName, token)
		if err != nil {
			log.Fatal(err)
		}

		tasks, err := openai.AskForTasks("Break this project into development tasks as JSON: "+idea, openaiKey)
		if err != nil {
			log.Fatal(err)
		}

		for _, task := range tasks {
			err := github.CreateIssue(owner, repoName, token, task)
			if err != nil {
				log.Println("Failed to create issue:", err)
			}
		}

		fmt.Println("✅ Project setup complete:", repoName)
	},
}

func sanitizeRepoName(name string) string {
	// Replace spaces with dashes, lowercase, remove special chars (basic)
	return strings.ReplaceAll(strings.ToLower(name), " ", "-")
}

func init() {
	rootCmd.AddCommand(initCmd)
}
