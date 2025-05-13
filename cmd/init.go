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

var (
	repoName string
	stack    string
)

var initCmd = &cobra.Command{
	Use:   "init [idea]",
	Short: "Initialize a new project and create GitHub tasks",
	Long: `This command creates a GitHub repo and uses AI to break your idea into dev tasks.
		Example:
  		aiagent init "Pomodoro timer web app" --name pomodoro-timer --stack "Go, SQLite, React"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		idea := args[0]
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		token := os.Getenv("GITHUB_TOKEN")
		openaiKey := os.Getenv("OPENAI_API_KEY")
		owner := os.Getenv("GITHUB_USERNAME")
		projectName := sanitizeRepoName(repoName)
		if projectName == "" {
			projectName = "ai-" + sanitizeRepoName(idea)
		}

		fmt.Println("Creating project with name:", repoName, projectName)
		err = github.CreateRepo(projectName, token)
		if err != nil {
			log.Fatal(err)
		}

		finalPrompt := fmt.Sprintf("Build '%s' using %s. Break it into actionable tasks as JSON...", idea, stack)
		tasks, err := openai.AskForTasks(finalPrompt, openaiKey)
		if err != nil {
			log.Fatal(err)
		}

		for _, task := range tasks {
			err := github.CreateIssue(owner, projectName, token, task)
			if err != nil {
				log.Println("Failed to create issue:", err)
			}
		}

		fmt.Println("✅ Project setup complete:", repoName)

		fmt.Println("📝 Generating README.md via AI...")
		readme, err := openai.GenerateReadme(projectName, idea, stack, openaiKey)
		if err != nil {
			log.Fatalf("Failed to generate README: %v", err)
		}

		file := github.File{
			Path:    "README.md",
			Content: readme,
		}

		// The repository is empty at this point, so we need to initialize it
		// CreateBranchAndCommit will internally handle the empty repository case
		err = github.CreateBranchAndCommit(projectName, []github.File{file}, token)
		if err != nil {
			log.Fatalf("Failed to commit README: %v", err)
		}

		fmt.Println("✅ README committed to main branch in repo:", projectName)
	},
}

func sanitizeRepoName(name string) string {
	// Replace spaces with dashes, lowercase, remove special chars (basic)
	return strings.ReplaceAll(strings.ToLower(name), " ", "-")
}

func init() {
	initCmd.Flags().StringVarP(&repoName, "name", "n", "", "Custom name for the GitHub repository")
	initCmd.Flags().StringVarP(&stack, "stack", "s", "", "Tech stack (e.g. 'Go, PostgreSQL, React')")
	rootCmd.AddCommand(initCmd)
}
