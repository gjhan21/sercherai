# Newclient Formal Cutover Design

## Background

The repository currently has two frontend directories:

- `/Users/gjhan21/cursor/sercherai/client`
- `/Users/gjhan21/cursor/sercherai/newclient`

The user has confirmed that `newclient` is now the real frontend and the old `client` should no longer participate in any official runtime, build, deployment, or onboarding flow.

Today the codebase is in an inconsistent middle state:

- local development already starts `newclient` from `scripts/devctl.sh`
- deployment scripts and Linux deployment docs still describe the official frontend as `client`
- deploy output paths and nginx placeholders still publish to `client`
- many historical documents still reference the old `client` as if it were the current frontend

This spec defines a full naming and deployment cutover so the repository has a single official frontend identity: `newclient`.

## Goal

Make `newclient` the only official frontend across development, build, deployment, and active documentation, while keeping the old `client` directory in the repository strictly as historical reference material.

## Non-Goals

This work does not:

- delete `/Users/gjhan21/cursor/sercherai/client`
- rewrite historical implementation documents as if they originally targeted `newclient`
- redesign frontend product behavior or page structure
- merge old `client` source code into `newclient`
- change backend, admin, or strategy-engine responsibilities beyond naming and integration points required by the cutover

## User Decisions Locked In

The user explicitly chose the following:

- use the current `newclient` frontend as the official frontend
- perform a full naming cutover rather than a compatibility alias migration
- stop using `client` as the formal development command and environment name
- keep the old `client` directory in the repo as reference instead of deleting it
- include historical document cleanup as part of the cutover

## Source Of Truth After Cutover

After this migration, the official frontend source of truth is:

- source directory: `/Users/gjhan21/cursor/sercherai/newclient`
- dev service name: `newclient`
- dev env file: `/Users/gjhan21/cursor/sercherai/.run/newclient.env`
- dev variables: `NEWCLIENT_HOST`, `NEWCLIENT_PORT`
- deployed static root: `${WWW_DIR}/newclient`

The old `/Users/gjhan21/cursor/sercherai/client` directory remains readable but is no longer part of any supported operational workflow.

## Design Overview

The cutover is split into three layers that are delivered together but treated differently:

1. Official chain hard cut
2. Active documentation unification
3. Historical documentation de-noising

This avoids the two failure modes that would otherwise happen:

- keeping the old `client` name alive in production-facing paths, which would preserve confusion indefinitely
- rewriting historical documents so aggressively that they stop being historically accurate

## Layer 1: Official Chain Hard Cut

### Intent

Every supported operational path must refer to `newclient` and only `newclient`.

### Dev Control Changes

`/Users/gjhan21/cursor/sercherai/scripts/devctl.sh` should be updated so that:

- the service name `client` is replaced by `newclient`
- usage text, supported targets, status output, env lookup, port lookup, and service URLs all use `newclient`
- the environment file changes from `.run/client.env` to `.run/newclient.env`
- the variables change from `CLIENT_HOST` and `CLIENT_PORT` to `NEWCLIENT_HOST` and `NEWCLIENT_PORT`
- the underlying command continues to launch `/Users/gjhan21/cursor/sercherai/newclient`

No compatibility alias for `client` is kept. If someone still uses `client`, that should be treated as outdated usage rather than silently accepted.

### Build Changes

Deployment must stop building `/Users/gjhan21/cursor/sercherai/client`.

The deployment flow should instead build:

- PC bundle from `newclient` with `build:pc`
- H5 bundle from `newclient` with `build:h5`

If needed, `newclient/package.json` may expose a single aggregate `build` script that runs both, so deployment code can stay simple and explicit.

### Static Publish Changes

`/Users/gjhan21/cursor/sercherai/scripts/deploy_linux_server.sh` should publish the frontend to `${WWW_DIR}/newclient`, not `${WWW_DIR}/client`.

The publish logic must support one static root serving both:

- PC entry on `/`
- H5 entry on `/m/`

The implementation may either:

- publish PC and H5 outputs directly into a single assembled `${WWW_DIR}/newclient` tree
- or build them separately and merge them into the final runtime tree during deployment

The important invariant is that the deployed directory structure matches the router/base expectations of `newclient`.

### Nginx Changes

`/Users/gjhan21/cursor/sercherai/deploy/linux/sercherai.nginx.conf.template` and the rendering logic in `/Users/gjhan21/cursor/sercherai/scripts/deploy_linux_server.sh` should stop using `CLIENT_ROOT` and `CLIENT_PORT`.

They should instead use `NEWCLIENT_ROOT` and `NEWCLIENT_PORT`, with the runtime root set to `${WWW_DIR}/newclient`.

This is not just cosmetic. The configuration must make it obvious to future operators that the deployed frontend is `newclient`, not a hidden reuse of the old `client` label.

## Layer 2: Active Documentation Unification

### Intent

Any document that a teammate is likely to follow today must describe only the `newclient` path.

### In Scope

The priority active documents are:

- `/Users/gjhan21/cursor/sercherai/scripts/README.md`
- `/Users/gjhan21/cursor/sercherai/docs/DEPLOY_LINUX.md`

These docs should be rewritten so that:

- all official commands use `newclient`
- all environment file examples use `.run/newclient.env`
- all port variable examples use `NEWCLIENT_*`
- deployment instructions say the frontend being built and published is `newclient`
- references like "backend/admin/client" become "backend/admin/newclient"

After this change, a new engineer following the active docs should never be instructed to use the old `client` name for the current frontend.

## Layer 3: Historical Documentation De-Noising

### Intent

Historical docs may still talk about the old `client`, but they must stop misleading readers into thinking it is still the official frontend.

### Principle

Historical accuracy must be preserved.

That means the migration should not mass-rewrite old specs, plans, and retrospective docs to falsely claim they were always about `newclient`.

Instead, old documents that are still searchable and likely to be opened should receive a concise note that:

- the document describes work from the old `client` era
- the current official frontend has moved to `/Users/gjhan21/cursor/sercherai/newclient`

### Recommended Treatment

For documents with substantial old-client references, add a short top-level note near the beginning rather than editing every historical path in the body.

This keeps the history readable and trustworthy while removing present-day ambiguity.

### Priority Areas

Focus historical de-noising on documents under:

- `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs`
- `/Users/gjhan21/cursor/sercherai/docs/superpowers/plans`
- `/Users/gjhan21/cursor/sercherai/docs/vibe-stock-growth`

The goal is not perfect elimination of every `client` string in the repo. The goal is to prevent human confusion in the places people actually read.

## Directory Policy For Old Client

`/Users/gjhan21/cursor/sercherai/client` remains in the repository, but its role changes to:

- historical reference
- fallback code-reading resource
- non-operational artifact

It should not appear in:

- active development commands
- deployment scripts
- nginx render variables
- active setup and deployment instructions
- any wording that implies it is the current frontend

## Validation Strategy

The cutover is complete only if all of the following are true:

- `scripts/devctl.sh` exposes `newclient` instead of `client`
- local env examples reference `.run/newclient.env`
- deployment scripts build `newclient`, not `client`
- deployment scripts publish to `${WWW_DIR}/newclient`
- nginx rendering uses `NEWCLIENT_*` naming
- active docs describe `newclient` as the official frontend
- old `client` remains in the repo but is absent from official operational flows
- historical docs that still mention old `client` make that status explicit

## Risks

### Risk 1: Broken PC/H5 deployment structure

`newclient` builds PC and H5 separately. The deployment logic must assemble those outputs into a static structure that still supports `/` and `/m/` correctly.

### Risk 2: Partial naming migration

If command names change but deployment docs or nginx placeholders do not, the repo will become more confusing, not less.

### Risk 3: Over-cleaning history

If historical docs are rewritten too aggressively, the repo loses reliable architectural history. The cleanup must annotate history, not falsify it.

## Testing Expectations

The implementation should verify at minimum:

- `newclient` PC build succeeds
- `newclient` H5 build succeeds
- deployment script text and generated paths no longer reference official `client` naming
- active docs no longer instruct users to operate through old `client` naming
- repository scans of operational files confirm the old official naming has been removed from live paths

## Recommended Implementation Direction

Implement the migration in this order:

1. rename official dev service and env naming to `newclient`
2. switch deployment build and publish logic to `newclient`
3. update nginx placeholder naming and final static root
4. update active docs
5. add historical notes to high-value old-client documents
6. verify builds and scan for remaining official-chain references

This order keeps the operational source of truth aligned as early as possible and reduces the chance of documenting a state that the scripts do not actually implement.
