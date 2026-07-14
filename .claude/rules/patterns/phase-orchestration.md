# Pattern: Phase-Based Orchestration

**Use this when:** Deciding how to run agents across stages—should stages run in parallel (no barrier) or sequentially (with barrier)?

**The idea:** Use `pipeline()` for no-barrier streaming (stages run as soon as items enter them). Use `parallel()` for barriers (all items finish stage N before stage N+1 starts).

## Core Concepts

### Pipeline (No Barrier): Streaming Execution

Each item flows through all stages independently. Item A can be in stage 3 while item B is still in stage 1.

```javascript
const results = await pipeline(
  items,
  stage1: (item) => agent('Discover..'),
  stage2: (result, item) => agent('Audit..'),
  stage3: (result, item) => agent('Verify..'),
)

// Wall-clock time = slowest single item's total time
// If item 1 takes 3s per stage (9s total)
// and item 2 takes 2s per stage (6s total)
// total wall-clock: 9s (not 15s)
```

**Use pipeline when:**
- Stages are independent per item
- You want max parallelism
- Later stages don't need ALL prior results
- Wall-clock speed matters

### Parallel (Barrier): Synchronized Execution

All items finish stage N before any start stage N+1. Creates a barrier.

```javascript
const stage1 = await parallel(
  items.map(item => () => agent('Discover...'))
)

const deduped = dedupeAcrossAll(stage1)  // BARRIER: needs all stage1 results

const stage2 = await parallel(
  deduped.map(item => () => agent('Audit...'))
)

// Wall-clock time = sum of slowest item per stage
// Stage 1: slowest item takes 3s = 3s total
// Stage 2: slowest item takes 2s = 2s total
// total wall-clock: 5s (but could be up to 15s if sequential)
```

**Use parallel when:**
- Stage N needs ALL of stage N-1 results
- You're deduping across items
- You're merging/filtering/ranking
- Correctness > speed

## Decision Tree

```
Does stage N+1 need cross-item context from stage N?
  ├─ YES (merging, dedup, ranking)
    └─ Use parallel() → barrier after stage N
  └─ NO (each item processed independently)
    └─ Use pipeline() → no barrier
```

## Common Patterns

### Pattern 1: Find → Dedup → Verify

**Pipeline until dedup (no cross-item ops), then parallel.**

```javascript
// Stage 1: Each item audited independently
const audits = await pipeline(
  files,
  (file) => agent(`Audit ${file}...`)
)

// Barrier: dedup needs ALL audit results
const deduped = Array.from(new Set(
  audits.flatMap(a => a.issues).map(i => key(i))
)).map(k => issues.find(i => key(i) === k))

// Stage 2: Verify deduplicated issues
const verified = await parallel(
  deduped.map(issue => () =>
    agent(`Verify ${issue}...`)
  )
)
```

**Wall-clock:** find time + dedup time + verify time
**NOT:** find time × number_of_files

### Pattern 2: Search → Fetch → Cross-Check → Synthesize

**Pipeline with occasional barriers for critical dedup.**

```javascript
const searches = await parallel([
  () => agent('Search angle 1'),
  () => agent('Search angle 2'),
  () => agent('Search angle 3'),
])

// Barrier: collect all search results
const allResults = searches.flatMap(s => s.results)
const uniqueUrls = [...new Set(allResults.map(r => r.url))]

// Pipeline: fetch each URL (independent)
const fetched = await pipeline(
  uniqueUrls,
  (url) => agent(`Fetch and summarize ${url}`)
)

// Barrier: cross-check needs all fetched content
const verified = await parallel(
  allResults.map(result => () =>
    agent(`Check if "${result.claim}" supported by ${fetched.length} sources`)
  )
)
```

**Wall-clock:** search + fetch + verify (not search + fetch×N + verify)

### Pattern 3: Audit → Verify (Adversarial)

**Can use pipeline with built-in verification per item.**

```javascript
const results = await pipeline(
  files,
  // Stage 1: Audit
  (file) => agent(`Audit ${file}`),
  
  // Stage 2: Verify (per item, not cross-item)
  (auditResult, file) =>
    parallel([
      () => agent('Refute issue 1'),
      () => agent('Refute issue 2'),
      () => agent('Refute issue 3'),
    ])
      .then(votes => ({ file, auditResult, votes }))
)

// Result: each file fully processed (audit + triple-verification)
// before next file starts. But file 2 can start audit
// while file 1 is still in verification stage.
```

**Wall-clock:** Max(audit + 3×verify per file across all files)

## Cost Considerations

### Pipeline (No Barrier)

- **Pros:** Max parallelism, fastest wall-clock
- **Cons:** Results in script, not in context (use `args` to fetch)
- **Use when:** Speed matters, stages are independent

```javascript
// 100 items × 3 stages = up to 300 agents in flight
// But wall-clock: max time of 1 item through 3 stages
```

### Parallel (Barrier)

- **Pros:** All results available for cross-stage logic
- **Cons:** Slower wall-clock if early stages are fast
- **Use when:** Correctness matters, you need dedup/merge

```javascript
// 100 items, stage 1: 10 agents in parallel (100/10 batches)
// After stage 1 finishes, stage 2 starts
// Wall-clock: slower, but logic is clear
```

## Example: The Right Choice

**Audit codebase with dedup + adversarial verify:**

```javascript
// WRONG: Full barrier after audit
const audits = await parallel(
  files.map(f => () => agent(`Audit ${f}`))
)
const deduped = dedup(audits)
const verified = await parallel(
  deduped.map(i => () => adversarialVerify(i))
)
// Wall-clock: audit + dedup + verify (sequential barrier style)

// RIGHT: Pipeline until dedup, then barrier
const audits = await pipeline(
  files,
  (f) => agent(`Audit ${f}`)
)
const deduped = dedup(audits)  // collect all for dedup
const verified = await parallel(
  deduped.map(i => () => adversarialVerify(i))
)
// Wall-clock: max(audit per file) + dedup + verify
// Faster because files audit in parallel
```

## Composing Barriers and Pipelines

**Multi-barrier pattern:**

```javascript
// Barrier 1: collect all searches
const searches = await parallel([...])  // 3 agents

// Pipeline: fetch (independent)
const fetches = await pipeline(
  results.flatMap(r => r.urls),
  (url) => agent(`Fetch...`)
)

// Barrier 2: cross-check (needs all fetches)
const checks = await parallel(
  fetches.map(f => () => agent(`Cross-check...`))
)

// Pipeline: synthesis (could be per-topic or single)
const synthesis = await agent('Synthesize all checks...')
```

**Rationale:**
- Searches have overhead → barrier worth it (3 agents)
- Fetches are independent → pipeline saves wall-clock
- Checks need all fetches → barrier necessary
- Synthesis is single agent → no parallel benefit

## Red Flags

❌ **Using parallel() when you should pipeline():**
```javascript
// WRONG: Creates unnecessary barrier
const audits = await parallel(
  files.map(f => () => agent(`Audit ${f}`))
)
// This is slower than pipeline if stages are independent
```

❌ **Using pipeline() when you need cross-item logic:**
```javascript
// WRONG: Dedup has no visibility into all results
const audits = await pipeline(
  files,
  (f, result) => {
    // Can't dedup here: only have current result
    agent(`Audit ${f}`)
  }
)
```

✅ **Right pattern:**
```javascript
// Collect all first (barrier or pipeline into flat array)
const audits = await pipeline(files, (f) => agent(`Audit ${f}`))

// Then dedup with full visibility
const deduped = dedup(audits.flatMap(a => a.issues))

// Then verify in parallel
const verified = await parallel(
  deduped.map(issue => () => agent(`Verify ${issue}`))
)
```
