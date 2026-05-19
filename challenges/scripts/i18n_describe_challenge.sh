#!/usr/bin/env bash
# i18n_describe_challenge.sh
#
# Round-257 paired-mutation deep-doc challenge for digital.vasic.i18n.
#
# Validates that:
#   1. The deep-doc ledger (docs/test-coverage.md) lists every exported
#      symbol from pkg/{i18n,loader,middleware}.
#   2. The multi-locale fixture (tests/fixtures/i18n/payloads.json)
#      parses and contains at least 3 locales.
#   3. The multi-locale runner (challenges/runner/main.go) builds and
#      runs, byte-preserving non-ASCII messages through Bundle round-trip,
#      loader.LoadJSON / LoadJSONDir / LoadMap, template substitution,
#      fallback semantics, and middleware.New real HTTP transport.
#   4. The README enumerates the three packages and the round-257
#      anti-bluff guarantees section.
#
# Paired-mutation invariant (CONST-035 + CONST-050(B)):
#   With --anti-bluff-mutate the script plants a deliberate symbol-rename
#   mutation in the ledger (in a tmp copy), reruns validation, and asserts
#   the gate FAILS with exit 99. This proves the gate actually catches
#   ledger-vs-source drift instead of rubber-stamping it.
#
# Exit codes:
#   0  — gate PASS on clean tree
#   1  — gate FAIL on clean tree (real failure to fix)
#   99 — paired-mutation correctly detected (good — proves anti-bluff)
#   2  — usage / environment error

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODULE_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"

MUTATE=0
for arg in "$@"; do
    case "$arg" in
        --anti-bluff-mutate) MUTATE=1 ;;
        --help|-h)
            sed -n '1,30p' "$0"
            exit 0
            ;;
        *)
            echo "unknown argument: $arg" >&2
            exit 2
            ;;
    esac
done

PASS=0
FAIL=0
TOTAL=0

pass() { PASS=$((PASS+1)); TOTAL=$((TOTAL+1)); echo "  PASS: $1"; }
fail() { FAIL=$((FAIL+1)); TOTAL=$((TOTAL+1)); echo "  FAIL: $1"; }

LEDGER="${MODULE_DIR}/docs/test-coverage.md"
FIXTURE="${MODULE_DIR}/tests/fixtures/i18n/payloads.json"
RUNNER="${MODULE_DIR}/challenges/runner/main.go"
README="${MODULE_DIR}/README.md"

LEDGER_WORK="${LEDGER}"
TMP_LEDGER=""
if [ "${MUTATE}" -eq 1 ]; then
    TMP_LEDGER="$(mktemp)"
    cp "${LEDGER}" "${TMP_LEDGER}"
    # Plant a rename: GetMessage -> GetBogus_MUTATED
    sed -i 's/GetMessage/GetBogus_MUTATED/g' "${TMP_LEDGER}"
    LEDGER_WORK="${TMP_LEDGER}"
    echo "=== I18n Describe Challenge (anti-bluff-mutate mode) ==="
else
    echo "=== I18n Describe Challenge (clean mode) ==="
fi
echo ""

# Section 1: ledger presence and freshness
echo "Section 1: docs/test-coverage.md ledger"
if [ ! -f "${LEDGER_WORK}" ]; then
    fail "ledger missing at ${LEDGER_WORK}"
else
    pass "ledger present"
    if grep -q "round-257" "${LEDGER_WORK}"; then
        pass "ledger marked round-257"
    else
        fail "ledger missing round-257 marker"
    fi
    if grep -q "execution of tests and Challenges MUST guarantee" "${LEDGER_WORK}"; then
        pass "ledger carries Article XI §11.9 mandate"
    else
        fail "ledger missing Article XI §11.9 mandate"
    fi
fi

# Section 2: every exported pkg symbol appears in ledger
echo ""
echo "Section 2: exported symbols cross-reference"

extract_symbols() {
    local pkg_dir="$1"
    local files
    files=$(find "${pkg_dir}" -maxdepth 1 -type f -name '*.go' \
        ! -name '*_test.go')
    [ -z "${files}" ] && return 0
    # shellcheck disable=SC2086
    grep -hE '^(func ([A-Z][A-Za-z0-9_]*\()|func \([^)]+\) ([A-Z][A-Za-z0-9_]*\()|type [A-Z][A-Za-z0-9_]* )' \
        ${files} 2>/dev/null \
        | sed -E 's/^func \([^)]+\) ([A-Z][A-Za-z0-9_]*)\(.*$/\1/; s/^func ([A-Z][A-Za-z0-9_]*)\(.*$/\1/; s/^type ([A-Z][A-Za-z0-9_]*).*$/\1/' \
        | sort -u
}

CHECKED=0
MISSING=0
for pkg in i18n loader middleware; do
    PKG_DIR="${MODULE_DIR}/pkg/${pkg}"
    if [ ! -d "${PKG_DIR}" ]; then
        fail "pkg/${pkg} missing — cannot cross-reference"
        continue
    fi
    while IFS= read -r sym; do
        [ -z "${sym}" ] && continue
        CHECKED=$((CHECKED + 1))
        if grep -qE "\\b${sym}\\b" "${LEDGER_WORK}"; then
            : # symbol cross-referenced
        else
            fail "ledger missing symbol ${pkg}.${sym}"
            MISSING=$((MISSING + 1))
        fi
    done < <(extract_symbols "${PKG_DIR}")
done
if [ "${CHECKED}" -gt 0 ] && [ "${MISSING}" -eq 0 ]; then
    pass "all ${CHECKED} exported symbols cross-referenced in ledger"
fi

# Section 3: multi-locale fixture sanity
echo ""
echo "Section 3: multi-locale fixture"
if [ ! -f "${FIXTURE}" ]; then
    fail "fixture missing at ${FIXTURE}"
else
    pass "fixture present"
    LOCALE_COUNT=$(grep -oE '"locale":\s*"[^"]+"' "${FIXTURE}" | sort -u | wc -l)
    if [ "${LOCALE_COUNT}" -ge 3 ]; then
        pass "fixture covers ${LOCALE_COUNT} locales (>=3)"
    else
        fail "fixture covers only ${LOCALE_COUNT} locales (<3)"
    fi
fi

# Section 4: runner builds + runs against every package
echo ""
echo "Section 4: multi-locale runner build + run (real os + http transport)"
if [ ! -f "${RUNNER}" ]; then
    fail "runner missing at ${RUNNER}"
else
    pass "runner source present"
    cd "${MODULE_DIR}"
    if go build -o /tmp/i18n_round257_runner ./challenges/runner/ 2>/tmp/i18n_build.log; then
        pass "runner builds"
        if /tmp/i18n_round257_runner -fixtures "${FIXTURE}" > /tmp/i18n_run.log 2>&1; then
            pass "runner exit 0 across every package + locale"
            if grep -q "PASS: \[i18n\]\[sr\]" /tmp/i18n_run.log; then
                pass "i18n pkg Cyrillic (sr) GetMessage round-trip"
            else
                fail "i18n pkg Cyrillic (sr) missing from runner output"
            fi
            if grep -q "PASS: \[i18n\]\[ja\]" /tmp/i18n_run.log; then
                pass "i18n pkg Japanese (ja) GetMessage round-trip"
            else
                fail "i18n pkg Japanese (ja) missing from runner output"
            fi
            if grep -q "PASS: \[loader\]\[LoadJSONDir\]\[ar\]" /tmp/i18n_run.log; then
                pass "loader Arabic (ar) on-disk round-trip"
            else
                fail "loader Arabic (ar) on-disk section missing"
            fi
            if grep -q "PASS: \[loader\]\[LoadJSON\]" /tmp/i18n_run.log; then
                pass "loader LoadJSON single-file exercised"
            else
                fail "loader LoadJSON single-file missing"
            fi
            if grep -q "PASS: \[template\]\[zh-CN\]" /tmp/i18n_run.log; then
                pass "template substitution rune-safe for Han (zh-CN)"
            else
                fail "template substitution zh-CN missing"
            fi
            if grep -q "PASS: \[fallback\]" /tmp/i18n_run.log; then
                pass "fallback semantics enforced"
            else
                fail "fallback section missing"
            fi
            if grep -q "PASS: \[middleware\]\[query-overrides-header\]" /tmp/i18n_run.log; then
                pass "middleware query-overrides-header invariant enforced"
            else
                fail "middleware query-overrides-header section missing"
            fi
            if grep -q "PASS: \[middleware\]\[default-fallback\]" /tmp/i18n_run.log; then
                pass "middleware default-fallback path exercised"
            else
                fail "middleware default-fallback section missing"
            fi
        else
            fail "runner exit non-zero — see /tmp/i18n_run.log"
            sed -n '1,40p' /tmp/i18n_run.log
        fi
    else
        fail "runner build failed — see /tmp/i18n_build.log"
        sed -n '1,40p' /tmp/i18n_build.log
    fi
    rm -f /tmp/i18n_round257_runner
fi

# Section 5: README round-257 anti-bluff section
echo ""
echo "Section 5: README round-257 anti-bluff section"
if grep -q "Anti-bluff guarantees" "${README}"; then
    pass "README declares Anti-bluff guarantees"
else
    fail "README missing Anti-bluff guarantees section"
fi
if grep -q "round-257" "${README}"; then
    pass "README marked round-257"
else
    fail "README missing round-257 marker"
fi

# Cleanup mutated ledger if any
if [ -n "${TMP_LEDGER}" ]; then
    rm -f "${TMP_LEDGER}"
fi

echo ""
echo "=== Summary: ${PASS}/${TOTAL} PASS, ${FAIL} FAIL ==="

if [ "${MUTATE}" -eq 1 ]; then
    if [ "${FAIL}" -gt 0 ]; then
        echo "anti-bluff-mutate: gate correctly detected planted mutation (exit 99)"
        exit 99
    else
        echo "anti-bluff-mutate: gate FAILED to detect planted mutation — bluff!"
        exit 1
    fi
fi

if [ "${FAIL}" -gt 0 ]; then
    exit 1
fi
exit 0
