# Frozen extraction contract fixtures

These are hand-authored, publisher-neutral HTML sources, **not** captured CNN, NPR, or Seeking Alpha pages. They contain no live login data, scripts or third-party assets that tests fetch.

- `structural_lead_contract.html`: metadata thumbnail differs from the opening editorial `picture`; a contributor portrait precedes it; a later image follows article prose. The contract asserts selection of the source editorial lead, not the preview or portrait.
- `source_role_contract.html`: ancestor context determines author portrait, editorial figure and linked contributor portrait; existing lead roles remain protected.

`internal/reader/frozen_extraction_contract_integration_test.go` runs against these files without network access and tests the actual article handoff/assembler and DOM-role interfaces. Keep fixture semantics stable when migrating parser, media and assembly packages. A fixture change must be reviewed as a behavior contract change rather than silently updating expectations.

**Coverage limitation:** These do not replace sanitized, permission-appropriate real publisher HTML acceptance fixtures. CI still cannot verify current CNN or NPR behavior from live pages. Such captures should be added in a separately reviewed change with expected title/byline/body and media identity, after inspecting for copyright or private content concerns.
