export const meta = {
  name: 'loop-until-converged',
  description: 'Find issues iteratively, stop when no new findings for N rounds',
  phases: [
    { title: 'Search', detail: 'search for issues in a round' },
    { title: 'Dedup', detail: 'filter new findings vs seen set' },
    { title: 'Report', detail: 'summarize findings' },
  ],
}

// args: { task: 'Find flaky tests', rounds: 3, dryRounds: 2 }

const task = args?.task || 'Find issues'
const maxRounds = args?.rounds || 5
const dryThreshold = args?.dryRounds || 2

const seen = new Set()
const allFindings = []
let dryCount = 0
let round = 0

log(`Starting search: ${task}`)

while (round < maxRounds && dryCount < dryThreshold) {
  round++
  log(`Round ${round}/${maxRounds}`)

  phase(`Search Round ${round}`)
  const found = await agent(
    `Round ${round}: ${task}\n\nAvoid re-reporting these (already found): ${Array.from(seen).slice(0, 10).join(', ')}\n\nReturn JSON: { "findings": [{ "id": string, "description": string, "severity": string }] }`,
    {
      label: `search-round-${round}`,
      schema: {
        type: 'object',
        required: ['findings'],
        properties: {
          findings: {
            type: 'array',
            items: {
              type: 'object',
              required: ['id', 'description', 'severity'],
              properties: {
                id: { type: 'string' },
                description: { type: 'string' },
                severity: { type: 'string', enum: ['critical', 'high', 'medium', 'low'] },
              },
            },
          },
        },
      },
    }
  )

  const fresh = found.findings.filter((f) => !seen.has(f.id))

  if (fresh.length === 0) {
    dryCount++
    log(`No new findings this round (dry count: ${dryCount}/${dryThreshold})`)
  } else {
    dryCount = 0
    fresh.forEach((f) => {
      seen.add(f.id)
      allFindings.push({ ...f, foundInRound: round })
    })
    log(`Found ${fresh.length} new findings`)
  }
}

log(`Completed after ${round} rounds (${dryCount} dry rounds)`)

phase('Report')
const report = await agent(
  `Summarize these findings from iterative search:\n\n${allFindings.map((f) => `- [${f.severity}] ${f.description} (round ${f.foundInRound})`).join('\n')}\n\nReturn JSON: { "summary": string, "byCategory": object, "recommendations": string[] }`,
  {
    label: 'summarize',
    schema: {
      type: 'object',
      required: ['summary', 'byCategory', 'recommendations'],
      properties: {
        summary: { type: 'string' },
        byCategory: { type: 'object' },
        recommendations: { type: 'array', items: { type: 'string' } },
      },
    },
  }
)

return {
  report: report,
  statistics: {
    totalRounds: round,
    totalFindings: allFindings.length,
    bySeverity: {
      critical: allFindings.filter((f) => f.severity === 'critical').length,
      high: allFindings.filter((f) => f.severity === 'high').length,
      medium: allFindings.filter((f) => f.severity === 'medium').length,
      low: allFindings.filter((f) => f.severity === 'low').length,
    },
  },
}
