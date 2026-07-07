# cnabtool — Architecture Overview & Project Summary

## 1. Project Summary

**cnabtool** is a Go CLI utility for inspecting and manipulating **CNAB** (Cloud Native Application Bundle) artifacts stored in **OCI** (Open Container Initiative) container registries. It enables operators to fetch manifests, walk the full dependency graph of a CNAB project, and safely delete CNAB components from a registry.

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
│   └── delete <reference> [--dry-run]               Delete a CNAB project from registry
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
│   │   ├── client.go          OCI registry HTTP client (GET/DELETE)
│   │   └── client_test.go     ParseReference unit tests
│   ├── config/
│   │   └── config.go          Viper-based config (file/env/flags)
│   ├── content/
│   │   ├── manifest.go        GetManifest + ResponsePrettyPrint
│   │   ├── inspect.go         AddIndex/AddCnab/InspectCnab/ShowCnabReport
│   │   └── delete.go          DeleteCnab (leaf-first deletion)
│   ├── data/
│   │   └── data.go            All data models + global state (Gc, Sensitives, maps)
│   └── logging/
│       └── logging.go         5-level structured logging with credential redaction
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
| **Dependency graph** | `ItemByDigest` / `ItemByTag` maps in `pkg/data` for O(1) lookups; uplink/downlink chains built during inspection |
| **Logging** | Five verbosity levels; all log output passes through `maskcredentials()` to redact passwords and basic auth tokens |
| **Deletion strategy** | Identifies leaf nodes (items referenced by exactly one parent) and deletes them first, then the parent |

---

## 5. Data Flow

### Inspect flow
1. **Parse reference** → `RegClient.ParseReference("registry/repo/image:tag@digest")`
2. **Fetch manifest** → `GetManifest()` issues authenticated GET with Accept headers
3. **Walk index** → `AddCnab()` parses OCI index, iterates `manifests` entries, registers components by annotation
4. **Build graph** → `InspectCnab()` iterates all tags, fetches each manifest, resolves uplink/downlink chains, marks lost references
5. **Report** → `ShowCnabReport()` outputs JSON (compact or raw)

### Delete flow
1. Same inspection walk as above
2. Identify deletable items: children with only one parent reference + the parent itself
3. Issue `DELETE /manifests/<digest>` for each tag (202 = success)

---

## 6. Data Models (`pkg/data/data.go`)

- **`Config`** — Verbosity, Timeout, Unsecure, Client, Scheme, Raw, DryRun, Credentials
- **`Credentials`** — Username / Password
- **`RegIndex`** — Reference, Tag, Media, Annotation, Date, Digest, DownLinks, UpLinks, Lost count, Content
- **`ProjectList`** — Slice of all discovered `*RegIndex` items
- **`ItemByDigest` / `ItemByTag`** — Lookup maps for O(1) access
- **`CnabItem`** — Digest + annotation pair (uplink/downlink reference)
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

## 8. Observations

- **Single global state** — `Gc` (global config pointer) and package-level maps (`ItemByDigest`, `ItemByTag`, `ItemsQueue`) in `pkg/data` serve as mutable global state shared across content operations. This works for a CLI tool but may need scoping for testability or concurrency.
- **Sensitive data handling** — Passwords and basic auth tokens are explicitly tracked in `Sensitives` and redacted from all log output, which is a good security practice for a registry tool.
- **Patch dependency** — `yaml.v3` is replaced with a local patched version in `fixes/yaml.v3`, suggesting a compatibility or bug-fix workaround.
- **Limited test coverage** — Only `pkg/client/client_test.go` exists (testing `ParseReference`). The content inspection and deletion logic has no automated tests.
