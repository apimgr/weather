# AI Implementation Plan

This file is the repository-local implementation plan. Per `AI.md`, if `PLAN.AI.md` exists, this is the active plan.

## Current Goal

Finish the remaining verified `audit-graphql-runtime` work without drifting from `AI.md`, with PART 34 multi-user behavior treated as fully non-negotiable.

## Current Verified State

1. The live multi-user REST and web runtime has been substantially aligned with `AI.md`, including auth/session fixes, invite flows, current-user security surfaces, and the mounted passkey/WebAuthn runtime.
2. The remaining verified auth-related gap is no longer the REST/web passkey runtime; it is the missing GraphQL passkey/security parity now that `/auth/passkey`, `/api/v1/auth/passkey/*`, and `/api/v1/users/security/passkeys*` are mounted.
3. `TODO.AI.md` now mirrors the active SQL task ledger and the verified progress notes so the repo contains the handoff state needed for migration to a new development server.

## Active Plan

1. Re-read the relevant `AI.md` sections and `.claude/rules/*.md` files before each new implementation pass; never guess or infer behavior that is specified.
2. Continue `audit-graphql-runtime` by tracing the mounted passkey/security runtime and identifying the exact GraphQL contract required to mirror the live behavior.
3. Implement GraphQL passkey/security parity runtime-first:
   - reuse shared REST/runtime helpers instead of duplicating business logic
   - only expose surfaces that are actually implemented in the mounted runtime
   - hand-sync checked-in GraphQL artifacts if generation remains unavailable
4. Keep `TODO.AI.md`, `PLAN.AI.md`, and `.git/COMMIT_MESS` current as verified work changes.
5. Keep using the SQL task ledger during active work, but mirror meaningful task state back into `TODO.AI.md` so the repository remains the authoritative migration handoff.

## Immediate Next Verified Task

Audit the exact GraphQL passkey/security parity gap against the now-mounted runtime and implement the next verified missing surface without assuming schema or response shape.
