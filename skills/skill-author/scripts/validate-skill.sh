#!/usr/bin/env bash
# validate-skill.sh — Validates an agentkit skill directory against the agentskills.io spec
# Usage: bash validate-skill.sh <skill-dir>
# Exit: 0 if all checks PASS or WARN; 1 if any check FAIL

set -euo pipefail

SKILL_DIR="${1:-}"
PASS=0
WARN=1
FAIL=2

total_pass=0
total_warn=0
total_fail=0

emit() {
    local level="$1"
    local message="$2"
    case "$level" in
        $PASS) echo "PASS  $message"; ((total_pass++)) || true ;;
        $WARN) echo "WARN  $message"; ((total_warn++)) || true ;;
        $FAIL) echo "FAIL  $message"; ((total_fail++)) || true ;;
    esac
}

if [[ -z "$SKILL_DIR" ]]; then
    echo "Usage: $0 <skill-dir>"
    echo "Example: $0 skills/aws/"
    exit 1
fi

# Normalize: strip trailing slash
SKILL_DIR="${SKILL_DIR%/}"

echo "Validating: $SKILL_DIR"
echo "---"

# --- Check 1: SKILL.md exists ---
SKILL_MD="$SKILL_DIR/SKILL.md"
if [[ -f "$SKILL_MD" ]]; then
    emit $PASS "SKILL.md exists at $SKILL_MD"
else
    emit $FAIL "SKILL.md missing — expected at $SKILL_MD"
fi

# --- Check 2: Line count ---
if [[ -f "$SKILL_MD" ]]; then
    line_count=$(wc -l < "$SKILL_MD" | tr -d ' ')
    if [[ "$line_count" -lt 500 ]]; then
        emit $PASS "SKILL.md line count: $line_count (< 500)"
    elif [[ "$line_count" -lt 600 ]]; then
        emit $WARN "SKILL.md line count: $line_count (500–600 — consider moving content to references/)"
    else
        emit $FAIL "SKILL.md line count: $line_count (> 600 — must reduce; move content to references/)"
    fi
fi

# --- Check 3: Frontmatter — name field ---
if [[ -f "$SKILL_MD" ]]; then
    name_val=$(grep -m1 "^name:" "$SKILL_MD" | sed 's/^name:[[:space:]]*//' | tr -d '"' || true)
    folder_name=$(basename "$SKILL_DIR")
    if [[ -z "$name_val" ]]; then
        emit $FAIL "Frontmatter missing 'name:' field"
    elif [[ "$name_val" == "$folder_name" ]]; then
        emit $PASS "Frontmatter name '$name_val' matches folder name '$folder_name'"
    else
        emit $FAIL "Frontmatter name '$name_val' does not match folder name '$folder_name'"
    fi
fi

# --- Check 4: Frontmatter — description field ---
if [[ -f "$SKILL_MD" ]]; then
    if grep -q "^description:" "$SKILL_MD"; then
        # Check that description starts with "Use when" (within the first 10 lines after the field)
        desc_line=$(grep -A5 "^description:" "$SKILL_MD" | grep -i "use when" || true)
        if [[ -n "$desc_line" ]]; then
            emit $PASS "Frontmatter 'description' contains 'Use when' activation signal"
        else
            emit $WARN "Frontmatter 'description' present but does not start with 'Use when' — check activation clarity"
        fi
    else
        emit $FAIL "Frontmatter missing 'description:' field"
    fi
fi

# --- Check 5: name format (lowercase-hyphens only) ---
if [[ -f "$SKILL_MD" ]]; then
    name_val=$(grep -m1 "^name:" "$SKILL_MD" | sed 's/^name:[[:space:]]*//' | tr -d '"' || true)
    if [[ -n "$name_val" ]]; then
        if echo "$name_val" | grep -qE "^[a-z0-9-]+$"; then
            emit $PASS "name format valid: '$name_val' (lowercase-hyphens)"
        else
            emit $FAIL "name format invalid: '$name_val' (use lowercase letters, numbers, hyphens only)"
        fi
    fi
fi

# --- Check 6: references/ directory if SKILL.md mentions reference files ---
if [[ -f "$SKILL_MD" ]]; then
    mentions_refs=$(grep -i "references/" "$SKILL_MD" | grep -v "^#\|<!--" || true)
    refs_dir="$SKILL_DIR/references"
    if [[ -n "$mentions_refs" ]]; then
        if [[ -d "$refs_dir" ]]; then
            ref_count=$(find "$refs_dir" -name "*.md" | wc -l | tr -d ' ')
            if [[ "$ref_count" -gt 0 ]]; then
                emit $PASS "references/ directory exists with $ref_count .md file(s)"
            else
                emit $FAIL "references/ directory exists but contains no .md files"
            fi
        else
            emit $FAIL "SKILL.md mentions 'references/' but directory does not exist"
        fi
    else
        if [[ -d "$refs_dir" ]]; then
            emit $WARN "references/ directory exists but SKILL.md does not link to it — add a Reference Files table"
        else
            emit $PASS "No references/ required (single-domain skill)"
        fi
    fi
fi

# --- Check 7: Reference file line counts ---
if [[ -d "$SKILL_DIR/references" ]]; then
    while IFS= read -r ref_file; do
        ref_lines=$(wc -l < "$ref_file" | tr -d ' ')
        ref_name=$(basename "$ref_file")
        if [[ "$ref_lines" -lt 50 ]]; then
            emit $FAIL "references/$ref_name: $ref_lines lines (too thin — stub risk)"
        elif [[ "$ref_lines" -lt 200 ]]; then
            emit $WARN "references/$ref_name: $ref_lines lines (thin — consider expanding or merging into SKILL.md)"
        elif [[ "$ref_lines" -le 600 ]]; then
            emit $PASS "references/$ref_name: $ref_lines lines (OK)"
        else
            emit $FAIL "references/$ref_name: $ref_lines lines (too large — consider splitting)"
        fi
    done < <(find "$SKILL_DIR/references" -name "*.md" | sort)
fi

# --- Check 8: Stub / placeholder detection ---
stub_matches=$(grep -ri "TODO\|FIXME\|coming soon\|placeholder\|stub\|not yet implemented\|to be added" "$SKILL_DIR" 2>/dev/null || true)
if [[ -z "$stub_matches" ]]; then
    emit $PASS "No stub or placeholder text found"
else
    emit $FAIL "Stub or placeholder text found (remove before merging):"
    echo "$stub_matches" | head -10 | sed 's/^/       /'
fi

# --- Check 9: Injection safety — mid-document YAML separator ---
if [[ -f "$SKILL_MD" ]]; then
    # Count lines with only "---"; first two (open/close of frontmatter) are expected
    yaml_sep_count=$(grep -c "^---$" "$SKILL_MD" || true)
    if [[ "$yaml_sep_count" -le 2 ]]; then
        emit $PASS "No mid-document YAML separators (injection safe)"
    else
        emit $FAIL "Found ${yaml_sep_count} '---' lines — extra YAML separators may indicate injection; check SKILL.md"
    fi
fi

# --- Check 10: Injection safety — instruction override patterns ---
injection_matches=$(grep -ri "ignore previous.*instruction\|disregard.*instruction\|\[INST\]\|<<SYS>>" "$SKILL_DIR" 2>/dev/null || true)
if [[ -z "$injection_matches" ]]; then
    emit $PASS "No instruction-override injection patterns found"
else
    emit $FAIL "Injection patterns found (must remove before merging):"
    echo "$injection_matches" | head -5 | sed 's/^/       /'
fi

# --- Check 11: No personal credentials or absolute home paths ---
cred_matches=$(grep -ri "api_key[[:space:]]*=\|access_token[[:space:]]*=\|secret_key[[:space:]]*=\|password[[:space:]]*=\|/Users/[a-z]\|/home/[a-z]" "$SKILL_DIR" 2>/dev/null || true)
if [[ -z "$cred_matches" ]]; then
    emit $PASS "No credentials or personal absolute paths found"
else
    emit $FAIL "Credentials or personal paths found (must remove before committing to public repo):"
    echo "$cred_matches" | head -5 | sed 's/^/       /'
fi

# --- Summary ---
echo ""
echo "---"
echo "Results: ${total_pass} PASS  ${total_warn} WARN  ${total_fail} FAIL"
echo ""

if [[ "$total_fail" -gt 0 ]]; then
    echo "STATUS: FAIL — fix all FAIL items before merging"
    exit 1
elif [[ "$total_warn" -gt 0 ]]; then
    echo "STATUS: WARN — review WARN items and add PR comments explaining exceptions"
    exit 0
else
    echo "STATUS: PASS — all checks passed"
    exit 0
fi
