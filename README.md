# cnabtool

A CLI utility for inspecting and manipulating [CNAB](https://cnab.io) (Cloud Native Application Bundle) artifacts stored in [OCI](https://opencontainers.org) (Open Container Initiative) container registries.

## Features

- **Fetch manifests** — retrieve the OCI index manifest of a CNAB project as formatted JSON
- **Inspect projects** — walk the full dependency graph of a CNAB project, resolving all component tags, uplinks, and downlinks
- **Delete projects** — safely remove a CNAB project from a registry, deleting leaf components before their parents
- **Credential-safe logging** — passwords and basic auth tokens are automatically redacted from all log output

## Installation

### From source

```bash
make build
cp bin/cnabtool /usr/local/bin
```

### Cross-compile

```bash
make dist
```

Produces binaries for `darwin-amd64`, `linux-amd64`, and `windows-amd64` in the `dist/` directory.

## Configuration

cnabtool loads configuration from three sources, merged in order of increasing priority:

1. **Config file** — searched for in `/etc/cnabtool/`, `$HOME/.cnabtool/`, and the current directory (`config.yaml`)
2. **Environment variables** — prefixed with `CNAB_` (e.g., `CNAB_USERNAME`, `CNAB_VERBOSITY`)
3. **CLI flags** — override all other settings

### Default config file

```yaml
credentials:
  username: "registry-user"
  password: "registry-password"
timeout: 10000
verbosity: 2
```

### CLI flags

| Flag | Short | Env | Description | Default |
|---|---|---|---|---|
| `--config` | `-c` | — | Path to config file | auto-discovered |
| `--verbosity` | `-v` | `CNAB_VERBOSITY` | Log level (0–4) | `2` |
| `--username` | `-u` | `CNAB_USERNAME` | Registry username | — |
| `--password` | `-p` | `CNAB_PASSWORD` | Registry password | — |
| `--timeout` | `-t` | `CNAB_TIMEOUT` | HTTP timeout in milliseconds | `10000` |

### Verbosity levels

| Level | Name | Output |
|---|---|---|
| `0` | Quiet | No output |
| `1` | Error | Errors only |
| `2` | Normal | Messages + errors |
| `3` | Info | Info + messages + errors |
| `4` | Debug | All output |

## Usage

```
cnabtool <command> [arguments]
```

### `version`

Print version information.

```bash
cnabtool version
# Output: Version: 0.1.1 (commit-hash)
```

### `content manifest`

Retrieve the OCI index manifest for a given registry reference and print it as pretty-printed JSON.

```bash
cnabtool content manifest registry.example.com/project/cnab:tag@sha256:abc123...
```

### `content inspect`

Fetch the manifest, walk all tags in the CNAB project, build the full dependency graph (uplinks/downlinks), and output a JSON report.

```bash
cnabtool content inspect registry.example.com/project/cnab:tag
```

**Flags:**

| Flag | Description | Default |
|---|---|---|
| `--raw` | Output full raw item data instead of a compact summary | `false` |

### `content delete`

Delete a CNAB project from the registry. Identifies all child tags referenced by only one parent (leaf nodes) and deletes them first, then removes the parent.

```bash
# Dry-run: show what would be deleted without making changes
cnabtool content delete registry.example.com/project/cnab:tag --dry-run

# Actually delete the project
cnabtool content delete registry.example.com/project/cnab:tag
```

**Flags:**

| Flag | Description | Default |
|---|---|---|
| `--dry-run` | Show items that would be deleted without performing deletions | `false` |

## How It Works

### Reference format

```
registry.example.com/repository/image:tag@sha256:digest
```

All three components (tag, digest) are optional. The parser handles:

- Tag + digest: `registry/repo:tag@sha256:abc`
- Digest only: `registry/repo@sha256:abc`
- Tag only: `registry/repo:tag`

### Inspection flow

1. Parse the registry reference into registry, repository, tag, and digest components
2. Fetch the top-level OCI index manifest via authenticated GET request
3. Parse the index and register each component by its media type and annotations
4. Iterate over all tags in the repository, fetching each manifest and resolving uplink/downlink chains
5. Mark any references as "lost" if they cannot be resolved
6. Output a JSON report (compact by default, full detail with `--raw`)

### Deletion strategy

The delete command uses a **leaf-first** approach:

1. Build the same dependency graph as inspect
2. Identify leaf nodes — items referenced by exactly one parent
3. Delete leaf nodes first, then their parents
4. HTTP 202 indicates success; other status codes are logged with the response body

## Architecture

```
cnabtool/
├── main.go                    Entry point
├── cmd/
│   ├── cli.go                 Cobra command tree + global flags
│   ├── content.go             content manifest/inspect/delete subcommands
│   └── version.go             version subcommand
├── pkg/
│   ├── client/
│   │   ├── client.go          OCI registry HTTP client (GET/DELETE)
│   │   └── client_test.go     ParseReference unit tests
│   ├── config/
│   │   └── config.go          Viper-based config (file/env/flags)
│   ├── content/
│   │   ├── manifest.go        GetManifest + ResponsePrettyPrint
│   │   ├── inspect.go         Dependency graph inspection
│   │   └── delete.go          Leaf-first deletion
│   ├── data/
│   │   └── data.go            Data models + global state
│   └── logging/
│       └── logging.go         Structured logging with credential redaction
├── Makefile                   Build, test, lint, dist targets
└── go.mod                     Go module definition
```

### Key packages

| Package | Responsibility |
|---|---|
| `cmd` | CLI command definitions using Cobra |
| `config` | Configuration loading via Viper (file → env → flags) |
| `client` | HTTP client for OCI registry interactions with Basic Auth and media type fallback |
| `content` | CNAB content operations: manifest retrieval, inspection, deletion |
| `data` | All data structures: `Config`, `RegIndex`, `ProjectList`, lookup maps |
| `logging` | Five-level structured logging; sensitive data redaction in all output |

## Development

```bash
make run              # Build and run
make test             # Run tests
make lint             # Run golangci-lint
make fmt              # Run gofumpt
make clean            # Remove bin/ and dist/
```

## License

See [LICENSE](LICENSE).
