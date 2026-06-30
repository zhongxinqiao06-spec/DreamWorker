# PR Template

## Summary

- PR ID:
- Phase:
- Owner:
- Priority:
- Related plan:
- Related risk:

## Scope

- In:
- Out:

## Contract / Schema Impact

- Schema changed: yes/no
- Event changed: yes/no
- API changed: yes/no
- Migration required: yes/no

## Tests

- Unit:
- Integration:
- Contract:
- Renderer:
- E2E:
- Security smoke:

## Verification

Commands or manual smoke steps:

```powershell
# fill in
```

## Rollback

- Feature flag:
- Migration rollback:
- Data backup:
- Safe disable path:

## Checklist

- [ ] PR maps to one dev phase.
- [ ] PR has an independent verification path.
- [ ] High-risk actions go through PolicyEngine.
- [ ] State changes write EventStore.
- [ ] Renderer does not access Node, Go, SQLite, filesystem or secrets.
- [ ] trace_id is propagated where applicable.
- [ ] Docs and fixtures are updated.
