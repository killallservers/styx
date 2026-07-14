# Pattern: Deduplication Against Seen Set

**Use this when:** Running discovery or iteration where you want to avoid re-checking items you've already processed.

**The idea:** Maintain a Set of item keys. Before processing new items, filter out anything already in the set. This prevents wasted work in loops.

## Core Pattern

```javascript
const seen = new Set()

// Round 1: Find initial items
const round1 = await agent('Find issues...')
round1.forEach(item => seen.add(key(item)))

// Round 2: Find more, skip already-seen
const round2 = await agent('Find more issues, ignore: ...')
const fresh = round2.filter(item => !seen.has(key(item)))

if (fresh.length === 0) {
  dryCount++  // no new items = dry round
} else {
  dryCount = 0
  fresh.forEach(item => seen.add(key(item)))
}
```

## Key Parameters

- **Key function:** Must uniquely identify duplicates
  - Simple: `key(item) = item.id`
  - Compound: `key(item) = ${item.file}:${item.line}`
  - Content-based: `key(item) = hash(item.description)`

- **Storage:** Set for O(1) lookup
  - Small runs (< 1000 items): JavaScript Set in memory
  - Large runs (> 1000 items): Could use database, but Set is fine for workflow runtime

- **Hint to agent:** Tell the agent what's already found
  ```javascript
  agent(
    `Find more issues, but ignore these already found:\n` +
    `${Array.from(seen).slice(0, 20).join(', ')}\n` +
    `(showing first 20 of ${seen.size})`
  )
  ```

## When to Use

✅ **Use dedup when:**
- Running iterative searches (loop-until-converged pattern)
- Processing batches where overlap is possible
- You want to track what you've already processed
- The cost of re-checking > cost of maintaining the set

❌ **Skip when:**
- Single pass, no loops
- Items are guaranteed unique (e.g., file paths from filesystem)
- Dedup logic would be more expensive than re-checking

## Loop-Until-Dry + Dedup

The canonical pattern: iteratively search, dedup, stop when dry.

```javascript
const seen = new Set()
const allFindings = []
let dryCount = 0
let round = 0

while (dryCount < 2) {
  round++
  
  // Hint: tell agent what we've already found
  const found = await agent(
    `Round ${round}: Find flaky tests.\n` +
    `Already found: ${Array.from(seen).join(', ')}\n` +
    `Don't re-report these.`
  )
  
  // Dedup: only keep new findings
  const fresh = found.filter(f => !seen.has(f.testName))
  
  if (fresh.length === 0) {
    dryCount++
  } else {
    dryCount = 0
    fresh.forEach(f => {
      seen.add(f.testName)
      allFindings.push(f)
    })
  }
}
```

## Cost Considerations

- **Dedup overhead:** Negligible (Set lookup is O(1))
- **Benefit:** Prevents wasted agent calls
  - If 30% of findings repeat: saves 30% of agent cost in each round
  - 5 rounds × 30% savings = 150% cost reduction vs no dedup

## Example: Finding Flaky Tests

```javascript
const seen = new Set()
const flaky = []
let round = 0
let dryRounds = 0

while (dryRounds < 2 && round < 10) {
  round++
  
  const results = await agent(
    `Run tests and record failures. Round ${round}.\n` +
    `Already known flaky tests: ${Array.from(seen).join(', ')}\n` +
    `Return JSON: { "failures": [{ "testName": string, "count": number }] }`
  )
  
  // Dedup
  const newFlaky = results.failures.filter(f => !seen.has(f.testName))
  
  if (newFlaky.length === 0) {
    dryRounds++
  } else {
    dryRounds = 0
    newFlaky.forEach(f => {
      seen.add(f.testName)
      flaky.push(f)
    })
  }
  
  log(`Round ${round}: ${newFlaky.length} new flaky tests`)
}

return { flakyTests: flaky }
```

## Composing with Other Patterns

**Dedup + Adversarial Verify:**
```javascript
const found = [/* audit findings */]
const fresh = found.filter(f => !seen.has(key(f)))

// Only verify new findings, not old ones
const verified = await parallel(
  fresh.map(f => () => adversarialVerify(f))
)

fresh.forEach(f => seen.add(key(f)))
```

**Dedup + Perspective-Diverse:**
```javascript
const issues = [/* found items */]
const fresh = issues.filter(i => !seen.has(key(i)))

const reviewed = await parallel(
  fresh.map(i => () =>
    perspectiveDiverseReview(i)
  )
)

fresh.forEach(i => seen.add(key(i)))
```

## Pitfall: Dedup vs Set vs Merge

- **Dedup:** Keep first instance, discard later ones
- **Set:** Unique items (what dedup creates)
- **Merge:** Combine findings from different sources

Use dedup when:
```javascript
// Right: loop-until-dry, avoid re-checking
const fresh = new_items.filter(x => !seen.has(key(x)))
```

NOT when:
```javascript
// Wrong: you found the same issue from 2 sources
// Want to track both sources, not discard the 2nd
// Use merge instead:
const allIssues = [...round1, ...round2]  // keep duplicates with different sources
```
