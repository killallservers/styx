export const meta = {
  name: 'judge-panel',
  description: 'Generate solutions from multiple perspectives, judge and synthesize',
  phases: [
    { title: 'Draft', detail: 'N agents draft solutions from different angles' },
    { title: 'Judge', detail: 'independently score each draft' },
    { title: 'Synthesize', detail: 'combine winner with best ideas from runners-up' },
  ],
}

// args: { decision: 'Should we migrate to Drizzle ORM?' }

const decision = args?.decision || 'Make a decision'

phase('Draft')
const angles = [
  { name: 'MVP-first', prompt: 'Prioritize shipping the fastest, simplest solution' },
  { name: 'Risk-first', prompt: 'Prioritize minimizing technical risk and debt' },
  { name: 'Cost-first', prompt: 'Prioritize minimizing cost and resource usage' },
  { name: 'User-first', prompt: 'Prioritize the user experience and feature velocity' },
]

const drafts = await parallel(
  angles.map((angle) => () =>
    agent(
      `${angle.prompt}\n\nDecision: ${decision}\n\nProvide your recommendation. Return JSON: { "recommendation": string, "reasoning": string, "tradeoffs": string[], "score": number (1-10) }`,
      {
        label: `draft:${angle.name}`,
        schema: {
          type: 'object',
          required: ['recommendation', 'reasoning', 'tradeoffs', 'score'],
          properties: {
            recommendation: { type: 'string' },
            reasoning: { type: 'string' },
            tradeoffs: { type: 'array', items: { type: 'string' } },
            score: { type: 'number' },
          },
        },
      }
    )
)

const validDrafts = drafts.filter(Boolean).map((d, i) => ({
  angle: angles[i]?.name,
  ...d,
}))

log(`Generated ${validDrafts.length} drafts`)

phase('Judge')
const judgments = await parallel(
  validDrafts.map((draft) => () =>
    agent(
      `Independently evaluate this solution:\n\nAngle: ${draft.angle}\nRecommendation: ${draft.recommendation}\nReasoning: ${draft.reasoning}\n\nScore it (1-10) and identify strengths/weaknesses. Return JSON: { "score": number, "strengths": string[], "weaknesses": string[] }`,
      {
        label: `judge:${draft.angle}`,
        schema: {
          type: 'object',
          required: ['score', 'strengths', 'weaknesses'],
          properties: {
            score: { type: 'number' },
            strengths: { type: 'array', items: { type: 'string' } },
            weaknesses: { type: 'array', items: { type: 'string' } },
          },
        },
      }
    )
  )
)

const scored = validDrafts.map((draft, i) => ({
  ...draft,
  judgment: judgments[i],
  finalScore: (draft.score + (judgments[i]?.score || 0)) / 2,
}))

const sortedByScore = scored.sort((a, b) => b.finalScore - a.finalScore)
const winner = sortedByScore[0]
const runnersUp = sortedByScore.slice(1)

log(`Winner: ${winner.angle} (score: ${winner.finalScore.toFixed(1)})`)
log(`Runners-up: ${runnersUp.map((r) => `${r.angle} (${r.finalScore.toFixed(1)})`).join(', ')}`)

phase('Synthesize')
const synthesis = await agent(
  `Synthesize the best decision combining these perspectives:\n\nWinner: ${winner.angle}\n${winner.recommendation}\n\nRunners-up ideas to consider:\n${runnersUp.map((r) => `- [${r.angle}] Best idea: ${r.judgment?.strengths?.[0] || 'N/A'}`).join('\n')}\n\nReturn JSON: { "decision": string, "rationale": string, "incorporatedIdeas": string[], "implementation": string, "risks": string[] }`,
  {
    label: 'synthesize',
    schema: {
      type: 'object',
      required: ['decision', 'rationale', 'incorporatedIdeas', 'implementation', 'risks'],
      properties: {
        decision: { type: 'string' },
        rationale: { type: 'string' },
        incorporatedIdeas: { type: 'array', items: { type: 'string' } },
        implementation: { type: 'string' },
        risks: { type: 'array', items: { type: 'string' } },
      },
    },
  }
)

return {
  decision: synthesis,
  allDrafts: scored,
  winner: {
    angle: winner.angle,
    score: winner.finalScore,
    recommendation: winner.recommendation,
  },
}
