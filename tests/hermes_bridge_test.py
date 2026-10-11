# ABOUTME: Verifies Hermes callback, isolation and session contracts with native-interface doubles.
# ABOUTME: Runs entirely locally without credentials, hosted models or private runtime state.
import argparse
import contextlib
import importlib.util
import io
import json
import os
import pathlib
import sys
import tempfile
import types
import unittest
from unittest.mock import patch

ROOT = pathlib.Path(__file__).resolve().parents[1]
SPEC = importlib.util.spec_from_file_location(
    "hermes_bridge", ROOT / "internal/harness/hermes_bridge.py"
)
BRIDGE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(BRIDGE)


class NativeInterfaceTest(unittest.TestCase):
    def test_reasoning_starts_response_without_leaking_prose(self):
        response, metadata = self.invoke()
        self.assertEqual(response, "VERDICT: PASS")
        self.assertIn("session_id: native-fixture", metadata)
        self.assertIn("sdlc_response: reasoning", metadata)
        self.assertNotIn("private reasoning", metadata)
        self.assertNotIn("partial prose", metadata)

    def test_enabled_tools_are_rejected_before_conversation(self):
        with self.assertRaisesRegex(RuntimeError, "enabled native tools"):
            self.invoke(tools=[{"function": {"name": "read_file"}}])

    def test_compression_does_not_pass_with_lost_context(self):
        response, _ = self.invoke(compress=True)
        self.assertEqual(response, "CONTEXT-LOST")

    def test_failed_result_cannot_supply_a_verdict(self):
        with self.assertRaisesRegex(RuntimeError, "provider call failed"):
            self.invoke(failed=True)

    def test_context_compression_cannot_hide_a_provider_failure(self):
        with self.assertRaisesRegex(RuntimeError, "provider call failed"):
            self.invoke(failed=True, compress=True)

    def test_resume_preserves_the_native_session_and_history(self):
        response, _ = self.invoke(resume="native-fixture")
        self.assertEqual(response, "VERDICT: PASS")

    def test_authentication_failure_releases_session(self):
        with self.assertRaisesRegex(RuntimeError, "Authentication failed"):
            self.invoke(authenticated=False)

    def test_provider_failures_publish_safe_structured_diagnostics(self):
        cases = [
            ({"error": "HTTP 404: model: sonnet"}, "model_not_found", 404, False),
            (
                {"error": "HTTP 403: You've reached your 5-hour usage limit."},
                "quota_exhausted",
                403,
                False,
            ),
            (
                {"failure_reason": "billing", "failure_retryable": False},
                "billing",
                None,
                False,
            ),
            (
                {"failure_reason": "server_error", "failure_retryable": True},
                "server_error",
                None,
                True,
            ),
            ({"error": "secret-key private evidence"}, "unknown", None, None),
            ({"error": "HTTP 404: endpoint not found"}, "unknown", 404, None),
            (
                {
                    "error": "HTTP 400: prompt is too long: 2060374 tokens > 1000000 maximum"
                },
                "context_overflow",
                400,
                False,
            ),
            (
                {"failure_reason": "secret-key", "failure_retryable": 0},
                "unknown",
                None,
                None,
            ),
        ]
        for result, reason, status, retryable in cases:
            with self.subTest(reason=reason):
                metadata = io.StringIO()
                with contextlib.redirect_stderr(metadata):
                    with self.assertRaisesRegex(RuntimeError, "provider call failed"):
                        BRIDGE.reject_provider_result(result)
                envelope = json.loads(
                    metadata.getvalue().removeprefix("sdlc_failure: ")
                )
                self.assertEqual(
                    envelope,
                    {
                        "version": 1,
                        "reason": reason,
                        "status": status,
                        "retryable": retryable,
                    },
                )
                self.assertNotIn("secret-key", metadata.getvalue())

    def test_anthropic_short_alias_rejected_before_native_initialization(self):
        args = argparse.Namespace(provider="claude", m="sonnet", check=False)
        metadata = io.StringIO()
        with contextlib.redirect_stderr(metadata):
            with self.assertRaisesRegex(RuntimeError, "full Anthropic model ID"):
                BRIDGE.run(args, "verified evidence")
        self.assertIn('"reason": "model_not_found"', metadata.getvalue())

    def invoke(
        self, tools=None, compress=False, failed=False, resume=None, authenticated=True
    ):
        test = self
        lifecycle = []

        class DB:
            def close(self):
                lifecycle.append("closed")

        class Agent:
            session_id = "native-fixture"

            def run_conversation(self, user_message, conversation_history):
                test.assertEqual(user_message, "verified evidence")
                test.assertEqual(conversation_history, ["retained"] if resume else [])
                test.assertTrue(self.skip_background_review)
                test.assertNotIn("HERMES_EPHEMERAL_SYSTEM_PROMPT", os.environ)
                self.reasoning_callback("private reasoning")
                self.stream_delta_callback("partial prose")
                if compress:
                    self.session_id = "compressed"
                return {"final_response": "VERDICT: PASS", "failed": failed}

        class CLI:
            session_id = "native-fixture"

            def __init__(self, **kwargs):
                test.assertEqual(os.environ["HERMES_SAFE_MODE"], "1")
                test.assertEqual(kwargs["toolsets"], ["none"])
                test.assertTrue(kwargs["ignore_rules"])
                test.assertEqual(kwargs["resume"], resume)
                self.agent = Agent()
                self.agent.tools = tools or []
                self._session_db = DB()
                self.conversation_history = ["retained"] if resume else []

            def _ensure_runtime_credentials(self):
                return authenticated

            def _claim_active_session(self, surface, stderr):
                return True

            def _release_active_session(self):
                lifecycle.append("released")

            def _resolve_turn_agent_config(self, prompt):
                return {"model": "fixture", "runtime": {}}

            def _init_agent(self, **kwargs):
                return True

        native = types.ModuleType("cli")
        native.HermesCLI = CLI
        native._flush_one_shot_session_store = lambda cli: None
        previous = os.getcwd()
        try:
            with tempfile.TemporaryDirectory() as directory:
                args = argparse.Namespace(
                    directory=directory,
                    m="fixture",
                    provider="fixture",
                    resume=resume,
                    check=False,
                )
                metadata = io.StringIO()
                with (
                    patch.dict(sys.modules, {"cli": native}),
                    patch.dict(
                        os.environ,
                        {"HERMES_EPHEMERAL_SYSTEM_PROMPT": "ambient persona"},
                    ),
                    contextlib.redirect_stderr(metadata),
                ):
                    result = BRIDGE.run(args, "verified evidence")
                return result, metadata.getvalue()
        finally:
            os.chdir(previous)
            self.assertEqual(lifecycle, ["released", "closed"])


if __name__ == "__main__":
    unittest.main()
