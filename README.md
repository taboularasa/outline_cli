# Outline CLI

A command-line tool for interacting with the Outline API, enabling local editing of technical documents in Markdown format.

## Features

- Pull documents from Outline for local editing
- Push local changes back to Outline
- Compare local and remote document versions
- API key configuration via config file

## Installation

1. Ensure you have Go 1.21 or later installed
2. Clone this repository
3. Run 'just setup' to install dependencies
4. Run 'just build' to build the binary

## Configuration

Create a config file at ~/.outline-cli/config.json with your Outline API credentials:

{
    "api_key": "your-api-key",
    "outline_url": "https://your-outline-instance.com"
}

## Usage

Commands:
- outline pull [docID] : Fetch the latest version of a document
- outline push [docID] : Push local changes to Outline
- outline diff [docID] : Compare local and remote versions

Example:
1. Pull a document:
   outline pull abc123

2. Edit the document locally:
   edit abc123.md

3. Push changes back to Outline:
   outline push abc123

## Development

This project uses Just as a command runner. Available commands:

- just test : Run all tests
- just lint : Run linter
- just fmt : Format code
- just build : Build the binary
- just ci : Run all CI checks

## License

MIT License
