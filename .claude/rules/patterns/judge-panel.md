# Pattern: Judge Panel with Synthesis

**Use this when:** Making a decision where multiple valid approaches exist, and you want a well-reasoned choice.

**The idea:** Have N agents draft solutions from different angles (MVP-first, risk-first, cost-first, user-first). Independently judge each draft. Synthesize by combining the winner with best ideas from runners-up.

## Core Pattern

```javascript
// Step 1: Draft from different angles
const angles = [
  { name: 'MVP-first', prompt: 'Prioritize shipping fast' },
  { name: 'Risk-first', prompt: 'Prioritize minimizing risk' },
  { name: 'Cost-first', prompt: 'Prioritize minimizing cost' },
]

const drafts = await parallel(
  angles.map(angle => () =>
    agent(
      `${angle.prompt}. Decision: ${question}\n` +
      `Return JSON: { "recommendation": string, "reasoning": string, "score": number }`,
      { schema: DRAFT_SCHEMA }
    )
  )
)

// Step 2: Judge each draft independently
const judgments = await parallel(
  drafts.map(draft => () =>
    agent(
      `Evaluate this independently:\n${draft.recommendation}\n` +
      `Return JSON: { "score": number, "strengths": string[], "weaknesses": string[] }`,
      { schema: JUDGMENT_SCHEMA }
    )
  )
)

// Step 3: Synthesize winner + best ideas
const scored = drafts.map((d, i) => ({
  ...d,
  judgment: judgments[i],
  finalScore: (d.score + judgments[i].score) / 2,
}))

const winner = scored.sort((a, b) => b.finalScore - a.finalScore)[0]
const runnersUp = scored.slice(1)

const synthesis = await agent(
  `Combine this winner with best ideas from runners-up:\n` +
  `Winner: ${winner.recommendation}\n` +
  `Best ideas: ${runnersUp.map(r => r.judgment.strengths[0]).join(', ')}\n` +
  `Return JSON: { "decision": string, "incorporatedIdeas": string[] }`
)
```

## Key Parameters

- **Number of angles:** 3-5 is standard
  - 2 angles: too narrow
  - 3 angles: good balance (MVP, Risk, Cost)
  - 4-5 angles: covers more perspectives (add User, Ops, etc.)

- **Scoring:** Average of self-score + independent judge
  - Self-score: agent's confidence (1-10)
  - Judge score: independent evaluation (1-10)
  - Avoids one judge dominating

- **Synthesis:** Take winner + 1-2 best ideas from runners-up
  - Don't try to merge all angles (creates Frankenstein)
  - Pick the best single approach, then steal what works

## When to Use

✅ **Use judge panel when:**
- No obvious "right answer"
- Different stakeholders have different priorities
- You want documented tradeoffs
- The decision is reversible but expensive to change
- You need to justify the choice later

❌ **Skip when:**
- One approach is clearly dominant
- You just need a quick recommendation
- The decision is low-stakes
- Budget/tokens are very limited

## Common Angles for Different Domains

**For architecture/system design:**
- MVP-first: Minimize time to market
- Risk-first: Minimize technical risk
- Cost-first: Minimize infrastructure cost
- Scale-first: Prioritize scalability

**For refactoring/migrations:**
- Speed-first: Fastest implementation
- Correctness-first: Most robust approach
- Maintainability-first: Easiest to understand long-term
- Compatibility-first: Minimize breaking changes

**For feature prioritization:**
- User-first: Highest user impact
- Revenue-first: Highest business impact
- Tech-debt-first: Enables future work
- Risk-first: Reduces technical risk

## Cost Considerations

- **Drafting:** N agents × ~10 tokens = 10N tokens
- **Judging:** N agents × ~10 tokens = 10N tokens
- **Synthesis:** 1 agent × ~15 tokens = 15 tokens
- **Total for 3 angles:** 60 + 15 = 75 tokens

**vs. Single agent:** ~20 tokens. So judge panel is ~4x more expensive, but gives you documented tradeoffs and higher confidence.

Use when the decision's value > 4× cost of agent work.

## Example: ORM Migration Decision

```javascript
const decision = 'Should we migrate from TypeORM to Drizzle?'

const drafts = await parallel([
  () => agent(
    `Prioritize shipping fast. Should we migrate from TypeORM to Drizzle?\n` +
    `Consider implementation speed, team learning curve, risk of breaking changes.\n` +
    `Return JSON: { "recommendation": string, "score": number }`
  ),
  () => agent(
    `Prioritize minimizing risk. Should we migrate from TypeORM to Drizzle?\n` +
    `Consider type safety, query correctness, tooling maturity.\n` +
    `Return JSON: { "recommendation": string, "score": number }`
  ),
  () => agent(
    `Prioritize user experience and feature velocity. Should we migrate?\n` +
    `Consider ease of use, validation integration, query patterns.\n` +
    `Return JSON: { "recommendation": string, "score": number }`
  ),
])

// Each gets independently scored
const judgments = await parallel(
  drafts.map(d => () =>
    agent(
      `Evaluate: ${d.recommendation}. Strengths and weaknesses?\n` +
      `Return JSON: { "score": number, "strengths": string[], "weaknesses": string[] }`
    )
  )
)

// Synthesize
const scored = drafts.map((d, i) => ({
  ...d,
  finalScore: (d.score + judgments[i].score) / 2,
}))

const best = scored.sort((a, b) => b.finalScore - a.finalScore)

return {
  winner: best[0].recommendation,
  rationale: 'Drizzle + TypeORM hybrid during migration',
  incorporatedIdeas: best.slice(1).map(d => d.judgment.strengths[0]),
}
```

## Composing with Other Patterns

**Judge panel + research:**
```javascript
// Research the state of each option
const research = await agent('Compare ORM options...')

// Then let judge panel evaluate them
const drafts = await parallel([
  () => agent('Based on this research, MVP-first approach...'),
  () => agent('Based on this research, risk-first approach...'),
])
```

**Judge panel + adversarial verify:**
```javascript
// Synthesize best option
const synthesized = await agent('Combine winner with best ideas...')

// Then adversarially verify the synthesis
const verdict = await agent(
  'Try to find a fatal flaw in this decision...'
)
```
