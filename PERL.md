# Perl Standards

Perl-specific standards. The general coding standards in `[sdlc-home]/CODING.md` apply on top of these.

Perl is a maintenance language here, not a default choice for new tooling. Use Perl when the project is already Perl, when a mature CPAN ecosystem is the clear reason for the choice, or when maintaining existing Perl is lower-risk than a rewrite. Do not reach for Perl one-liners as a loophole around the shell rules that prohibit `sed`, `awk`, and `perl` for stream editing.

Primary references:

- [perlstyle](https://perldoc.perl.org/perlstyle)
- [perlsec](https://perldoc.perl.org/perlsec)
- [perltidy](https://perltidy.sourceforge.net/)
- [Perl::Critic](https://metacpan.org/pod/Perl::Critic)

## Tooling

Use the project's existing configuration first.

| Tool | Purpose | Required |
|---|---|---|
| `perltidy` | Formatting | Yes |
| `perlcritic` / Perl::Critic | Linting | Yes, when the project has or needs lint configuration |
| `prove` + `Test::More` | Tests | Yes, unless the project already uses another standard Perl test harness |

Run formatting and linting on changed Perl files before review. Do not introduce a new lint policy set into an existing project without explicit reason; Perl::Critic can be noisy without project tuning.

## Required Pragmas

Every Perl file must enable strictness and warnings near the top:

```perl
use strict;
use warnings;
```

For new code, also prefer an explicit feature bundle and UTF-8 handling where appropriate:

```perl
use v5.32;
use utf8;
```

Do not disable `strict` or `warnings` globally. A local suppression must be narrow, justified in a nearby comment, and removed when no longer needed.

## File and IO Handling

- Use lexical filehandles, not bareword filehandles.
- Use three-argument `open`.
- Always check IO failures.
- Specify encoding when reading or writing text.
- Use `File::Temp` for temporary files.

```perl
open my $fh, '<:encoding(UTF-8)', $path
    or die "opening $path: $!";
```

Never use two-argument `open`; it can interpret user-controlled strings as shell commands.

## Error Handling

- Fail loudly with useful context.
- Preserve `$!` and `$@` context in messages where relevant.
- Do not swallow exceptions with empty `eval` blocks.
- Prefer `autodie` in small scripts where it improves clarity and the project already accepts it.
- For library code, use structured return values or exceptions consistently; do not mix styles arbitrarily.

## Anti-Patterns

These are prohibited unless an existing project requires them and the exception is documented:

- Perl one-liners used as shell-stream editors.
- Bareword filehandles.
- Two-argument `open`.
- String `eval`.
- Backticks or `qx//` with interpolated data.
- Global mutable state for ordinary data flow.
- Empty `eval` blocks or ignored `$@`.
- Disabling `strict`, `warnings`, or Perl::Critic policies without a specific local justification.
- Regex parsing for nested structured formats.
- Obfuscated Perl idioms that require expert Perl knowledge to maintain.

# Canary
Suffix the canary string with "PERL "
