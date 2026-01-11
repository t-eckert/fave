# Fave Plugin Marketplace

This repository provides a Claude Code Plugin Marketplace for the **Fave bookmark manager** skill.

## What is Fave?

Fave is a bookmark manager with a client-server architecture written in Go. This plugin enables Claude Code to interact with your Fave instance to intelligently manage bookmarks.

## Features

- **Intelligent content analysis**: Automatically extracts titles, descriptions, and generates relevant tags from URLs
- **Natural language commands**: Say "add this to my faves" instead of memorizing commands
- **Context aware**: Finds URLs in conversation history
- **Full CRUD operations**: Add, list, search, get, update, delete, and open bookmarks
- **Uses your local CLI**: Respects your configuration and handles auth automatically

## Installation

### 1. Install the Fave CLI

First, ensure you have the Fave CLI installed:

```bash
# Clone the repository
git clone https://github.com/t-eckert/fave.git
cd fave

# Build and install
go build -o ~/go/bin/fave
```

Verify installation:
```bash
which fave
# Should return: /Users/yourusername/go/bin/fave
```

### 2. Configure Your Fave Client

Create the configuration file:

```bash
mkdir -p ~/.config/fave
cat > ~/.config/fave/client.json <<EOF
{
  "host": "https://your-fave-server.com",
  "password": "your-auth-password"
}
EOF
```

Test the connection:
```bash
fave health
# Should return: Server is healthy
```

### 3. Add the Marketplace to Claude Code

In a Claude Code conversation, run:

```
/plugin marketplace add t-eckert/fave
```

### 4. Install the Fave Plugin

```
/plugin install fave@fave-marketplace
```

That's it! The skill is now available in all your Claude Code conversations.

## Usage

### Natural Language Commands

Claude will automatically activate the Fave skill when you mention bookmarks:

- **"Add this to my faves"** - When discussing a URL
- **"Search my bookmarks for golang"** - Search bookmarks
- **"Show me all my bookmarks"** - List all bookmarks
- **"Open bookmark 42"** - Open a specific bookmark

### Explicit Commands

You can also use explicit slash commands:

```
/fave add <url>              Add a bookmark with AI analysis
/fave search <query>         Search bookmarks
/fave list                   List all bookmarks
/fave get <id>               Get bookmark details
/fave update <id>            Update a bookmark
/fave delete <id>            Delete a bookmark
/fave open <id>              Open bookmark in browser
```

### Example Workflow

```
You: I found this great article: https://go.dev/blog/context
You: Add this to my faves

Claude: I'll add that bookmark. Let me analyze the content first.
[Fetches and analyzes the page]

Claude: Here's what I found:
- Name: Go Concurrency Patterns: Context
- Description: The context package provides a standard way to solve...
- Tags: golang, concurrency, context, programming, backend

Would you like to modify anything before I save it?

You: Looks good!

Claude: Bookmark added with ID: 42
```

## Available Operations

- **Add**: Intelligently analyze URLs and create bookmarks with auto-generated descriptions and tags
- **Search**: Find bookmarks using regex pattern matching
- **List**: View all your bookmarks
- **Get**: Retrieve details for a specific bookmark
- **Update**: Modify existing bookmarks
- **Delete**: Remove bookmarks (with confirmation)
- **Open**: Launch bookmarks in your default browser

## Configuration

The plugin uses the Fave CLI, which reads configuration from:

1. **CLI flags** (highest priority): `--host`, `--password`
2. **Environment variables**: `FAVE_HOST`, `FAVE_PASSWORD`
3. **Config file**: `~/.config/fave/client.json`
4. **Defaults** (lowest priority)

Example `~/.config/fave/client.json`:
```json
{
  "host": "https://fave.example.com",
  "password": "your-password-here"
}
```

## Troubleshooting

### Plugin Not Found

Ensure the marketplace is added:
```
/plugin marketplace list
```

If not listed, add it:
```
/plugin marketplace add t-eckert/fave
```

### Connection Errors

Test your Fave connection:
```bash
fave health
```

Check your configuration:
```bash
cat ~/.config/fave/client.json
```

Verify server is accessible (check VPN, Tailscale, firewall, etc.)

### CLI Not Found

Verify the Fave CLI is installed and in PATH:
```bash
which fave
echo $PATH | grep "go/bin"
```

If not in PATH, add `~/go/bin` to your PATH in `~/.zshrc` or `~/.bashrc`:
```bash
export PATH="$HOME/go/bin:$PATH"
```

## Development

To develop or modify the skill:

1. Clone this repository
2. Make changes to `plugins/fave/skills/fave/SKILL.md`
3. Test locally:
   ```
   /plugin marketplace add /path/to/fave
   /plugin install fave@fave-marketplace
   ```
4. Submit a pull request

## License

MIT - See LICENSE file for details

## Author

Thomas Eckert (thomas.eckert@hey.com)

## Links

- **Repository**: https://github.com/t-eckert/fave
- **Issues**: https://github.com/t-eckert/fave/issues
- **Fave Server**: Run your own instance with `fave serve`
