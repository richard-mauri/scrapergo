# Repository layout

ScraperGo uses a single Go `package main` at the repository root. Keeping its Go source and `*_test.go` regression tests together allows `go test ./...`, `go build .`, `./devbuild.sh`, and `./scrapergo.sh` to work without changing imports, application behavior, or deployment.

- `main.go` and other root `*.go` files: executable and supporting article extraction logic.
- Root `*_test.go` files: regression coverage. Do not remove or move them into separate Go package directories merely to make the root look smaller.
- `testdata/`: fixture data consumed by Go tests.
- `docs/`: architecture notes; `docs/archive/` preserves historical version/implementation notes formerly in the repository root.
- `safari-extension/`: companion Safari browser extension.
- `AGENTS.md`: development workflow and safety instructions.
- `README.md`: project overview and usage.
- `devbuild.sh`, `scrapergo.sh`: established macOS build and runtime launchers. They must remain at the root.
- `dist/`: generated multi-architecture executables; ignored by Git.
- `clean-article-reader`: a possible root executable created by a bare `go build` command; ignored by Git, not used by either launcher.
- `scrapergo.log`, `.cloudflared/`: local runtime log and credentials; ignored by Git. The cleanup does not remove or modify them.

To test a housekeeping candidate on Mac:

```bash
git fetch origin
git checkout chore/repository-layout-cleanup
git pull --ff-only origin chore/repository-layout-cleanup
go test -count=1 ./...
./devbuild.sh
./scrapergo.sh
```

Only historical Markdown files were relocated. Existing file contents are preserved, and source/test/package names, module path, application version, launcher configuration, ports, tunnel, tokens, and user data are unchanged.
