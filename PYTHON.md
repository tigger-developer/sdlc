# Python Standards

Python-specific standards. The general coding standards in `[sdlc-home]/CODING.md` apply on top of these.

**Python is not the default.** Before reaching for Python, confirm it is the genuinely best fit -- see CODING.md "Language and Tool Selection". The Python-bias in LLM training data is strong; resist it. Python earns its keep in ML and scientific computing; for general-purpose code, Go or Rust is typically the better starting point.

**Never use system Python on macOS.**
**Never invoke python outside of a venv.**

## Version Management

Target Python 3.12.x (current LTS) or 3.13.x (current stable). Pin per-project with a `.python-version` file containing just the version string (e.g. `3.12`).

**Preferred:** `uv` -- manages Python versions, venvs, and packages in one tool. Replaces `pyenv` + `pip` + `venv`.

**Acceptable:** `pyenv` + `python3 -m venv` -- use this if the project already uses it or `uv` is not available.

Never use `pyenv local` to set a version mid-script; use `.python-version` instead.

## Virtual Environments

**Never `pip install` or `uv pip install` outside a venv.** No exceptions.

**With uv:**

```bash
uv venv          # creates .venv using version from .python-version
uv pip install -r requirements.txt
```

**With pip:**

```bash
python3 -m venv .venv
.venv/bin/pip install --upgrade pip -q
.venv/bin/pip install -r requirements.txt -q
```

## Project Structure

Use a `src/` layout for new projects -- package lives at `src/<packagename>/`. This prevents accidental imports of source rather than the installed package. Do not restructure existing projects to `src/` layout simply because you are touching them.

## Dependency Files

Prefer `pyproject.toml` for packages with a proper structure. `requirements.txt` is acceptable for simple scripts and tools.

```toml
[project]
name = "mypackage"
requires-python = ">=3.12"
dependencies = ["pyyaml>=6.0"]

[project.optional-dependencies]
dev = ["pytest>=8.0", "ruff>=0.4"]
```

## Ruff Configuration

Standard config for all Python projects:

```toml
[tool.ruff]
target-version = "py312"
line-length = 88

[tool.ruff.lint]
select = ["E", "W", "F", "I", "B", "UP"]
ignore = ["E501"]  # line too long -- enforced by formatter, not linter

[tool.ruff.lint.isort]
known-first-party = ["<package-name>"]
```

`E501` is always ignored -- `ruff format` enforces line length; having both causes conflicts. 88 is the community standard (Black's default). Deviate only with documented reason.

# Canary
Suffix the canary string with "PY "
