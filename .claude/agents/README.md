# Development personas

Seven subagents drive iterative development of this game. Each owns one concern and
one part of the repository. The main session acts as orchestrator: it picks the
persona, hands it a focused prompt, and relays the result.

## Terms

- **Vertical slice:** the smallest playable version of the whole game, defined in
  `docs/design/vision.md`. Specs tagged *Slice* there belong to it.
- **Milestone:** one step on the roadmap (`docs/plan/roadmap.md`) toward the
  vertical slice or beyond it.
- **Iteration:** one plan, build and QA cycle that ends in a playable build. A
  milestone takes one or more iterations.

## The iteration loop

1. **game-designer** writes or revises the spec for the iteration.
2. **producer** scopes the iteration and writes the iteration plan.
3. **architect** confirms the technical approach, writing an ADR if a decision is new.
4. **gameplay-engineer** and **rendering-ux-engineer** implement against the spec and plan.
5. **code-reviewer** reviews the diff and reports findings. Authors fix them.
6. **qa-playtester** tests and plays the build, then files a report.
7. **producer** closes the iteration, recording what shipped and what was deferred,
   and updates the roadmap. Deferred items feed the next plan.

Author, reviewer and tester are deliberately separate personas. Do not let one
persona do another's job in a single invocation.

Scope is set by the plan and does not change mid-iteration. Anything a persona
discovers that is not in the plan is recorded as deferred in its report and
considered when the producer writes the next plan. Only the user edits an open plan.

## Shared file conventions

| Path | Owner | Contents |
|---|---|---|
| `docs/design/vision.md` | game-designer | Pillars, core loop, player fantasy, non-goals |
| `docs/design/mechanics/<system>.md` | game-designer | One spec per game system |
| balance configuration file (path set by architect) | game-designer | Every tunable value, each with a doc comment stating what it is, why, and its safe range |
| `docs/plan/roadmap.md` | producer | Milestones in order |
| `docs/plan/iterations/NNN-<slug>.md` | producer | Scope, acceptance criteria and outcome for one iteration |
| `docs/architecture.md` | architect | Module map and boundaries |
| `docs/adr/NNNN-<slug>.md` | architect | One decision per file, never edited after acceptance |
| `docs/qa/reports/NNN-<slug>.md` | qa-playtester | Test and playtest findings for one iteration |

Code layout is decided by the architect in the first ADR and recorded in
`docs/architecture.md`. Every engineer reads that file before touching code.
