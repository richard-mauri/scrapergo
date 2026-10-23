# Repository layout (current)

The only production Go file in repository root is `main.go`. Root tests are `architecture_boundary_test.go` and `version_test.go`; functional, regression and boundary tests belong in their owning `internal/` packages. See [architecture contract](../ARCHITECTURE.md).

- `internal/`: production packages, neighboring domain unit tests and integration tests at their actual service boundaries.
- `testdata/` and internal package test fixtures: frozen synthetic test resources, not live publisher snapshots.
- `safari-extension/`: temporary Safari Web Extension and [setup instructions](../safari-extension/README.md).
- `docs/`: explanatory and historical material; `docs/archive/` retains versioned historical research.
- `README.md`: user overview and API; `AGENTS.md`: short mandatory contributor instructions; `DEVELOPMENT.md`: contributor workflow; `ARCHITECTURE.md`: ownership contract.
- `devbuild.sh`, `scrapergo.sh`: local Mac build and machine-specific runtime launcher at root. `dist/`, private build counter, tunnel tokens and runtime logs are not tracked source files.

For branch-specific Mac test instructions use [DEVELOPMENT.md](../DEVELOPMENT.md). Prior root-logic and root-test retention instructions were superseded by the completed migrations.
