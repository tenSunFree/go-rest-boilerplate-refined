#!/usr/bin/env bash
# scripts/coverage.sh
#
# Runs Go tests with coverage (mirroring CI's `-race -covermode=atomic`),
# converts the Go coverage profile to LCOV, and generates an HTML report
# using genhtml/lcov-viewer -- same tree-style report as the Flutter
# project's scripts/coverage.sh, for a consistent local dev experience.
#
# The reported percentage is filtered to match what Codecov shows: paths
# listed under `ignore:` in codecov.yml are excluded from the profile
# before computing the summary / HTML report. codecov.yml remains the
# single source of truth for exclusions -- this script parses it rather
# than hardcoding a second copy of the same list, so the two can never
# drift out of sync.
#
# Usage:
#   bash scripts/coverage.sh
#
# Optional:
#   NO_OPEN=1 bash scripts/coverage.sh   # skip auto-opening the browser
set -euo pipefail
cd "$(dirname "$0")/.."

COVERAGE_DIR="coverage"
COVERAGE_PROFILE="$COVERAGE_DIR/coverage.out"
FILTERED_PROFILE="$COVERAGE_DIR/coverage.filtered.out"
LCOV_FILE="$COVERAGE_DIR/lcov.info"
HTML_DIR="$COVERAGE_DIR/html"
REPORT_PATH="$(pwd)/$HTML_DIR/index.html"
CODECOV_CONFIG="codecov.yml"

mkdir -p "$HTML_DIR"

echo ""
echo "==> go mod download"
go mod download

echo ""
echo "==> go test -race -covermode=atomic -coverprofile=$COVERAGE_PROFILE ./..."
set +e
go test -race -count=1 -covermode=atomic -coverprofile="$COVERAGE_PROFILE" ./...
race_status=$?
set -e

if [[ $race_status -ne 0 ]]; then
  echo ""
  echo "WARNING: '-race' build failed (commonly missing a working C compiler/cgo on this machine)."
  echo "Falling back to a non-race run for local coverage. CI still runs with -race on Linux."
  echo ""
  echo "==> go test -covermode=atomic -coverprofile=$COVERAGE_PROFILE ./... (no -race)"
  go test -count=1 -covermode=atomic -coverprofile="$COVERAGE_PROFILE" ./...
fi

if [[ ! -s "$COVERAGE_PROFILE" ]]; then
  echo "ERROR: $COVERAGE_PROFILE was not generated or is empty."
  exit 1
fi

source_count=$(tail -n +2 "$COVERAGE_PROFILE" | cut -d: -f1 | sort -u | wc -l | tr -d ' ')
echo ""
echo "Coverage profile generated: $COVERAGE_PROFILE ($source_count source files)"

# ------------------------------------------------------------
# Filter out codecov.yml's `ignore:` paths so the local percentage
# matches what Codecov reports on GitHub. If codecov.yml or its
# `ignore:` list is missing, this is a no-op and the raw profile is
# used as-is.
# ------------------------------------------------------------
cp "$COVERAGE_PROFILE" "$FILTERED_PROFILE"

if [[ -f "$CODECOV_CONFIG" ]]; then
  MODULE="$(go list -m)"

  ignore_patterns=$(awk '
    /^ignore:/ { f=1; next }
    f && /^[^[:space:]#]/ { f=0 }
    f && /^[[:space:]]*-/ {
      line=$0
      sub(/^[[:space:]]*-[[:space:]]*/, "", line)
      gsub(/^"|"$/, "", line)
      print line
    }
  ' "$CODECOV_CONFIG")

  if [[ -n "$ignore_patterns" ]]; then
    echo ""
    echo "==> excluding codecov.yml ignore paths from local report:"
    while IFS= read -r pattern; do
      [[ -z "$pattern" ]] && continue
      echo "    - $pattern"
      # Only handles simple "dir/**" globs (which is all codecov.yml
      # currently uses). Strip the trailing "/**" and drop any line
      # whose import path starts with module/pattern/.
      prefix="${pattern%/\*\*}"
      grep -v -E "^${MODULE}/${prefix}/" "$FILTERED_PROFILE" > "${FILTERED_PROFILE}.tmp" \
        && mv "${FILTERED_PROFILE}.tmp" "$FILTERED_PROFILE"
    done <<< "$ignore_patterns"
  fi
fi

echo ""
echo "==> coverage summary (raw, unfiltered — includes ignored paths)"
go tool cover -func="$COVERAGE_PROFILE" | tail -1

echo ""
echo "==> coverage summary (filtered — matches Codecov's codecov.yml ignore list)"
go tool cover -func="$FILTERED_PROFILE" | tail -1

# ------------------------------------------------------------
# Convert the *filtered* Go coverage profile to LCOV so we can reuse the
# same genhtml/lcov-viewer HTML report style as the Flutter project, and
# so the tree view / percentages match Codecov.
# Requires: go install github.com/jandelgado/gcov2lcov@latest
# ------------------------------------------------------------
echo ""
echo "==> converting Go coverage profile to LCOV"
if ! command -v gcov2lcov >/dev/null 2>&1; then
  echo ""
  echo "gcov2lcov is not installed; cannot generate the LCOV-style HTML report."
  echo "Install it with:"
  echo "       go install github.com/jandelgado/gcov2lcov@latest"
  echo ""
  echo "Falling back to Go's built-in HTML report instead."
  go tool cover -html="$FILTERED_PROFILE" -o "$REPORT_PATH"
  echo "HTML report: $REPORT_PATH"
else
  gcov2lcov -infile="$FILTERED_PROFILE" -outfile="$LCOV_FILE"

  # ------------------------------------------------------------
  # Generate the HTML report.
  # Prefers genhtml (from lcov) if installed; otherwise falls back to
  # @lcov-viewer/cli (npm, no admin/Perl required) if available.
  # Set NO_OPEN=1 to skip auto-opening the browser afterwards.
  # ------------------------------------------------------------
  if command -v genhtml >/dev/null 2>&1; then
    genhtml "$LCOV_FILE" -o "$HTML_DIR"
    echo "HTML report: $REPORT_PATH"
  elif command -v lcov-viewer >/dev/null 2>&1; then
    lcov-viewer lcov -o "$HTML_DIR" "$LCOV_FILE"
    echo "HTML report: $REPORT_PATH"
  else
    echo ""
    echo "Neither genhtml nor lcov-viewer are installed; HTML report generation will be skipped."
    echo "LCOV report available at: $LCOV_FILE"
    echo "Choose one to install:"
    echo "  1) genhtml (lcov package, requires system administrator privileges):"
    echo "       Windows (choco, admin required): choco install lcov -y"
    echo "       macOS                     : brew install lcov"
    echo "       Ubuntu/Debian             : sudo apt-get install lcov"
    echo "  2) lcov-viewer (npm, no system administrator privileges required):"
    echo "       npm install -g @lcov-viewer/cli"
    exit 0
  fi
fi

if [[ "${NO_OPEN:-0}" != "1" ]]; then
  echo ""
  echo "==> opening report in default browser"
  if command -v open >/dev/null 2>&1; then
    # macOS
    open "$REPORT_PATH"
  elif command -v xdg-open >/dev/null 2>&1; then
    # Linux (desktop) / most WSL setups with a browser bridge configured
    xdg-open "$REPORT_PATH" >/dev/null 2>&1 &
  elif command -v wslview >/dev/null 2>&1; then
    # WSL, if wslu is installed (sudo apt install wslu)
    wslview "$REPORT_PATH"
  elif [[ "${OS:-}" == "Windows_NT" ]] || command -v cmd.exe >/dev/null 2>&1; then
    # Git Bash / WSL fallback via Windows explorer
    cmd.exe /c start "" "$(wslpath -w "$REPORT_PATH" 2>/dev/null || echo "$REPORT_PATH")" >/dev/null 2>&1 || true
  else
    echo "No available open command found. Please open manually: $REPORT_PATH"
  fi
fi