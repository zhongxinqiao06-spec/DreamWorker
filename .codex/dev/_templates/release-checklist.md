# Release Checklist

## Release Metadata

- Version:
- Date:
- Owner:
- Build type: dev / preview / mvp / hotfix

## Gates

- [ ] TypeScript typecheck passed.
- [ ] Go tests passed.
- [ ] Go vet passed.
- [ ] Contract tests passed.
- [ ] Renderer tests passed.
- [ ] E2E smoke passed.
- [ ] Security smoke passed.
- [ ] Golden tasks passed.
- [ ] SLO smoke passed.
- [ ] Diagnostics export redaction passed.

## Data Safety

- [ ] Migration backup tested.
- [ ] Migration restore tested.
- [ ] EventStore replay tested.
- [ ] Artifact paths verified.

## Security

- [ ] Renderer boundary smoke passed.
- [ ] No secret in renderer events.
- [ ] Revoked capability cannot run.
- [ ] High-risk action requires approval.
- [ ] Markdown sanitizer passed.

## Packaging

- [ ] Go Engine bundled.
- [ ] First-run onboarding works.
- [ ] Safe mode works.
- [ ] Engine startup failure fallback works.
- [ ] Config reset works.

## Rollback

- Rollback version:
- Restore steps:
- Known incompatible changes:
