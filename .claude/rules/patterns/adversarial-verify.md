# Pattern: Adversarial Verification

**Use this when:** You need high confidence that findings are real, not false positives or AI hallucinations.

**The idea:** Instead of one agent verifying a finding, spawn multiple independent agents instructed to refute it. If a majority still believe the finding is real, report it.

## Core Pattern

```javascript
const finding = { title: 'Missing auth check', file: 'src/routes/api.ts' }

const votes = await parallel(
  Array.from({ length: 3 }, () => () =>
    agent(
      `Try to refute this finding: ${finding.title} in ${finding.file}.\n` +
      `If you cannot prove it's wrong, assume it's real.\n\n` +
      `Return JSON: { "refuted": boolean, "reasoning": string }`,
      { schema: VERDICT_SCHEMA }
    )
  )
)

const refuted = votes.filter(v => v.refuted).length
const survives = refuted < 2  // requires 2+ votes to refute
```

## Key Parameters

- **Number of agents:** Typically 3 (odd number for majority voting)
  - 3 agents: 2+ agree = confirmed
  - 5 agents: 3+ agree = confirmed
  - More agents = higher confidence, higher cost
  
- **Prompt framing:** "Try to refute" is stronger than "verify"
  - Default to refutation mindset
  - Makes false positives harder to survive

- **Voting threshold:** Majority or supermajority
  - 2/3 = faster convergence, some false positives survive
  - 3/5 = more conservative, higher cost

## When to Use

✅ **Use adversarial verify when:**
- Findings will trigger action (alerts, blocking, remediation)
- False positives are costly (user disruption, wasted effort)
- AI hallucinations are a concern
- You're auditing sensitive areas (security, data)

❌ **Skip when:**
- You just need to list candidates for human review
- The cost of 3 agents > value of one false positive caught
- You're in early discovery and high recall > precision

## Composing with Other Patterns

**Adversarial + Dedup:**
```javascript
const seen = new Set()
const findings = [/* initial audit */]
const deduped = findings.filter(f => !seen.has(key(f)))
const verified = await parallel(
  deduped.map(f => () => adversarialVerify(f))
)
deduped.forEach(f => seen.add(key(f)))  // mark as checked
```

**Adversarial + Loop-Until-Dry:**
```javascript
while (dryCount < 2) {
  const found = await agent('Find issues...')
  const fresh = found.filter(f => !seen.has(key(f)))
  const verified = await parallel(
    fresh.map(f => () => adversarialVerify(f))
  )
  if (verified.length === 0) dryCount++
  else dryCount = 0, verified.forEach(v => seen.add(key(v)))
}
```

## Cost Considerations

- **Single agent:** ~5 tokens/finding
- **3 adversarial agents:** ~15 tokens/finding
- **5 adversarial agents:** ~25 tokens/finding

For 100 findings: single agent = 500 tokens, 3 agents = 1500 tokens.

Use `/effort high` when running adversarial verification on large sets.

## Example: Auditing Routes

```javascript
// audit-codebase finds potential security issues
const auditFindings = [
  { file: 'src/routes/admin.ts', issue: 'no auth check on DELETE' },
  { file: 'src/routes/users.ts', issue: 'password in query param' },
]

// Each finding gets adversarially verified
const verified = await parallel(
  auditFindings.map(f => () =>
    agent(
      `Try to prove this is NOT a real security issue:\n\n` +
      `File: ${f.file}\nIssue: ${f.issue}\n\n` +
      `Return JSON: { "refuted": boolean }`,
      { schema: { type: 'object', properties: { refuted: { type: 'boolean' } } } }
    )
  )
)

// Only report findings that survived refutation
const confirmed = auditFindings.filter((_, i) =>
  verified[i]?.refuted !== true
)
```
