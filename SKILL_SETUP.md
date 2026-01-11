# Setting Up the Fave Skill for Claude Code

This guide explains how to install the Fave bookmark manager skill with Claude Code via the Plugin Marketplace.

## Quick Start

1. **Install the `fave` CLI** (if not already installed):
   ```bash
   cd ~/Repos/github.com/t-eckert/fave
   go build -o ~/go/bin/fave
   ```

2. **Configure the Fave client:**
   ```bash
   mkdir -p ~/.config/fave
   cat > ~/.config/fave/client.json <<EOF
   {
     "host": "https://your-homelab-fave-instance.com",
     "password": "your-auth-password"
   }
   EOF
   ```

3. **Test the connection:**
   ```bash
   fave health
   ```
   Should return: "Server is healthy"

4. **Install the skill via Claude Code Plugin Marketplace:**

   In a Claude Code conversation:

   ```
   /plugin marketplace add t-eckert/fave
   /plugin install fave@fave-marketplace
   ```

   That's it! The skill is now available in all conversations.

5. **Use the skill:**

   Natural language:
   - "Add this to my faves" - when discussing a URL
   - "Search my bookmarks for golang"
   - "Show me all my bookmarks"

   Explicit commands:
   - `/fave add <url>`
   - `/fave search <query>`
   - `/fave list`
   - `/fave get <id>`
   - `/fave update <id>`
   - `/fave delete <id>`
   - `/fave open <id>`

## Alternative: Local Development Installation

If you're developing the skill locally, you can install from a local path:

```
/plugin marketplace add /path/to/fave
/plugin install fave@fave-marketplace
```

## Example Usage

**Adding a bookmark with AI analysis:**
```
You: I found this great article: https://go.dev/blog/context
You: Add this to my faves

Claude: I'll add that bookmark. Let me analyze the content first.
[Fetches and analyzes the page]
Claude: Here's what I found:
- Name: Go Concurrency Patterns: Context
- Description: The context package provides a standard way...
- Tags: golang, concurrency, context, programming, backend

Would you like to modify anything before I save it?

You: Looks good!

Claude: [Runs: fave add "Go Concurrency Patterns: Context" "https://go.dev/blog/context" -d "..." -t golang -t concurrency -t context -t programming -t backend]
Bookmark added with ID: 42
```

**Searching bookmarks:**
```
You: Find my bookmarks about golang

Claude: [Runs: fave search "golang"]
Found 3 bookmarks matching "golang":

42: Go Concurrency Patterns: Context (matched: tags)
58: Effective Go (matched: name)
71: Writing Web Applications in Go (matched: description)

Would you like to see details for any of these?
```

## Features

- **Intelligent content analysis**: Uses WebFetch to automatically extract titles, descriptions, and generate relevant tags
- **Natural language**: Say "add this to my faves" instead of memorizing commands
- **Context aware**: Finds URLs in conversation history
- **Full CRUD operations**: Add, list, search, get, update, delete, and open bookmarks
- **Uses your local CLI**: Respects your configuration and handles auth automatically

## Troubleshooting

**Skill not activating:**
- Ensure the skill file is in the correct Claude Code skills directory
- Check that Claude Code can access the skill file
- Try using explicit `/fave` commands first

**Connection errors:**
- Test with: `fave health`
- Check configuration: `cat ~/.config/fave/client.json`
- Verify server is accessible (VPN, Tailscale, firewall, etc.)

**CLI not found:**
- Verify installation: `which fave`
- Ensure `~/go/bin` is in your PATH
- Check: `echo $PATH | grep "go/bin"`

## Next Steps

Once the skill is working:

1. Start using natural language to bookmark interesting articles
2. Organize bookmarks with meaningful tags during the AI analysis phase
3. Use search to find bookmarks quickly
4. Build up a well-organized collection of resources

The AI analysis feature is particularly powerful for:
- Automatically categorizing content with relevant tags
- Generating concise descriptions from long articles
- Maintaining consistent bookmark metadata

Enjoy using Fave with Claude Code!
