export const meta = {
  name: 'migrate-in-parallel',
  description: 'Migrate files in parallel using isolated worktrees, verify results',
  phases: [
    { title: 'Discover', detail: 'find files matching pattern' },
    { title: 'Migrate', detail: 'transform each file in isolated worktree' },
    { title: 'Verify', detail: 'verify each migration result' },
  ],
}

// args: { pattern: 'src/**/*.ts', task: 'replace old-lib with new-lib' }

const pattern = args?.pattern || 'src/**/*.ts'
const task = args?.task || 'update code'

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
  return { results: [] }
}

log(`Found ${files.files.length} files to migrate`)

phase('Migrate')
const migrations = await pipeline(
  files.files,
  (file) =>
    agent(
      `Migrate this file: ${task}\n\nFile: ${file}\n\nReturn JSON: { "migrated": boolean, "changes": string, "errors": string[] }`,
      {
        label: `migrate:${file}`,
        isolation: 'worktree',
        schema: {
          type: 'object',
          required: ['migrated', 'changes', 'errors'],
          properties: {
            migrated: { type: 'boolean' },
            changes: { type: 'string' },
            errors: { type: 'array', items: { type: 'string' } },
          },
        },
      }
    )
)

const allResults = migrations
  .filter(Boolean)
  .map((result, i) => ({
    file: files.files[i],
    ...result,
  }))

const successful = allResults.filter((r) => r.migrated)
const failed = allResults.filter((r) => !r.migrated)

log(`Migrated ${successful.length}/${allResults.length} files`)

if (failed.length > 0) {
  log(`${failed.length} migrations failed:`)
  failed.forEach((f) => log(`  - ${f.file}: ${f.errors.join(', ')}`))
}

phase('Verify')
const verified = await parallel(
  successful.map((result) => async () => {
    const checks = await parallel([
      () =>
        agent(
          `Verify this migration result:\nFile: ${result.file}\nChanges: ${result.changes}\n\nReturn JSON: { "correct": boolean, "issues": string[] }`,
          {
            label: `verify-syntax:${result.file}`,
            schema: {
              type: 'object',
              required: ['correct', 'issues'],
              properties: {
                correct: { type: 'boolean' },
                issues: { type: 'array', items: { type: 'string' } },
              },
            },
          }
        ),
      () =>
        agent(
          `Check if this migration maintains semantics:\nFile: ${result.file}\nChanges: ${result.changes}\n\nReturn JSON: { "correct": boolean, "issues": string[] }`,
          {
            label: `verify-logic:${result.file}`,
            schema: {
              type: 'object',
              required: ['correct', 'issues'],
              properties: {
                correct: { type: 'boolean' },
                issues: { type: 'array', items: { type: 'string' } },
              },
            },
          }
        ),
    ])

    const passed = checks.filter(Boolean).filter((c) => c.correct).length
    const allPassed = passed === checks.filter(Boolean).length

    return {
      ...result,
      verified: allPassed,
      checks: passed,
    }
  })
)

const verifiedOk = verified.filter((r) => r.verified)
log(`Verified ${verifiedOk.length}/${successful.length} migrations passed checks`)

return {
  results: allResults,
  summary: {
    total: allResults.length,
    migrated: successful.length,
    verified: verifiedOk.length,
    failed: failed.length,
  },
}
