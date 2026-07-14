export const meta = {
  name: 'research-question',
  description: 'Research a question across multiple sources, verify claims',
  phases: [
    { title: 'Search', detail: 'fan out searches on different angles' },
    { title: 'Fetch', detail: 'read and extract claims from sources' },
    { title: 'Verify', detail: 'cross-check claims against other sources' },
    { title: 'Synthesize', detail: 'create cited report of verified claims' },
  ],
}

// args: { question: 'How do competitors handle rate limiting?' }

const question = args?.question || 'Research this topic'

phase('Search')
const searches = await parallel([
  () =>
    agent(
      `Search for information about: ${question}\n\nReturn JSON: { "query": string, "results": [{ "title": string, "url": string, "snippet": string }] }`,
      {
        label: 'search-general',
        schema: {
          type: 'object',
          required: ['query', 'results'],
          properties: {
            query: { type: 'string' },
            results: {
              type: 'array',
              items: {
                type: 'object',
                required: ['title', 'url', 'snippet'],
                properties: {
                  title: { type: 'string' },
                  url: { type: 'string' },
                  snippet: { type: 'string' },
                },
              },
            },
          },
        },
      }
    ),
  () =>
    agent(
      `Search for recent trends about: ${question}\n\nReturn JSON: { "query": string, "results": [{ "title": string, "url": string, "snippet": string }] }`,
      {
        label: 'search-trends',
        schema: {
          type: 'object',
          required: ['query', 'results'],
          properties: {
            query: { type: 'string' },
            results: {
              type: 'array',
              items: {
                type: 'object',
                required: ['title', 'url', 'snippet'],
                properties: {
                  title: { type: 'string' },
                  url: { type: 'string' },
                  snippet: { type: 'string' },
                },
              },
            },
          },
        },
      }
    ),
  () =>
    agent(
      `Search for best practices about: ${question}\n\nReturn JSON: { "query": string, "results": [{ "title": string, "url": string, "snippet": string }] }`,
      {
        label: 'search-practices',
        schema: {
          type: 'object',
          required: ['query', 'results'],
          properties: {
            query: { type: 'string' },
            results: {
              type: 'array',
              items: {
                type: 'object',
                required: ['title', 'url', 'snippet'],
                properties: {
                  title: { type: 'string' },
                  url: { type: 'string' },
                  snippet: { type: 'string' },
                },
              },
            },
          },
        },
      }
    ),
])

const allResults = searches.filter(Boolean).flatMap((s) => s.results)
const uniqueUrls = [...new Set(allResults.map((r) => r.url))]

log(`Found ${uniqueUrls.length} unique sources across ${searches.length} search angles`)

phase('Fetch')
const sources = await pipeline(
  uniqueUrls.slice(0, 10),
  (url) =>
    agent(
      `Fetch and summarize this source:\n\nURL: ${url}\n\nExtract key claims and facts. Return JSON: { "url": string, "title": string, "claims": string[] }`,
      {
        label: `fetch:${url}`,
        schema: {
          type: 'object',
          required: ['url', 'title', 'claims'],
          properties: {
            url: { type: 'string' },
            title: { type: 'string' },
            claims: { type: 'array', items: { type: 'string' } },
          },
        },
      }
    )
)

const validSources = sources.filter(Boolean)
const allClaims = validSources.flatMap((s) => s.claims.map((c) => ({ claim: c, sources: [s.url] })))

log(`Extracted ${allClaims.length} claims from ${validSources.length} sources`)

phase('Verify')
const verified = await parallel(
  allClaims.map((item) => async () => {
    const checks = await parallel(
      validSources
        .filter((s) => !item.sources.includes(s.url))
        .slice(0, 3)
        .map((source) => () =>
          agent(
            `Does this claim appear in the content from ${source.title}?\n\nClaim: ${item.claim}\n\nReturn JSON: { "found": boolean, "supporting": boolean }`,
            {
              label: `check-claim:${source.url}`,
              schema: {
                type: 'object',
                required: ['found', 'supporting'],
                properties: {
                  found: { type: 'boolean' },
                  supporting: { type: 'boolean' },
                },
              },
            }
          )
        )
    )

    const supported = checks.filter(Boolean).filter((c) => c.supporting).length
    const isVerified = supported >= 1

    return {
      claim: item.claim,
      verified: isVerified,
      sources: item.sources,
      supportingChecks: supported,
    }
  })
)

const confirmedClaims = verified.filter((c) => c.verified)
log(`Verified ${confirmedClaims.length}/${verified.length} claims`)

phase('Synthesize')
const report = await agent(
  `Create a research report answering: ${question}\n\nBased on these verified claims:\n${confirmedClaims.map((c) => `- ${c.claim} (${c.sources.join(', ')})`).join('\n')}\n\nReturn JSON: { "title": string, "summary": string, "sections": [{ "heading": string, "content": string }], "sources": string[] }`,
  {
    label: 'synthesize-report',
    schema: {
      type: 'object',
      required: ['title', 'summary', 'sections', 'sources'],
      properties: {
        title: { type: 'string' },
        summary: { type: 'string' },
        sections: {
          type: 'array',
          items: {
            type: 'object',
            required: ['heading', 'content'],
            properties: {
              heading: { type: 'string' },
              content: { type: 'string' },
            },
          },
        },
        sources: { type: 'array', items: { type: 'string' } },
      },
    },
  }
)

return {
  report: report,
  statistics: {
    sourcesChecked: validSources.length,
    claimsExtracted: allClaims.length,
    claimsVerified: confirmedClaims.length,
  },
}
