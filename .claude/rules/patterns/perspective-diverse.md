# Pattern: Perspective-Diverse Review

**Use this when:** A finding can fail for multiple reasons, and you want to catch all failure modes.

**The idea:** Instead of redundantly cloning the same verifier N times, assign each verifier a distinct lens (correctness, security, performance, compatibility). Each brings different expertise to the evaluation.

## Core Pattern

```javascript
const finding = { title: 'Database query', description: '...' }

const reviews = await parallel([
  () => agent(
    `Review for CORRECTNESS: Does this work as intended? ${finding.description}`,
    { schema: VERDICT_SCHEMA }
  ),
  () => agent(
    `Review for SECURITY: Are there injection, data exposure, or access control issues? ${finding.description}`,
    { schema: VERDICT_SCHEMA }
  ),
  () => agent(
    `Review for PERFORMANCE: Is this slow, does it scale, does it leak resources? ${finding.description}`,
    { schema: VERDICT_SCHEMA }
  ),
])

const passed = reviews.filter(v => v.isValid).length
const allPassed = passed === reviews.length
```

## Key Parameters

- **Number of lenses:** 3-5 is ideal
  - 2 lenses: too narrow, miss issues
  - 3 lenses: covers common failure modes
  - 5 lenses: diminishing returns (very expensive)

- **Which lenses matter?** Depends on domain:
  - Code review: correctness, security, performance, maintainability
  - Architecture: scalability, reliability, cost, user experience
  - Data: accuracy, completeness, biases, consent
  - Systems: latency, throughput, availability, security

- **Voting:** Different than adversarial
  - Majority vote (2/3 passed checks = valid)
  - All-pass (all 3 must pass = strict)
  - Supermajority (4/5 must pass = balanced)

## When to Use

✅ **Use perspective-diverse when:**
- The problem domain has orthogonal failure modes
- Different reviewers bring different expertise
- You're evaluating something with multi-faceted quality (code, architecture)
- You want to avoid "everyone agrees because they use the same reasoning"

❌ **Skip when:**
- All your lenses ask the same question
- You just need quick vetting (use single agent)
- The failure modes are unrelated to each perspective

## vs. Adversarial Verify

**Perspective-diverse:**
- Multiple viewpoints, each looking for different problems
- "Can you find a security issue? A performance issue?"
- Breadth of evaluation

**Adversarial:**
- Multiple skeptics, all trying to refute the same claim
- "Try to prove this isn't a problem"
- Depth of evaluation for one claim

**Use both:**
```javascript
// First: perspective-diverse review to find issues
const reviews = await parallel([
  () => agent('Review for correctness...'),
  () => agent('Review for security...'),
  () => agent('Review for performance...'),
])

// Then: adversarially verify each found issue
const found = reviews.flatMap(r => r.issues)
const verified = await parallel(
  found.map(issue => () =>
    agent(`Try to refute: ${issue}`, { schema: VERDICT })
  )
)
```

## Cost Considerations

- **Single lens (1 agent):** ~5 tokens
- **3 lenses (3 agents):** ~15 tokens
- **5 lenses (5 agents):** ~25 tokens

For 50 findings: single = 250 tokens, 3 lenses = 750 tokens, 5 lenses = 1250 tokens.

Perspective-diverse is cheaper than redundant adversarial (different questions, not repeated questions).

## Example: Code Review

```javascript
const codeSnippet = `
  const users = await db.query(
    'SELECT * FROM users WHERE id = ' + userId
  )
`

const reviews = await parallel([
  () => agent(
    `Does this code work correctly? Trace through it: ${codeSnippet}`,
    { schema: { type: 'object', properties: { correct: { type: 'boolean' } } } }
  ),
  () => agent(
    `Is there a SQL injection vulnerability here? ${codeSnippet}`,
    { schema: { type: 'object', properties: { vulnerable: { type: 'boolean' } } } }
  ),
  () => agent(
    `Will this scale well with large datasets? ${codeSnippet}`,
    { schema: { type: 'object', properties: { performant: { type: 'boolean' } } } }
  ),
])

const correctness = reviews[0]?.correct
const security = !reviews[1]?.vulnerable  // inverted: no vuln = good
const performance = reviews[2]?.performant

const allPass = correctness && security && performance
```

## Composing with Other Patterns

**Perspective-diverse + Dedup:**
```javascript
const issues = [/* found items */]
const seen = new Set()
const fresh = issues.filter(i => !seen.has(key(i)))

const reviews = await parallel(
  fresh.map(issue => () =>
    perspectiveDiverseReview(issue)
  )
)

fresh.forEach(i => seen.add(key(i)))
```

**Perspective-diverse in Judge Panel:**
```javascript
// Draft solutions
const drafts = [...]

// Judge each draft using perspective-diverse review
const judgments = await parallel(
  drafts.map(draft => () =>
    agent(
      `Review this draft from multiple angles:\n` +
      `Correctness: Does it solve the problem?\n` +
      `Risk: Does it introduce new risks?\n` +
      `Cost: Is it resource-efficient?\n` +
      `${draft.recommendation}`,
      { schema: JUDGMENT_SCHEMA }
    )
  )
)
```
