import hashlib
import json
import unittest
from pathlib import Path

from jsonschema import Draft202012Validator, FormatChecker

ROOT = Path(__file__).resolve().parents[1]
CONTRACTS = ROOT / "contracts"

CANONICAL = [
    ("authority/v1/offline-lease.schema.json", ["offline-lease.json"], ["offline-lease-zero-missions.json"]),
    ("assurance/v1/release-assurance-profile.schema.json", ["release-assurance-profile.json"], ["release-assurance-profile-empty-runs.json"]),
    ("capability/v1/embodied-capability-binding.schema.json", ["embodied-capability-binding.json"], []),
    ("compatibility/v1/compatibility-report.schema.json", ["compatibility-report.json"], []),
    ("evidence/v1/effect-record.schema.json", ["effect-record.json"], []),
    ("fleet/v1/fleet-spec.schema.json", ["fleet-spec.json"], []),
    ("mission/v1/mission.schema.json", ["mission.json"], ["mission-missing-authority.json"]),
    ("mission/v1/mission-assignment.schema.json", ["mission-assignment.json"], ["mission-assignment-zero-fence.json"]),
    ("preflight/v1/mission-preflight-profile.schema.json", ["mission-preflight-profile.json"], ["mission-preflight-profile-missing-plan.json"]),
    ("replan/v1/replan-request.schema.json", ["replan-request.json"], ["replan-request-missing-reason.json"]),
    ("robot/v1/robot-descriptor.schema.json", ["robot-descriptor.json"], []),
    ("safety/v1/safety-envelope.schema.json", ["safety-envelope.json"], ["safety-envelope-invalid-fallback.json"]),
    ("state/v1/robot-state-reference.schema.json", ["robot-state-reference.json"], []),
]

def load(rel):
    return json.loads((CONTRACTS / rel).read_text(encoding="utf-8"))

def validator(rel):
    schema = load(rel)
    Draft202012Validator.check_schema(schema)
    return Draft202012Validator(schema, format_checker=FormatChecker())

class ContractTests(unittest.TestCase):
    def test_canonical_schema_fixtures(self):
        for rel, valid, invalid in CANONICAL:
            with self.subTest(schema=rel):
                v = validator(rel)
                for name in valid:
                    with self.subTest(fixture=name):
                        self.assertEqual([], list(v.iter_errors(load(f"fixtures/v1/valid/{name}"))))
                for name in invalid:
                    with self.subTest(fixture=name):
                        self.assertNotEqual([], list(v.iter_errors(load(f"fixtures/v1/invalid/{name}"))))

    def test_all_schemas_compile(self):
        for path in CONTRACTS.rglob("*.schema.json"):
            with self.subTest(schema=str(path.relative_to(ROOT))):
                Draft202012Validator.check_schema(json.loads(path.read_text(encoding="utf-8")))

    def test_schema_ids_are_unique(self):
        ids = {}
        for path in CONTRACTS.rglob("*.schema.json"):
            schema = json.loads(path.read_text(encoding="utf-8"))
            sid = schema.get("$id")
            if sid:
                with self.subTest(schema=str(path.relative_to(ROOT))):
                    self.assertNotIn(sid, ids)
                    ids[sid] = path

    def test_no_remote_schema_refs(self):
        def walk(value):
            if isinstance(value, dict):
                ref = value.get("$ref")
                if isinstance(ref, str):
                    self.assertTrue(ref.startswith("#/"), ref)
                for child in value.values():
                    walk(child)
            elif isinstance(value, list):
                for child in value:
                    walk(child)
        for path in CONTRACTS.rglob("*.schema.json"):
            if "events" in path.parts:
                continue
            with self.subTest(schema=str(path.relative_to(ROOT))):
                walk(json.loads(path.read_text(encoding="utf-8")))

    def test_sync_manifest_matches(self):
        manifest = json.loads((CONTRACTS / "sync-manifest.json").read_text(encoding="utf-8"))
        self.assertEqual("KINETOVELA/contracts", manifest["source"])
        for entry in manifest["files"]:
            digest = hashlib.sha256((CONTRACTS / entry["path"]).read_bytes()).hexdigest()
            with self.subTest(path=entry["path"]):
                self.assertEqual(entry["sha256"], digest)
