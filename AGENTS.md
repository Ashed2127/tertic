# Tertic

Go-based multi-agent CLI for Termux/Android. Build: `go build ./...`. Verify: `go vet ./...`.

## User preferences

- After completing a code change, commit and push to `origin/main` by default (user said "always").
- Progress UI is Devin-style: spinner shows live actions only — `Thinking...`, `Running: <cmd>`, `Editing: <path>`, `Reading: <path>` — never agent stage labels like "Planning..." or "Developing...". These are emitted via `Harness.NotifyProgress` with roles `"Thinking"`/`"Action"` and filtered in `cmd/agent/main.go` `OnProgress`.
