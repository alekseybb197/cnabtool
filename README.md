# cnabtool

A CLI utility for inspecting and manipulating [CNAB](https://cnab.io) (Cloud Native Application Bundle) artifacts stored in [OCI](https://opencontainers.org) (Open Container Initiative) container registries.

## Features

- **Fetch manifests** — retrieve the OCI index manifest of a CNAB project as formatted JSON
- **Inspect projects** — walk the full dependency graph of a CNAB project, resolving all component tags, uplinks, and downlinks (including untagged manifests)
- **Delete projects** — safely remove a CNAB project from a registry, deleting leaf components before their parents
- **Purge empty folders** — clean up empty "folders" in Artifactory after deletion with adaptive timeout detection
- **Credential-safe logging** — passwords and basic auth tokens are automatically redacted from all log output
- **Dry-run mode** — preview all operations without making changes

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

# Delete with folder cleanup (requires Artifactory)
cnabtool content delete registry.example.com/project/cnab:tag --purge

# Delete with explicit repo-key (when hostname differs from repo-key)
cnabtool content delete registry.example.com/project/cnab:tag --purge --repo-key my-repo-key
```

**Flags:**

| Flag | Description | Default |
|---|---|---|
| `--dry-run` | Show items that would be deleted without performing deletions | `false` |
| `--purge` | Remove empty parent folders via Artifactory API after delete | `false` |
| `--repo-key` | Artifactory repository key (auto-derived from hostname by default) | — |

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
5. Fetch untagged manifests by digest (config, invocation, component manifests)
6. Mark any references as "lost" if they cannot be resolved
7. Output a JSON report (compact by default, full detail with `--raw`)

### Deletion strategy

The delete command uses a **leaf-first** approach:

1. Build the same dependency graph as inspect
2. Identify leaf nodes — items referenced by exactly one parent
3. Delete by **digest** (not tag) to handle untagged components correctly
4. Skip items with `UpLinks > 1` (shared between parents)
5. Issue `DELETE /manifests/<digest>` for each unique digest
6. HTTP 202 indicates success; other status codes are logged with the response body

### Purge flow (`--purge` flag)

After deletion, `--purge` cleans up empty "folders" in Artifactory using the Artifactory REST API:

1. Determine `repo-key` (from `--repo-key` flag or auto-derived from hostname via `deriveRepoKey()`)
2. Start from `cl.Repository` path (e.g., `cnab/myapp/1.0.0/myapp`)
3. Loop: `GET /artifactory/api/storage/{repoKey}/{path}?list` → if `children` is empty, delete folder; else stop
4. Delete via `DELETE /artifactory/{repoKey}/{path}` with a dedicated 180-second timeout client
5. Move up with `path.Dir()` and repeat until a non-empty folder, root, or adaptive threshold is reached
6. **Adaptive termination:** if a DELETE takes >5× the average of previous deletions, the purge stops immediately — this prevents hanging on large parent directories

**Auto-derived repo-key:** `deriveRepoKey()` extracts the repo-key from hostname automatically (e.g., `registry.example.com` → `registry`). Port numbers are stripped (`host:port` → `host`). The `--repo-key` flag is only needed when the repo-key differs from the hostname.

**Dry-run transparency:** `--dry-run --purge` shows every folder check and potential deletion with `[dry-run] Purge: ...` messages.

## Usage Examples

```bash
# Get manifest as pretty JSON
cnabtool content manifest registry.example.com/project/cnab:tag

# Inspect CNAB project dependency graph
cnabtool content inspect registry.example.com/project/cnab:tag

# Delete CNAB project (dry-run)
cnabtool content delete registry.example.com/project/cnab:tag --dry-run

# Delete CNAB project with folder cleanup
cnabtool content delete registry.example.com/project/cnab:tag --purge

# Delete with explicit repo-key (when hostname differs)
cnabtool content delete registry.example.com/project/cnab:tag --purge --repo-key my-repo-key

# Verbose debug output
cnabtool content delete registry.example.com/project/cnab:tag --purge -v 4
```

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
│   │   ├── client.go          OCI registry HTTP client (GET/DELETE/WebRequestEx)
│   │   └── client_test.go     ParseReference, NewRegClient, FillResponse tests
│   ├── config/
│   │   ├── config.go          Viper-based config (file/env/flags)
│   │   └── config_test.go     Defaults, singleton, precedence tests
│   ├── content/
│   │   ├── manifest.go        GetManifest + ResponsePrettyPrint
│   │   ├── inspect.go         Dependency graph inspection + untagged fetch
│   │   ├── delete.go          Digest-based leaf-first deletion
│   │   └── purge.go           PurgeEmptyFolders (Artifactory API cleanup)
│   ├── data/
│   │   ├── data.go            Data models + global state (Gc, Sensitives, maps)
│   │   └── data_test.go       All types, global state, link management tests
│   └── logging/
│       ├── logging.go         5-level structured logging with credential redaction
│       └── logging_test.go    All log levels, masking, PrettyString tests
├── fixes/                     (excluded — patched yaml.v3)
├── bin/                       Build output
└── Makefile                   Build, test, lint, dist targets
```

### Key packages

| Package | Responsibility |
|---|---|
| `cmd` | CLI command definitions using Cobra |
| `config` | Configuration loading via Viper (file → env → flags) |
| `client` | HTTP client for OCI registry interactions with Basic Auth and media type fallback |
| `content` | CNAB content operations: manifest retrieval, inspection, deletion, purge |
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
