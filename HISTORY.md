# ScraperGo historical development reference

**Historical record, not current operational instructions.** Current requirements live in [AGENTS.md](AGENTS.md), [ARCHITECTURE.md](ARCHITECTURE.md), [DEVELOPMENT.md](DEVELOPMENT.md), and [README.md](README.md). Git history retains the exact former README/AGENTS handoff text; consult previous revisions for detailed provenance, not instructions for the next PR.

## Milestones

- Earlier work established clean article extraction, rich inline links, media role and source-identity handling, editorial lead selection, contributor header reconciliation, partial-page recognition, diagnostic traces, and Safari extension integration. Historical numbered candidate narratives remain recoverable from [README at the pre-consolidation commit](https://github.com/richard-mauri/scrapergo/blob/eaa08b8dcefa1ce2b9232c3e38839905cc575065/README.md).
- Design and migration proposals are recorded in [docs/structured-article-redesign.md](docs/structured-article-redesign.md), [docs/architecture-migration.md](docs/architecture-migration.md), and [docs/architecture-review.md](docs/architecture-review.md). Their PR states, package owners, and instructions describe earlier snapshots.
- PRs #53–#57 completed production architecture relocation and enforcement. The root became a minimal executable; backend implementations moved into owning `internal/` packages with deliberate composition and retrieval boundaries.
- PRs #58–#62 progressively relocated root regression suites and eliminated migration test facades. After #62, the two root tests were `architecture_boundary_test.go` and `version_test.go`.
- PR #63 removed the remaining two confirmed production compatibility forwarding wrappers. The consolidated baseline is `main` commit `eaa08b8dcefa1ce2b9232c3e38839905cc575065`; semantic version `0.2.85`.

## Archive and provenance

Preserve [docs/archive/](docs/archive/) for versioned image/link investigations and earlier release notes. The repository history provides exact content for deleted development-candidate narrative entries without implying an obsolete draft PR remains open. See [pre-consolidation AGENTS instructions](https://github.com/richard-mauri/scrapergo/blob/eaa08b8dcefa1ce2b9232c3e38839905cc575065/AGENTS.md) for historical handoffs only.

Historical planning documents may contain accurate lessons and test motivations, but do not override current code ownership, secure retrieval or owner approval. The ongoing known limitation is that synthetic tests and passing CI do not prove current behavior for every live publisher, Safari environment or tunnel deployment.
