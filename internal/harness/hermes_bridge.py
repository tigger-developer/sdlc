# ABOUTME: Exposes Hermes native response activity without publishing partial prose.
# ABOUTME: Keeps provider credentials, session persistence and tool isolation native.
import argparse
import contextlib
import io
import json
import os
import re
import sys
import threading
import time


class DiscardOutput(io.TextIOBase):
    def write(self, text: str) -> int:
        return len(text)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("command", choices=["chat"])
    parser.add_argument("--quiet", action="store_true", required=True)
    parser.add_argument("--query-file", choices=["-"], required=True)
    parser.add_argument("-m", required=True)
    parser.add_argument("--provider", required=True)
    parser.add_argument("-t", choices=["none"], required=True)
    parser.add_argument("--safe-mode", action="store_true", required=True)
    parser.add_argument("--in", dest="directory", required=True)
    parser.add_argument("--resume")
    parser.add_argument("--check", action="store_true")
    return parser.parse_args()


def configure_audit_agent(agent) -> None:
    if agent.tools:
        raise RuntimeError("Hermes audit isolation rejected enabled native tools")
    agent.quiet_mode = True
    agent.suppress_status_output = True
    agent.tool_progress_mode = "off"
    agent.skip_background_review = True
    agent.tool_gen_callback = None
    agent.thinking_callback = None  # wait notices are liveness, not responses
    agent.tool_progress_callback = None
    agent.tool_start_callback = None
    agent.tool_complete_callback = None
    agent.notice_callback = None
    agent.notice_clear_callback = None
    activity_lock = threading.Lock()
    last_activity = 0.0

    def activity(kind: str, text: str) -> None:
        nonlocal last_activity
        if not text:
            return
        with activity_lock:
            now = time.monotonic()
            if now - last_activity >= 0.25:
                print(f"sdlc_response: {kind}", file=sys.stderr, flush=True)
                last_activity = now

    agent.stream_delta_callback = lambda text: activity("content", text)
    agent.reasoning_callback = lambda text: activity("reasoning", text)


def reject_provider_result(result) -> None:
    # Publish only an allowlisted classification, never error bodies, messages,
    # reasoning, credentials or URLs. Older Hermes results omit classification.
    if not isinstance(result, dict):
        result = {}
    reasons = {
        "auth",
        "auth_permanent",
        "billing",
        "rate_limit",
        "upstream_rate_limit",
        "overloaded",
        "server_error",
        "timeout",
        "ssl_cert_verification",
        "context_overflow",
        "payload_too_large",
        "model_not_found",
        "provider_policy_blocked",
        "content_policy_blocked",
        "format_error",
    }
    reason = result.get("failure_reason")
    if not isinstance(reason, str) or reason not in reasons:
        reason = "unknown"
    retryable = result.get("failure_retryable")
    if not isinstance(retryable, bool):
        retryable = None
    error = str(result.get("error") or "")
    match = re.match(r"^HTTP ([45][0-9]{2}):\s*(.*)", error, re.DOTALL)
    status = int(match[1]) if match else None
    if reason == "unknown" and match:
        if status == 404 and re.match(r"^model:\s*\S+", match[2]):
            reason, retryable = "model_not_found", False
        elif status == 403 and re.match(
            r"^You've reached your [0-9]+-hour usage limit\b", match[2]
        ):
            reason, retryable = "quota_exhausted", False
        elif status == 400 and re.match(
            r"^prompt is too long:", match[2], re.IGNORECASE
        ):
            reason, retryable = "context_overflow", False
    print(
        "sdlc_failure: "
        + json.dumps(
            {
                "version": 1,
                "reason": reason,
                "status": status,
                "retryable": retryable,
            }
        ),
        file=sys.stderr,
        flush=True,
    )
    raise RuntimeError(
        "Hermes native provider call failed; see structured audit diagnostic"
    )


def run(args: argparse.Namespace, prompt: str) -> str:
    if (
        not args.check
        and args.provider.lower() in {"claude", "anthropic"}
        and args.m.lower() in {"sonnet", "opus", "haiku"}
    ):
        print(
            'sdlc_failure: {"version": 1, "reason": "model_not_found", "retryable": false}',
            file=sys.stderr,
            flush=True,
        )
        raise RuntimeError(
            "Hermes audit requires a full Anthropic model ID; Claude Code short aliases are not expanded by this adapter"
        )
    # Apply the native CLI's safe-mode boundary before importing Hermes modules.
    os.environ["HERMES_SAFE_MODE"] = "1"
    os.environ["HERMES_IGNORE_USER_CONFIG"] = "1"
    os.environ["HERMES_IGNORE_RULES"] = "1"
    os.environ["HERMES_SINGLE_QUERY_SESSION"] = "1"
    os.environ.pop("HERMES_EPHEMERAL_SYSTEM_PROMPT", None)
    os.environ["PYTHONDONTWRITEBYTECODE"] = "1"
    sys.dont_write_bytecode = True
    os.chdir(args.directory)
    from cli import HermesCLI, _flush_one_shot_session_store

    if args.check:
        print("Hermes native adapter interface available", file=sys.stderr)
        return ""
    cli = HermesCLI(
        model=args.m,
        provider=args.provider,
        toolsets=["none"],
        resume=args.resume,
        ignore_rules=True,
    )
    if cli._session_db is None:
        raise RuntimeError("Hermes session persistence unavailable")
    claimed = False
    try:
        claimed = cli._claim_active_session("cli", stderr=True)
        if not claimed:
            raise RuntimeError("Hermes native session is already active")
        if not cli._ensure_runtime_credentials():
            raise RuntimeError("Authentication failed: Hermes credentials unavailable")
        route = cli._resolve_turn_agent_config(prompt)
        if not cli._init_agent(
            model_override=route["model"],
            runtime_override=route["runtime"],
            request_overrides=route.get("request_overrides"),
        ):
            raise RuntimeError("Hermes native agent initialization failed")
        agent = cli.agent
        configure_audit_agent(agent)
        print(f"session_id: {cli.session_id}", file=sys.stderr, flush=True)
        result = agent.run_conversation(
            user_message=prompt,
            conversation_history=cli.conversation_history,
        )
        if (
            not isinstance(result, dict)
            or result.get("failed")
            or result.get("partial")
            or result.get("error")
        ):
            reject_provider_result(result)
        if agent.session_id != cli.session_id:
            return "CONTEXT-LOST"
        response = result.get("final_response", "")
        if not response:
            raise RuntimeError("Hermes omitted final response")
        return response
    finally:
        try:
            _flush_one_shot_session_store(cli)
        finally:
            try:
                if claimed:
                    cli._release_active_session()
            finally:
                cli._session_db.close()


def main() -> int:
    args = parse_args()
    prompt = sys.stdin.read(8 * 1024 * 1024 + 1)
    if len(prompt.encode("utf-8")) > 8 * 1024 * 1024:
        print("Hermes audit prompt exceeds 8 MiB", file=sys.stderr)
        return 2
    try:
        # Native CLI display output must never become audit evidence or activity.
        with contextlib.redirect_stdout(DiscardOutput()):
            response = run(args, prompt)
    except (ImportError, AttributeError, RuntimeError, ValueError, OSError) as error:
        print(f"Hermes native adapter error: {error}", file=sys.stderr)
        return 1
    if response:
        print(response, flush=True)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
