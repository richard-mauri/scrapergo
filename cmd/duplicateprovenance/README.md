# Duplicate provenance — Issue #119 diagnostic

The existing NewsGo `cmd/duplicateaudit` tool reports exact/overlapping candidate duplicate text **at the final ScraperGo read response**. This offline ScraperGo command reads the **already opt-in** private `diagnostics/recovery.jsonl` and summarizes how the candidate counts change across Readability and extraction stages. It does not fetch publisher pages, change article output, or reveal article URLs/text/IDs. The command is developed inside the **existing draft PR #120**, not a separate issue or PR.

## Collect the same 100-URL cohort

Run the PR #120 branch on your Mac after `go test -count=1 ./...`, `./devbuild.sh` and `./scrapergo.sh`. Its launcher retains `SCRAPERGO_RECOVERY_TRACE=1`, the existing private `.recovery-key`, and the native ARM64 binary. Before starting the batch, note the current `go run ./cmd/recoveryreport` visit total.

From the NewsGo checkout (in a separate Terminal), replay the **same** `newsgo-scrapergo-urls.json` used for the previous 100-URL comparison:

```sh
cd ~/projects/newsgo
go run ./cmd/duplicateaudit \
  -input ~/Downloads/newsgo-scrapergo-urls.json \
  -scrapergo https://scrapergo.themauris.org/api/read \
  -public -limit 100 \
  -output ~/Downloads/duplicate-audit-pr120-provenance.csv
```

Choose a new output filename rather than overwriting an existing CSV. Avoid separate browser activity during the batch if possible. Then return to the ScraperGo checkout:

```sh
cd ~/projects/scrapergo
go run ./cmd/recoveryreport
go run ./cmd/duplicateprovenance -last 100
```

The last command prints aggregate counts at `readability-render-text`, each reached recovery stage, and `reader-postprocessing`; it reads only recent records with the new duplicate metrics and ignores older historical records. No file sharing is needed: paste its **aggregate stdout** and upload the NewsGo CSV. If non-batch reads occurred concurrently, the 100 most recent trace records may not precisely align with the exported URL cohort. If the batch includes fallbacks not reaching Readability, fewer than 100 records may qualify.

## Interpretation

- Duplicate candidates already present at `readability-render-text` precede all ScraperGo recovery stages; inspect publisher source HTML versus Readability `RenderHTML`/`RenderText` before changing recovery logic.
- A stage whose `added` counts increase is a candidate for **introducing** duplicate appearances; a stage whose `removed` counts increase is a candidate for **eliminating** them. The net sum alone can hide simultaneous additions/removals.
- `reader-postprocessing` isolates changes after extraction in the canonical document assembly.
- Counts are deliberately analogous to NewsGo's exact/overlap *screening*, not proof of a defect. Legitimate repeated quotations, updates, images, bylines and source semantics must be preserved.

Trace records contain counts but no article text/full URL or ID in the report. The private JSONL itself includes a keyed opaque article ID and publisher host; it is ignored by Git and **must not be committed**. Existing historical records and the private `.recovery-key` are retained.

For publisher source evidence, use the existing **secure** `cmd/htmlcapture` and offline `cmd/readabilityprobe` commands described in `cmd/readabilityprobe/README.md`. Real HTML may be restricted and should remain local; do not upload or commit complete copyrighted captures or credentials. The production duplicate-text fix remains unproven until such evidence identifies its origin and regression tests preserve legitimate repetition.
