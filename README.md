# ORAS MCP Server

## Usage

Configure the ORAS MCP Server to run inside [VS Code agent mode](https://code.visualstudio.com/docs/copilot/chat/chat-agent-mode) for registry-aware chats.

The server is implemented with the [`oras-go`](https://github.com/oras-project/oras-go) library and does not depend on the [`oras`](https://github.com/oras-project/oras) CLI.

### Setup with Docker

It is easier to use docker to run the ORAS MCP Server.

#### Setup for VS Code

Add the following code to `.vscode/mcp.json`:

```json
{
    "servers": {
        "oras-mcp-server": {
            "type": "stdio",
            "command": "docker",
            "args": [
                "run",
                "--rm",
                "-i",
                "ghcr.io/oras-project/oras-mcp:main",
                "serve"
            ]
        }
    }
}
```

### Authentication

If you need to access private registries through the server, log in ahead of time using either the ORAS CLI or Docker:

- `oras login <registry>`
- `docker login <registry>`

The server will reuse the credentials from your local credential store when available.

## Example Chats

Q: What platform does the image ghcr.io/oras-project/oras support?

A: The image `ghcr.io/oras-project/oras` supports the following platforms:

1. **Linux (amd64)**
2. **Linux (arm64)**
3. **Linux (arm, variant v7)**
4. **Linux (s390x)**
5. **Linux (ppc64le)**

Let me know if you need further details!

Q: What's the latest Azure Linux 3.0 image in MCR? Does it have a signature? What's the type of it?

A: The latest Azure Linux 3.0 image in the Microsoft Container Registry (MCR) is tagged as `3.0.20250402`. It has a signature, and the type of the signature is `application/vnd.cncf.notary.signature`.
