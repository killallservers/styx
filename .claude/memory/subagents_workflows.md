---
name: subagents-workflows-key-concepts
description: Core differences and use cases for subagents vs workflows in Claude Code
metadata: 
  node_type: memory
  type: reference
  originSessionId: 8ce738f8-7ad8-4aa1-98f6-23767ce660f1
---

## Key Concepts

### Subagents
- Specialized AI workers spawned by Claude for side tasks
- Run in isolated context window with custom system prompt, tool access, permissions
- Decision-making is Claude's: turn-by-turn what to spawn next
- Results land back in main conversation
- Best for: preserving context by offloading exploration to separate window
- Scale: a few delegated tasks per turn

### Dynamic Workflows  
- JavaScript scripts that **orchestrate many subagents**
- Runtime executes the script (not Claude) while session stays responsive
- Plan is in **code**, not in Claude's context
- Results stay in script variables, only final answer to user
- Can encode repeatable quality patterns (e.g., adversarial verification)
- Resumable within same session if paused
- Scale: dozens to hundreds of agents per run

## When to Use Each

| Factor | Subagents | Workflows |
|--------|-----------|-----------|
| Who orchestrates | Claude | Script/runtime |
| Intermediate results | Context window | Script variables |
| Repeatable | Worker definition | Orchestration itself |
| Scale | Few per turn | Many parallel agents |

## Workflow Advantages

1. **Preserves context** - intermediate results don't bloat conversation
2. **Codified plan** - save and rerun the same orchestration
3. **Quality patterns** - agents can verify each other's findings before reporting
4. **Cost control** - spawn to faster/cheaper models for specific stages
5. **Resumable** - pause and resume within session

## Workflow Triggers

- Include `ultracode` keyword in prompt
- `/effort ultracode` - enables for whole session, plans workflow for each substantive task
- `/deep-research` - bundled workflow for cross-source fact-checking

## Key Insight: The Plan Moves Into Code

With subagents/skills, Claude holds the plan (turn by turn decisions).
With workflows, **the script holds the plan** - loop, branching, dedup logic, adversarial checking.
This moves quality control from execution to orchestration.

## Common Quality Patterns

- **Adversarial verify**: N independent skeptics per finding, majority vote
- **Perspective-diverse verify**: different agents check via different lenses (correctness, security, perf)
- **Judge panel**: N independent drafts, scored and synthesized
- **Loop-until-dry**: keep searching until K rounds find nothing new
- **Multi-modal sweep**: parallel agents each searching different way
