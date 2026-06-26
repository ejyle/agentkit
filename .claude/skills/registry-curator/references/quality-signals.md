# Quality Signals and Ranking Rubric

This rubric produces a quality score from 0 to 100 for each discovered skill. Apply
it after deduplication so that each skill is scored once from its highest-trust source.

## Scoring Formula

```
Score = (stars_factor * 40) + (recency_factor * 35) + (download_factor * 15) + (security_factor * 10)
```

Maximum score is 100. Each factor is a decimal from 0.0 to 1.0.

## Factor Definitions

### stars_factor (weight: 40)

Computed as `log10(stars + 1) / 5`. This gives approximately:

| Stars | log10(stars+1) | stars_factor |
|-------|---------------|--------------|
| 0 | 0.0 | 0.00 |
| 100 | 2.00 | 0.40 |
| 1,000 | 3.00 | 0.60 |
| 10,000 | 4.00 | 0.80 |
| 100,000 | 5.00 | 1.00 |

**CAVEAT — multi-skill repos:** For skills hosted in shared repositories such as
`anthropics/skills`, the repo's total star count reflects source trust, not individual
skill popularity. Use a flat `stars_factor = 0.9` for all entries from `anthropics/skills`.
Do not apply the log formula in this case.

For GitHub-only skills with individual repos (e.g., `obra/superpowers`), use the
actual repo star count in the log formula.

### recency_factor (weight: 35)

Based on the date of the last commit to the skill directory or repo:

| Last Commit Age | recency_factor |
|-----------------|---------------|
| Within 30 days | 1.0 |
| Within 90 days | 0.7 |
| Within 365 days | 0.3 |
| Older than 1 year | 0.0 |
| Unknown / unavailable | 0.1 |

When the last commit date refers to the containing repo rather than the specific skill
subdirectory, use the repo-level date as a conservative proxy.

### download_factor (weight: 15)

Based on monthly install or download count where available:

| Monthly Downloads | download_factor |
|------------------|----------------|
| >= 10,000 | 1.0 |
| >= 1,000 | 0.5 |
| >= 100 | 0.2 |
| < 100 or unavailable | 0.0 |

GitHub-only skills commonly have no download statistics. Set `download_factor = 0.0`
and do **not** exclude them — many high-quality skills are GitHub-only.

### security_factor (weight: 10)

Based on the source registry's trust tier:

| Source | security_factor |
|--------|----------------|
| anthropics/skills | 1.0 |
| skillsdirectory.com (grade-A vetted) | 1.0 |
| agentskills.io | 0.5 |
| mcpmarket.com | 0.5 |
| travisvn/awesome-claude-skills | 0.5 |
| VoltAgent/awesome-agent-skills | 0.5 |
| Web search result or unknown origin | 0.0 |

## Example Scores

| Skill | stars_factor | recency_factor | download_factor | security_factor | Score |
|-------|-------------|---------------|----------------|----------------|-------|
| anthropics/skills — docx | 0.90 (flat) | 0.70 (90d) | 0.0 (no data) | 1.0 (official) | 0.90×40 + 0.70×35 + 0.0×15 + 1.0×10 = **70.5** |
| travisvn/awesome-claude-skills — git-helper | 0.52 (3.2K stars) | 0.70 (60d) | 0.2 (250/mo) | 0.5 (curated) | 0.52×40 + 0.70×35 + 0.2×15 + 0.5×10 = **56.3** |
| Web-search-found — ai-test-helper | 0.32 (1K stars) | 0.30 (200d) | 0.0 (no data) | 0.0 (web search) | 0.32×40 + 0.30×35 + 0.0×15 + 0.0×10 = **23.3** |

## Presentation Guidance

Display quality score as an integer (round to nearest whole number). Prefix with a label:

| Range | Label | Example |
|-------|-------|---------|
| >= 70 | HIGH | `HIGH 88` |
| 40 - 69 | MED | `MED 57` |
| < 40 | LOW | `LOW 22` |

Always include Stars, Last Commit, Source, and Score columns in the ranked table.

Ranked table column order: Rank, Skill Name, Source, Stars, Last Commit, Score, Install Hint.

## Edge Cases

**Skill with no GitHub repo:** Set `stars_factor = 0.0` and `download_factor = 0.0`. Only
qualifies for inclusion if the source is `skillsdirectory.com` or `anthropics/skills`, where
the security_factor compensates. Such skills will score LOW unless they have strong recency.

**Last commit date unavailable:** Set `recency_factor = 0.1`. Do not exclude the skill — it
may be stable and widely used.

**Same skill name from two registries:** Keep the highest-scoring entry. Note the alternate
source in parentheses in the table's Notes column (e.g., "also on skillsdirectory.com").
Apply the security_factor of the kept source, not the alternate.

**Star count not visible (private or brand-new repo):** Set `stars_factor = 0.0`. Rely on
recency and security factors. Such skills are unlikely to reach HIGH tier unless they come
from `anthropics/skills`.

**Skill version mismatch (SKILL.md frontmatter version differs from registry listing):**
Use the version from SKILL.md (authoritative). Log a warning in the summary but do not
penalize the quality score.

**Multiple installs of same skill at different paths in the same repo:** Treat each
subdirectory as a separate candidate with the same stars and security_factor but potentially
different recency if subdirectory commit history is available.
