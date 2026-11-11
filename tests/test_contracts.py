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
    ("mission/v1/mission-operation.schema.json",
     ["mission-operation.json"],
     ["mission-operation-missing-assignment.json",
      "mission-operation-zero-fence.json",
      "mission-operation-missing-delegated-operation.json"]),
    ("effect/v1/effect-report.schema.json",
     ["effect-report-reported.json", "effect-report-unknown.json"],
     ["effect-report-missing-fence.json",
      "effect-report-bad-status.json",
      "effect-report-completion-claim.json"]),
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

    def test_mission_api_paths(self):
        spec = json.loads((CONTRACTS / "mission/v1/mission-api.openapi.json").read_text(encoding="utf-8"))
        self.assertEqual("3.1.0", spec["openapi"])
        self.assertEqual("v1", spec["info"]["x-contract-version"])
        self.assertEqual("reviewed", spec["info"]["x-contract-status"])
        self.assertEqual("review.md", spec["info"]["x-review-ref"])
        self.assertEqual({
            "/missions",
            "/missions/{mission_id}",
            "/missions/{mission_id}/assignments",
            "/missions/{mission_id}/operations",
            "/missions/{mission_id}/cancel",
            "/missions/{mission_id}/effect-reports",
        }, set(spec["paths"]))
        self.assertEqual({"get", "post"}, set(spec["paths"]["/missions"]))
        self.assertEqual({"get"}, set(spec["paths"]["/missions/{mission_id}"]))
        self.assertEqual({"post"}, set(spec["paths"]["/missions/{mission_id}/assignments"]))
        self.assertEqual({"post"}, set(spec["paths"]["/missions/{mission_id}/operations"]))
        self.assertEqual({"post"}, set(spec["paths"]["/missions/{mission_id}/cancel"]))
        self.assertEqual({"post"}, set(spec["paths"]["/missions/{mission_id}/effect-reports"]))

    def test_mission_api_refs_resolve(self):
        spec = json.loads((CONTRACTS / "mission/v1/mission-api.openapi.json").read_text(encoding="utf-8"))
        def walk(value):
            if isinstance(value, dict):
                ref = value.get("$ref")
                if isinstance(ref, str):
                    self.assertTrue(ref.startswith("#/"), ref)
                    target = spec
                    for part in ref[2:].split("/"):
                        target = target[part]
                for child in value.values():
                    walk(child)
            elif isinstance(value, list):
                for child in value:
                    walk(child)
        walk(spec)

    def test_mission_api_components_match_schemas(self):
        spec = json.loads((CONTRACTS / "mission/v1/mission-api.openapi.json").read_text(encoding="utf-8"))
        pairs = {
            "Mission": "mission/v1/mission.schema.json",
            "MissionAssignment": "mission/v1/mission-assignment.schema.json",
            "MissionOperation": "mission/v1/mission-operation.schema.json",
            "EffectReport": "effect/v1/effect-report.schema.json",
        }
        for component, rel in pairs.items():
            with self.subTest(component=component):
                a = set(spec["components"]["schemas"][component]["required"])
                b = set(load(rel)["required"])
                self.assertEqual(b, a)
                self.assertEqual(load(rel)["properties"].keys(), spec["components"]["schemas"][component]["properties"].keys())

    def test_mission_api_fixtures(self):
        spec = json.loads((CONTRACTS / "mission/v1/mission-api.openapi.json").read_text(encoding="utf-8"))
        def v(name):
            root = dict(spec)
            root["$schema"] = "https://json-schema.org/draft/2020-12/schema"
            root["$ref"] = f"#/components/schemas/{name}"
            return Draft202012Validator(root, format_checker=FormatChecker())
        cases = [
            ("MissionView", "valid/mission-view.json", True),
            ("AssignmentRequest", "valid/assignment-request.json", True),
            ("EffectReportResult", "valid/effect-report-result.json", True),
            ("MissionView", "invalid/mission-view-missing-cancel.json", False),
            ("AssignmentRequest", "invalid/assignment-request-zero-fence.json", False),
        ]
        for component, fixture, accepted in cases:
            with self.subTest(fixture=fixture):
                errors = list(v(component).iter_errors(load(f"fixtures/v1/{fixture}")))
                self.assertEqual(accepted, not errors, errors)
