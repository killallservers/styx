export const meta = {
  name: 'audit-codebase',
  description: 'Audit codebase files for issues, verify findings adversarially',
  phases: [
    { title: 'Discover', detail: 'find target files matching pattern' },
    { title: 'Audit', detail: 'audit each file in parallel' },
    { title: 'Verify', detail: 'adversarially verify each finding' },
  ],
}

// args: { pattern: 'src/routes/*.ts', issue: 'missing authentication checks' }
// or prompt user if args undefined

const pattern = args?.pattern || 'src/**/*.ts'
const issue = args?.issue || 'security issues'

phase('Discover')
const files = await agent(
  `List all files matching this glob pattern: ${pattern}\n\nReturn a JSON object with a "files" array of file paths.`,
  {
    label: 'find-files',
    schema: {
      type: 'object',
      required: ['files'],
      properties: { files: { type: 'array', items: { type: 'string' } } },
    },
  }
)

if (!files?.files?.length) {
  log(`No files found matching ${pattern}`)
  return { findings: [] }
}

log(`Found ${files.files.length} files to audit`)

phase('Audit')
const audits = await pipeline(
  files.files,
  (file) =>
    agent(
      `Audit this file for ${issue}. Return a JSON object with "issues" (array of strings) and "severity" (critical/high/medium/low).`,
      {
        label: `audit:${file}`,
        schema: {
          type: 'object',
          required: ['issues', 'severity'],
          properties: {
            issues: { type: 'array', items: { type: 'string' } },
            severity: { type: 'string', enum: ['critical', 'high', 'medium', 'low'] },
          },
        },
      }
    )
)

const allFindings = audits
  .filter(Boolean)
  .map((result, i) => ({
    file: files.files[i],
    ...result,
  }))
  .filter((f) => f.issues?.length > 0)

if (!allFindings.length) {
  log('No issues found')
  return { findings: [] }
}

log(`Found ${allFindings.length} files with issues, verifying...`)

phase('Verify')
const verified = await parallel(
  allFindings.map((finding) => async () => {
    const votes = await parallel([
      () =>
        agent(
          `Adversarially review this finding. Is it real or a false positive?\nFile: ${finding.file}\nIssues: ${finding.issues.join(', ')}\n\nReturn JSON: { "isReal": boolean, "reasoning": string }`,
          {
            label: `verify1:${finding.file}`,
            schema: {
              type: 'object',
              required: ['isReal', 'reasoning'],
              properties: {
                isReal: { type: 'boolean' },
                reasoning: { type: 'string' },
              },
            },
          }
        ),
      () =>
        agent(
          `Try to refute this finding. Is it actually a problem?\nFile: ${finding.file}\nIssues: ${finding.issues.join(', ')}\n\nReturn JSON: { "isReal": boolean, "reasoning": string }`,
          {
            label: `verify2:${finding.file}`,
            schema: {
              type: 'object',
              required: ['isReal', 'reasoning'],
              properties: {
                isReal: { type: 'boolean' },
                reasoning: { type: 'string' },
              },
            },
          }
        ),
      () =>
        agent(
          `Independently evaluate this finding.\nFile: ${finding.file}\nIssues: ${finding.issues.join(', ')}\n\nReturn JSON: { "isReal": boolean, "reasoning": string }`,
          {
            label: `verify3:${finding.file}`,
            schema: {
              type: 'object',
              required: ['isReal', 'reasoning'],
              properties: {
                isReal: { type: 'boolean' },
                reasoning: { type: 'string' },
              },
            },
          }
        ),
    ])

    const realVotes = votes.filter(Boolean).filter((v) => v.isReal).length
    const survives = realVotes >= 2

    return {
      ...finding,
      verdict: survives ? 'confirmed' : 'rejected',
      votes: realVotes,
    }
  })
)

const confirmed = verified.filter((f) => f.verdict === 'confirmed')
log(`Verified ${confirmed.length}/${allFindings.length} findings`)

return { findings: confirmed }
