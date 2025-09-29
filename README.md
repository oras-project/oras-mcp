# ORAS MCP Server

[![Build Status](https://github.com/oras-project/oras-mcp/actions/workflows/build.yml/badge.svg?event=push)](https://github.com/oras-project/oras-mcp/actions/workflows/build.yml?query=workflow%3Abuild+event%3Apush)
[![codecov](https://codecov.io/gh/oras-project/oras-mcp/branch/main/graph/badge.svg)](https://codecov.io/gh/oras-project/oras-mcp)

<p style="text-align: left;">
<a href="https://oras.land/"><img src="https://oras.land/img/oras.svg" alt="banner" width="100px"></a>
</p>

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

##### Mount Docker credentials (Linux only)

Linux users can share `docker login` credentials by mounting the config file:

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
                "-v",
                "${env:HOME}/.docker/config.json:/root/.docker/config.json:ro",
                "ghcr.io/oras-project/oras-mcp:main",
                "serve"
            ]
        }
    }
}
```

Adjust the path if you keep credentials under `${env:DOCKER_CONFIG}`. The container understands only inline `auths` entries; helper-based configs (`credsStore`, `credHelpers`) won't work. On macOS/Windows, install the [released binary](#setup-from-released-binaries) instead.

### Setup from Released Binaries

1. Visit the [GitHub releases page](https://github.com/oras-project/oras-mcp/releases) and download the archive that matches your operating system and CPU architecture (`oras-mcp_<version>_<os>_<arch>.tar.gz` for Linux or macOS, `oras-mcp_<version>_windows_<arch>.zip` for Windows).
2. Extract the archive; the folder contains a single executable named `oras-mcp` (or `oras-mcp.exe` on Windows) and a copy of the project license.
3. Move the binary to a directory on your `PATH` (for example, `/usr/local/bin` on Linux/macOS or `%LOCALAPPDATA%\Programs\oras-mcp` on Windows) or reference it directly from its extracted location.
4. Run `oras-mcp serve --help` (or `oras-mcp.exe serve --help`) to confirm the binary works on your system.
5. To integrate with VS Code agent mode, update `.vscode/mcp.json` to point at the extracted binary:

   ```json
   {
       "servers": {
           "oras-mcp-server": {
               "type": "stdio",
               "command": "/absolute/path/to/oras-mcp",
               "args": [
                   "serve"
               ]
           }
       }
   }
   ```

   On Windows, set `"command": "C:/path/to/oras-mcp.exe"`.

### Authentication

`oras-mcp` reads credentials from the same stores used by the ORAS and Docker CLIs, but you need to expose those stores to the server process:

- **Released binary** – Run `oras login <registry>` or `docker login <registry>` on the host machine; the binary will pick up the cached credentials automatically.
- **Docker container** – On Linux you can mount your Docker config as shown in the [credential section](#mount-docker-credentials-linux-only); ensure the file contains inline `auths` entries. Docker Desktop (macOS/Windows) depends on keychain helpers, so use the released binary there.

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

## Release Process

Maintainers can find the full release steps in the [release checklist](./RELEASE_CHECKLIST.md).
