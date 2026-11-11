# Contract Tests

Run the contract suite with:

```
python -m unittest discover -s tests -v
```

Requirements: Python 3 with `jsonschema` 4.x installed.

This suite validates contracts only, not runtime behavior. It compiles every schema, checks canonical schemas against their fixtures, enforces unique schema IDs, rejects remote `$ref` values outside the pre-existing `contracts/events` set, and verifies the sync manifest digests.
