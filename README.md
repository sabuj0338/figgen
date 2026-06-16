# Figgen: Figma to Next.js AI Agent

Figgen is an advanced, stateful CLI tool that automatically converts Figma designs into production-ready Next.js repositories using Google's Gemini AI. It acts as a multi-agent system, extracting design nodes, planning the React component architecture, and writing clean, structured code into a target boilerplate.

## 🌟 Features

- **Stateful Execution:** Uses an agentic `plan` and `run` workflow. If the AI errors out or hits a rate limit, you don't lose your progress. You can resume generation right where it left off.
- **Direct Figma Integration:** Uses the Figma REST API to extract raw design trees automatically—no manual exporting required.
- **Smart Architecture:** Analyzes your design to determine what needs to be a reusable component, a shadcn/ui generic component, or a full page.
- **Automated Boilerplating:** Clones your specified Next.js repository as the target output directory automatically before generating code.
- **Strict Engineering Constraints:** Enforces coding standards based on your `figgen.config.yaml` and internal AI rules.

## 🛠 Prerequisites

Copy the provided `.env.example` file to create your `.env` file in the root of the project:

```bash
cp .env.example .env
```

Then, open `.env` and add your API keys:

```env
# Required for Gemini models
GEMINI_API_KEY="your_google_ai_studio_key"

# Required for OpenAI models
OPENAI_API_KEY="your_openai_key"

# Required for Anthropic models
ANTHROPIC_API_KEY="your_anthropic_key"

# (Optional) If you are using Ollama, no API key is required. Make sure Ollama is running on your machine.
OLLAMA_HOST="http://localhost:11434"

# (Optional) Set your default LLM provider ("gemini", "openai", "anthropic", "ollama")
DEFAULT_PROVIDER="gemini"

# (Optional) Set your default model (e.g., "gpt-4o", "claude-3-5-sonnet-20240620", "qwen2.5-coder:1.5b")
DEFAULT_MODEL=""

# Required to fetch design data. Create a Personal Access Token in your Figma account settings.
FIGMA_TOKEN="your_figma_personal_access_token"
```

## 🚀 How to Run

### Step 1: Build the CLI

Compile the Go program into a binary.

```bash
go build -o figgen
```

### Step 2: Configure

Ensure your `figgen.config.yaml` is present in the root directory. This tells the AI what framework you are using and what boilerplate repository to clone.

### Step 3: Plan the Architecture

Run the `plan` command. This will:

1. Clone the boilerplate to `./out`.
2. Extract the design from your Figma URL.
3. Run the AI Planner to map the architecture.
4. Save the execution state to `./out/.figgen/tasks.json` and a human-readable `./out/.figgen/tasks.md`.

```bash
# Step 1: Clone repo, extract figma, and generate tasks.md
./figgen plan

# Use a different LLM provider:
./figgen plan --figma "<url>" --provider anthropic --model claude-3-5-sonnet-20240620
```

### Step 4: Execute the Tasks

Run the `run` command to start the AI Coder. It will read the `tasks.json` state, pick the first uncompleted task, generate the React code, and write it to the filesystem (e.g., `src/components/ui/Button.tsx`).

```bash
# Execute a single task (great for testing)
./figgen run

# Execute all pending tasks continuously
./figgen run --all

# Run code generation entirely offline with a local model!
./figgen run --all --provider ollama --model qwen2.5-coder:1.5b
```

> **Tip:** Open `./out/.figgen/tasks.md` in your editor while running `./figgen run --all` to watch the tasks get checked off in real time!

## 📂 Project Structure

- `cmd/`: Cobra CLI commands (`plan` and `run`).
- `internal/agents/`: Gemini integration (`planner.go`, `coder.go`).
- `internal/config/`: Configuration parsing.
- `internal/figma/`: Figma REST API extraction.
- `internal/filesystem/`: Writes generated code to the local target directories.
- `internal/github/`: Auto-clones Next.js boilerplates.
- `internal/state/`: Manages the iterative `tasks.json` execution tracker.
