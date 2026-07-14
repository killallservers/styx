# Pattern: Cost-Aware Scaling

**Use this when:** Running workflows with unknown scope, and you want to scale agent count based on available token budget.

**The idea:** Check the workflow budget before spawning many agents. Scale discovery to fit within the user's token budget, rather than running to completion and overspending.

## Core Pattern

```javascript
const budget = { total: null, spent: () => 0, remaining: () => Infinity }

const maxRounds = budget.total 
  ? Math.min(5, Math.floor(budget.remaining() / 50_000))
  : 5

let round = 0
while (round < maxRounds && budget.remaining() > 50_000) {
  round++
  
  const found = await agent('Find issues...')
  log(`Round ${round}: found ${found.length} items`)
  log(`Budget remaining: ${Math.round(budget.remaining() / 1000)}k tokens`)
}
```

## Key Parameters

- **Budget check:** `budget.total` is null if no user limit set
  - If null: workflow runs to completion (traditional)
  - If set: scale to fit within limit

- **Minimum threshold:** Reserve tokens for final synthesis
  - `budget.remaining() > 50_000` = at least 50k tokens left
  - Prevents running out of budget mid-workflow

- **Scaling calculation:** agents or rounds based on remaining budget
  ```javascript
  const maxAgents = budget.total
    ? Math.max(2, Math.floor(budget.remaining() / 10_000))
    : 20  // no limit
  ```

## When to Use

✅ **Use cost-aware scaling when:**
- Workflow scope is unknown (could find 10 items or 1000)
- User has set a token budget (`/effort` limits)
- You're doing discovery that could expand indefinitely
- You want graceful degradation if budget runs low

❌ **Skip when:**
- Scope is known and bounded (migrate these 50 files)
- Budget is unlimited and doesn't matter
- You must run to completion (non-negotiable)

## Budget API

```javascript
// Available globally in workflow scripts
const budget = {
  total: number | null,        // User's token limit (null if unlimited)
  spent: () => number,         // Tokens used so far this turn
  remaining: () => number,     // max(0, total - spent)
}

// Example checks
if (budget.total && budget.remaining() < 50_000) {
  log('Low budget, wrapping up')
  break  // exit loop
}

const scaledFinders = budget.total
  ? Math.max(1, Math.floor(budget.remaining() / 25_000))
  : 10
```

## Example: Flaky Test Discovery with Budget

```javascript
export const meta = {
  name: 'find-flaky-tests-budgeted',
  description: 'Find flaky tests, scale to available budget',
  phases: [{ title: 'Discover' }],
}

const seen = new Set()
const flaky = []
let round = 0

// Scale to budget: if 500k tokens available, run 5 rounds
// If 100k available, run 1 round
const maxRounds = budget.total 
  ? Math.min(10, Math.ceil(budget.remaining() / 100_000))
  : 10

log(`Budget: ${Math.round(budget.total / 1000)}k tokens, running up to ${maxRounds} rounds`)

while (round < maxRounds && budget.remaining() > 25_000) {
  round++
  
  const found = await agent(
    `Run tests and find failures. Round ${round}.\n` +
    `Already found flaky: ${Array.from(seen).join(', ')}\n` +
    `Return JSON: { "failures": [{ "testName": string }] }`
  )
  
  const fresh = found.failures.filter(f => !seen.has(f.testName))
  
  if (fresh.length === 0) break
  
  fresh.forEach(f => {
    seen.add(f.testName)
    flaky.push(f)
  })
  
  log(`Round ${round}: ${fresh.length} new flaky tests`)
  log(`Remaining budget: ${Math.round(budget.remaining() / 1000)}k tokens`)
}

return { flakyTests: Array.from(flaky) }
```

## Cost Considerations

- **High budget (500k+):** Run full discovery, multiple rounds, high verification
- **Medium budget (200k):** 3-5 rounds, selective verification
- **Low budget (50k):** 1-2 rounds, minimal overhead
- **No budget:** Run to completion (traditional)

## Composing with Other Patterns

**Cost-aware + Loop-Until-Dry:**
```javascript
const budget = arguments.budget  // implicit

while (dryCount < 2 && budget.remaining() > 50_000) {
  const found = await agent('Find issues...')
  const fresh = found.filter(f => !seen.has(key(f)))
  
  if (fresh.length === 0) dryCount++
  else dryCount = 0
  
  if (budget.remaining() < 50_000) {
    log('Budget exhausted, exiting')
    break
  }
}
```

**Cost-aware + Parallel Scaling:**
```javascript
// Scale number of parallel agents based on budget
const agentCount = budget.total
  ? Math.max(2, Math.floor(budget.remaining() / 15_000))
  : 10

const results = await parallel(
  items.slice(0, agentCount).map(item => () =>
    agent(`Process: ${item}`)
  )
)
```

## User Perspective

Users set budget with `/effort` commands:

```
/effort +200k  # Add 200k tokens to available budget
/effort +500k  # Allocate 500k for this task

# Workflow adapts:
# - Cost-aware discovery scales number of rounds
# - Parallel stages scale agent count
# - Synthesis still runs regardless (priority)
```

The workflow should degrade gracefully: with 100k budget, do 1 high-quality round. With 500k, do 5 rounds with full verification.
