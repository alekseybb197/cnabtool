# cnabtool — Architecture Overview & Project Summary

## 1. Project Summary

**cnabtool** is a Go CLI utility for inspecting and manipulating **CNAB** (Cloud Native Application Bundle) artifacts stored in **OCI** (Open Container Initiative) container registries. It enables operators to fetch manifests, walk the full dependency graph of a CNAB project, and safely delete CNAB components from a registry, including cleanup of empty "folders" via Artifactory REST API.

- **Module:** `cnabtool` (Go 1.20)
- **Version:** 0.1.1
- **Dependencies:** Cobra, Viper, pflag, jsonparser; local patched `yaml.v3`
- **Build targets:** darwin-amd64, linux-amd64, windows-amd64

---

## 2. CLI Interface

```
cnabtool
├── version                                          Print version info
├── content                                          Content operations (requires subcommand)
│   ├── manifest <reference>                         Get content manifest as JSON
│   ├── inspect <reference> [--raw]                  Inspect all items in a CNAB project
│   └── delete <reference> [--dry-run] [--purge]     Delete a CNAB project from registry
└── Global flags
    --config, -c      Config file path
    --verbosity, -v   Log level (0–4)
    --username, -u    Registry username
    --password, -p    Registry password
    --timeout, -t     Timeout in milliseconds (default 10000)
```

---

## 3. Package Architecture

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
│   │   └── client_test.go     ParseReference, NewRegClient, FillResponse, WebRequestEx, GetTagList tests
│   ├── config/
│   │   ├── config.go          Viper-based config (file/env/flags)
│   │   └── config_test.go     Defaults, singleton, reset, precedence tests
│   ├── content/
│   │   ├── manifest.go        GetManifest + ResponsePrettyPrint
│   │   ├── inspect.go         AddIndex/AddCnab/InspectCnab/ShowCnabReport
│   │   ├── delete.go          DeleteCnab (digest-based deletion)
│   │   └── purge.go           PurgeEmptyFolders (Artifactory API cleanup)
│   ├── data/
│   │   ├── data.go            All data models + global state (Gc, Sensitives, maps)
│   │   └── data_test.go       All types, global state, link management, Lost counting tests
│   └── logging/
│       ├── logging.go         5-level structured logging with credential redaction
│       └── logging_test.go    All log levels, masking, PrettyString tests
├── fixes/                     (excluded — patched yaml.v3)
├── bin/                       Build output
└── Makefile                   build / run / dist / clean / fmt / test / lint
```

---

## 4. Design Patterns & Key Decisions

| Concern | Approach |
|---|---|
| **CLI framework** | Cobra + Viper for hierarchical commands, config merging, and flag binding |
| **Config loading** | Three-tier merge: config file → env vars (`CNAB_*`) → CLI flags (highest priority) |
| **Registry interaction** | `RegClient` struct encapsulates reference parsing, HTTP GET/DELETE, and response decoding with Basic Auth |
| **Media type fallback** | `GetRegIndex()` retries with multiple media types (oci-image-index → docker-manifest-v2 → docker-manifest-v1 → cnab) |
| **Arbitrary HTTP** | `WebRequestEx(method, url)` for vendor-specific APIs (Artifactory REST API) |
| **Dependency graph** | `ItemByDigest` / `ItemByTag` maps in `pkg/data` for O(1) lookups; uplink/downlink chains built during inspection |
| **Logging** | Five verbosity levels; all log output passes through `maskcredentials()` to redact passwords and basic auth tokens |
| **Deletion strategy** | Collects all items by digest (not tag), protects against duplicates with `deletedDigests` map, deletes leaf nodes first, then parents |
| **Purge strategy** | After deletion, walks up the directory tree via Artifactory Storage API, deletes empty folders recursively with adaptive timeout detection |

---

## 5. Data Flow

### Inspect flow
1. **Parse reference** → `RegClient.ParseReference("registry/repo/image:tag@digest")`
2. **Fetch manifest** → `GetManifest()` issues authenticated GET with Accept headers
3. **Walk index** → `AddCnab()` parses OCI index, iterates `manifests` entries, registers components by annotation
4. **Walk tags** → `InspectCnab()` iterates all tags from `tags/list`, fetches each manifest, resolves uplink/downlink chains
5. **Fetch untagged** → For each DownLink not found by tag, makes a direct `GET` by digest to handle config/invocation manifests that have no tag
6. **Report** → `ShowCnabReport()` outputs JSON (compact or raw)

### Delete flow
1. Parse reference to get project metadata
2. Iterate all CNAB indexes in `ItemByTag`, collect all DownLinks (children) and the index itself (parent)
3. Delete by **digest** (not tag), skip items with `UpLinks > 1` (shared between parents), protect against duplicates
4. Issue `DELETE /manifests/<digest>` for each item (202 = success)

### Purge flow (`--purge` flag)
1. After deletion, determine `repo-key` (from `--repo-key` flag or auto-derived from hostname via `deriveRepoKey()`)
2. Start from `cl.Repository` path (e.g., `cnab/myapp/1.0.0/myapp`)
3. Loop: `GET /artifactory/api/storage/{repoKey}/{path}?list` → if `children` is empty, delete folder; else stop
4. Delete via `DELETE /artifactory/{repoKey}/{path}` with a dedicated 180-second timeout client
5. Move up with `path.Dir()` and repeat until a non-empty folder, root, or adaptive threshold is reached
6. **Adaptive termination:** if a DELETE takes >5× the average of previous DELETEs, stop — this prevents hanging on large parent directories
7. In `--dry-run`, every folder check and potential deletion is logged with `[dry-run] Purge: ...`

---

## 6. Data Models (`pkg/data/data.go`)

- **`Config`** — Verbosity, Timeout, Unsecure, Client, Scheme, Raw, DryRun, **Purge**, **RepoKey**, Credentials
- **`Credentials`** — Username / Password
- **`RegIndex`** — Reference, Tag, Media, Annotation, Date, Digest, DownLinks, UpLinks, Lost count, Content
- **`ProjectList`** — Slice of all discovered `*RegIndex` items
- **`ItemByDigest` / `ItemByTag`** — Lookup maps for O(1) access
- **`CnabItem`** — Digest + annotation pair (uplink/downlink reference)
- **`ItemsQueue`** — Queue of digests to process
- **`Sensitives`** — List of strings to redact from log output
- **Item type constants** — `ItemTypeCnab`, `ItemTypeImage`, `ItemTypeConfig`, `ItemTypeStuff`

---

## 7. Build & Tooling (`Makefile`)

| Target | Description |
|---|---|
| `make build` | Builds to `bin/cnabtool` (CGO disabled, version injected via `-ldflags`) |
| `make run` | Runs via `go run` |
| `make dist` | Cross-compiles for darwin/amd64, linux/amd64, windows/amd64 |
| `make clean` | Removes `bin/` and `dist/` |
| `make fmt` | Runs `gofumpt` |
| `make test` | Runs `go test ./...` |
| `make lint` | Runs `golangci-lint` |

---

## 8. Changelog & Recent Fixes

### v0.1.1 — Purge path with adaptive timeout (current)

**Problem:** Artifactory DELETE on a directory with subdirectories can take **more than 60 seconds**. The old client would abort the connection, and the program would crash with a timeout error before reaching the end of the purge cycle.

**Solution:**
- Dedicated `http.Client` with **180-second timeout** for purge DELETE operations
- **Graceful timeout handling:** `*url.Error.Timeout()` is treated as a warning, not a fatal error — the purge loop breaks but the program completes normally
- **Adaptive termination:** tracks DELETE durations in a slice; if a DELETE takes >5× the average of previous deletions, the purge stops immediately (prevents hanging on large parent directories)

### v0.1.1 — Logging level refinement

**Problem:** On verbosity level 2 (Normal), purge produced no output at all — the user couldn't see what was being deleted. On level 4 (Debug), the storage API check requests were too verbose.

**Solution:**
- Added `logging.Normal()` function (level 2) that includes caller info (file, function, line)
- **Level 2 (Normal):** folder deletion messages, delete time, adaptive termination warnings, completion
- **Level 3+ (Info/Debug):** storage API checks, "folder not empty" messages
- Result on level 2:
  ```
  Purge: delete empty folder cnab/myapp/1.0.0/myapp/config
  Purge: folder cnab/myapp/1.0.0/myapp/config deleted in 200ms
  Purge: delete empty folder cnab/myapp/1.0.0/myapp
  Purge: folder cnab/myapp/1.0.0/myapp deleted in 150ms
  Purge: delete empty folder cnab/myapp
  [warning] Purge: delete time 3m0s is >5× average 177ms. Parent folder likely too heavy. Stopping purge.
  Purge: completed
  ```

### v0.1.1 — Fix: Fetch untagged manifests by digest (`pkg/content/inspect.go`)

**Problem:** CNAB components (config, invocation, component) have no tags — they exist only as `manifests` entries in the OCI Image Index with `digest` and `mediaType`. The old code only iterated tags from `tags/list`, so untagged components were always marked as `Lost`.

**Solution:** After iterating all tags, for each DownLink not found in `ItemByDigest`, make a direct `GET /v2/<repo>/manifests/<digest>` via `cl.GetRegIndex()`. If the response is 200, register the manifest in the global maps. Only increment `Lost` if the fetch fails.

### v0.1.1 — Fix: Digest-based deletion (`pkg/content/delete.go`)

**Problem:** The old code used `data.ItemByTag[cl.Tag].DownLinks` — this only accessed the first tag found during inspection. For fetched untagged components, `Tag` is empty, so `ItemByTag[""]` returned nil. Additionally, deletion used `ItemByTag[tag].Digest` — for untagged items this was nil, causing deletion of wrong digests or duplicates.

**Solution:** Rewrite to iterate all CNAB indexes in `ItemByTag`, collect all DownLinks by digest, use `deletedDigests` map to prevent duplicates, and issue `DELETE /manifests/<digest>` for each unique digest.

### v0.1.1 — Fix: Logging message format (`pkg/logging/logging.go`)

**Problem:** Log messages included the full `RegClient` struct in debug output, exposing credentials in logs.

**Solution:** Added `maskcredentials()` function that redacts passwords and basic auth tokens from all log output. Applied to `Error`, `Fatal`, `Message`, `Info`, and `Debug` functions.

### v0.1.1 — Initial content commands (`cmd/content.go`)

**Problem:** No way to inspect or delete CNAB content from the registry.

**Solution:** Added `content manifest`, `content inspect`, and `content delete` subcommands with full dependency graph traversal and leaf-first deletion strategy.

---

## 9. Usage Examples

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

---

## 10. Observations

- **Single global state** — `Gc` (global config pointer) and package-level maps (`ItemByDigest`, `ItemByTag`, `ItemsQueue`) in `pkg/data` serve as mutable global state shared across content operations. This works for a CLI tool but may need scoping for testability or concurrency.
- **Sensitive data handling** — Passwords and basic auth tokens are explicitly tracked in `Sensitives` and redacted from all log output, which is a good security practice for a registry tool.
- **Vendor-specific extension** — `--purge` adds Artifactory REST API dependency (`/api/storage/`, `/artifactory/`). This is not portable to other OCI registries (Harbor, Quay, etc.). Consider abstracting the purge logic behind an interface if multi-registry support is planned.
- **Auto-derived repo-key** — `deriveRepoKey()` extracts the repo-key from hostname automatically (e.g., `registry.example.com` → `registry`). Port numbers are stripped (`host:port` → `host`).
- **Dry-run transparency** — `--dry-run --purge` shows every folder check and potential deletion, making it easy to verify the purge path before committing.
- **Adaptive purge termination** — The >5× average timeout detection prevents hanging on large parent directories without requiring arbitrary timeout values.
- **Patch dependency** — `yaml.v3` is replaced with a local patched version in `fixes/yaml.v3`, suggesting a compatibility or bug-fix workaround.
- **Test coverage** — Automated tests exist for `client` (reference parsing, HTTP client), `config` (loading, precedence), `data` (all models, link management), and `logging` (all levels, masking). Content inspection, deletion, and purge logic has no automated tests.
- **Tag-based architecture limitation** — The `ItemByTag` map is the primary registry index. Untagged manifests require special handling (empty string key or digest-based lookup). Consider whether a dedicated `ItemByDigest`-only index would be cleaner.
