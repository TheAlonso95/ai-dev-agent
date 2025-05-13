# 🧠 AI Dev Agent

**AI Dev Agent** is a CLI tool that helps developers rapidly bootstrap new projects on GitHub using AI.

Just describe your idea, and the tool will:
- ✨ Create a new GitHub repository
- ✅ Generate an initial README and file structure (coming soon)
- 🧾 Break down your idea into development tasks
- 🐙 Create GitHub issues for each task

---

## 🚀 Example

```bash
aiagent init "Build a Pomodoro Timer web app using Go and JavaScript"
```

This command will:
- Create a GitHub repo like `pomodoro-timer`
- Use GPT-3.5 / Gemini to break the idea into actionable tasks
- Create issues in the repo like:
  - Setup the project structure
  - Create the timer UI
  - Add settings for custom durations

---

## ✨ Why This Project Exists

As a developer, it's easy to lose momentum when starting something new. Repeating boilerplate tasks (repo creation, file scaffolding, issue tracking) slows down creativity.

**AI Dev Agent** automates the boring stuff, so you can focus on writing code and shipping.

It also aims to:
- 💡 Inspire new devs by showing how to structure tasks
- 🤖 Serve as the foundation for future autonomous dev agents
- 🌍 Be an open-source showcase of Go + AI + GitHub automation

---

## 🛠️ Technologies Used

| Feature | Tech |
|--------|------|
| 🧠 AI Planner | [OpenAI GPT-3.5](https://platform.openai.com) or [Gemini API](https://ai.google.dev/) |
| 🐙 GitHub integration | [GitHub REST API v3](https://docs.github.com/en/rest) |
| ⚙️ CLI framework | [Cobra](https://github.com/spf13/cobra) |
| 🔐 Config | [godotenv](https://github.com/joho/godotenv) |

---

## 📂 Project Structure

```
ai-dev-agent/
├── cmd/             # CLI entrypoints (cobra commands)
│   └── init.go
├── internal/
│   ├── github/      # GitHub repo + issue creation
│   ├── openai/      # OpenAI/Gemini integration
│   ├── tasks/       # Task model and transformation logic
│   └── config/      # Env config loader
├── .env             # (Ignored) Contains API keys
├── .gitignore
├── go.mod
├── Makefile         # Common development commands
├── README.md
```

---

## 🧪 Setup & Usage

### 1. Clone the repo
```bash
git clone https://github.com/YOUR_USERNAME/ai-dev-agent
cd ai-dev-agent
```

### 2. Create a `.env` file
```env
GITHUB_TOKEN=ghp_xxxxxxxxxxxxx
OPENAI_API_KEY=sk-xxxxxxxxxxxx
GITHUB_USERNAME=your-github-username
```

> ✅ You can also set up Gemini if you prefer:
```env
GEMINI_API_KEY=your-gemini-api-key
```

### 3. Run the CLI
```bash
# Direct method
go run main.go init "Your app idea here"

# Using Makefile
make run ARGS="init 'Your app idea here'"
# OR for initialization specifically
make init ARGS="'Your app idea here'"
```

### 4. Using the Makefile
This project includes a Makefile to simplify common development tasks:

```bash
# Build the application
make build

# Run tests
make test

# Format code
make fmt

# Run linter
make lint

# Clean build artifacts
make clean

# Show all available commands
make help
```

---

## 🔮 Roadmap

- [x] `init` command with repo + issue generation
- [ ] AI Dev Agent: writes code based on issues
- [ ] QA Agent: browser tests via Playwright or Puppeteer
- [ ] Add GitHub Actions support
- [ ] OpenAI/Gemini selector in CLI
- [ ] Add templates (Go, Next.js, etc.)

---

## 🤝 Contributing

This is an early-stage tool but open to collaboration!
- Submit PRs
- Report issues
- Request features

---

## 📄 License

MIT — free to use, fork, and build upon.

---

> Made with Go, GitHub, and ambition. ✨
