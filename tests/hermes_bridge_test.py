# ABOUTME: Verifies Hermes callback, isolation and session contracts with native-interface doubles.
# ABOUTME: Runs entirely locally without credentials, hosted models or private runtime state.
import argparse
import contextlib
import importlib.util
import io
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

    def test_resume_preserves_the_native_session_and_history(self):
        response, _ = self.invoke(resume="native-fixture")
        self.assertEqual(response, "VERDICT: PASS")

    def test_authentication_failure_releases_session(self):
        with self.assertRaisesRegex(RuntimeError, "Authentication failed"):
            self.invoke(authenticated=False)

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
