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

We use the Model Context Protocol (MCP) to securely connect to Figma and extract the deep layout and semantic context of a specific design node directly via its URL.

1. **Get the Figma URL:** Open your Figma file, select the specific component, frame, or page you want to generate, and copy its URL from the browser (it must contain a `?node-id=...` parameter).
2. **Run the Planner:** In your terminal, run the `plan` command. This will clone your configured Next.js boilerplate to `./out`, start the local MCP server to fetch the design, and run the AI Architecture Planner.

```bash
# Start the planner with your specific Figma node URL
./figgen plan --figma "https://www.figma.com/design/your_file_id/File-Name?node-id=123-456"
```

3. The Go server will fetch the node, parse the design tree, run the AI Planner, and save the execution state to `./out/.figgen/tasks.json` and `./out/.figgen/tasks.md`. It also saves the raw Figma data to `./out/.figgen/figma_context.json` for code generation.

### Alternative Step 3: Plan the Architecture (Using the Local Plugin)

We also provide a local WebSocket bridge to bypass Figma's cloud API rate limits entirely if you prefer not to use the URL-based MCP method.

1. **Install the Plugin (One-time):** Open Figma > Plugins > Development > Import plugin from manifest. Select the `manifest.json` inside the `./figma-plugin` folder of this repository.
2. **Start the Listener:** In your terminal, run the `listen` command. This will clone your boilerplate to `./out` and wait for data.

```bash
# Start the local WebSocket server
./figgen listen
```

3. **Export from Figma:** Select a component or page in your Figma canvas, open the "Figgen Exporter" plugin, and click **Export**.
4. The Go server will instantly receive the data, run the AI Planner, and save the execution state to `./out/.figgen/tasks.json` and `./out/.figgen/tasks.md`.

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
