#!/usr/bin/env bash
# Benchmark the existing extraction code on deterministic synthetic input.
# No network requests, performance thresholds or persistent output files.
set -euo pipefail
cd "$(dirname "$0")"

BENCH='Benchmark(PreserveReadabilityStructure|ReadabilityExtraction)'

case "${1:-}" in
  "")
    [[ $# -eq 0 ]] || { echo "Usage: ./benchmark.sh [--smoke]" >&2; exit 2; }
    # Five measured runs for manual performance comparisons.
    exec go test ./internal/extraction -run '^$' -bench "$BENCH" -benchmem -benchtime=500ms -count=5
    ;;
  --smoke)
    [[ $# -eq 1 ]] || { echo "Usage: ./benchmark.sh [--smoke]" >&2; exit 2; }
    # Only verify the benchmarks run; not a performance regression threshold.
    exec go test ./internal/extraction -run '^$' -bench "$BENCH" -benchmem -benchtime=100ms -count=1
    ;;
  *)
    echo "Usage: ./benchmark.sh [--smoke]" >&2
    exit 2
    ;;
esac
