// ABOUTME: Classifies shell commands submitted to the SDLC pre-tool guard.
// ABOUTME: Uses shell syntax so document data is never mistaken for a command.
package commandguard

import (
	"path"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

const maxCommandDepth = 8

// CheckShell returns the reason to block a shell command, or an empty string.
func CheckShell(command string) string {
	return checkShell(command, 0)
}

func checkShell(command string, depth int) string {
	if depth > maxCommandDepth {
		return "nested shell command exceeds the guard's inspection limit."
	}
	program, err := syntax.NewParser(syntax.Variant(syntax.LangBash)).Parse(strings.NewReader(command), "")
	if err != nil {
		return "shell command could not be parsed safely."
	}
	var reason string
	syntax.Walk(program, func(node syntax.Node) bool {
		if reason != "" {
			return false
		}
		switch typed := node.(type) {
		case *syntax.Stmt:
			reason = checkInputRedirections(typed.Redirs)
		case *syntax.CallExpr:
			reason = checkCall(typed.Args, depth)
		}
		return reason == ""
	})
	return reason
}

func checkInputRedirections(redirections []*syntax.Redirect) string {
	for _, redirect := range redirections {
		switch redirect.Op {
		case syntax.RdrIn, syntax.RdrInOut:
			if name, ok := staticWord(redirect.Word); ok && exactEnvFile(name) {
				return "reading a file whose exact basename is .env is prohibited."
			}
		}
	}
	return ""
}

func checkCall(words []*syntax.Word, depth int) string {
	if len(words) == 0 {
		return ""
	}
	for len(words) > 0 {
		executable, ok := staticWord(words[0])
		if !ok {
			return "dynamic command names cannot be classified safely."
		}
		name := path.Base(executable)
		switch name {
		case "env", "command", "sudo", "time":
			remaining, reason, inspect := unwrap(name, words[1:])
			if reason != "" || !inspect {
				return reason
			}
			words = remaining
			continue
		case "python", "python3":
			return "direct python and python3 interpreter commands are prohibited; use the task-appropriate tool or a project-owned entry point."
		case "rm":
			return "rm is prohibited; use recoverable deletion such as trash."
		case "sed", "awk":
			return "sed and awk are prohibited; use format-aware tools or an explicit patch."
		case "source", ".":
			if len(words) == 2 && operatorEnvrcWord(words[1]) {
				return ""
			}
			return "sourcing arbitrary files bypasses command policy; run an explicit command instead."
		case "eval":
			return "eval executes uninspected shell text; use an explicit command instead."
		case "bash", "sh", "zsh":
			return checkShellWrapper(words[1:], depth)
		case "chmod":
			if hasLiteral(words[1:], "777") {
				return "chmod 777 is prohibited; use the least permissive mode that works."
			}
		case "git":
			if reason := checkGit(words[1:]); reason != "" {
				return reason
			}
		case "gh":
			if len(words) > 2 && literalEquals(words[1], "repo") && (literalEquals(words[2], "create") || literalEquals(words[2], "edit")) {
				return "GitHub repository create/edit widens or changes access; ask the user first."
			}
		}
		if reason := checkEnvArguments(name, words[1:]); reason != "" {
			return reason
		}
		if name != "echo" && name != "printf" && name != "rg" && hasAnyLiteral(words[1:], "--no-hooks", "--no-pre-commit-hook") {
			return "hook bypass flags are prohibited; run the hooks or surface the failing hook."
		}
		return ""
	}
	return ""
}

func checkShellWrapper(words []*syntax.Word, depth int) string {
	for index, word := range words {
		option, ok := staticWord(word)
		if !ok || !strings.HasPrefix(option, "-") {
			break
		}
		if strings.HasPrefix(option, "--") || !strings.Contains(option[1:], "c") {
			continue
		}
		if index+1 >= len(words) {
			return "shell command-string option is missing its command."
		}
		command, ok := staticWord(words[index+1])
		if !ok {
			return "dynamic shell command strings cannot be classified safely."
		}
		return checkShell(command, depth+1)
	}
	return checkEnvArguments("bash", words)
}

func checkGit(words []*syntax.Word) string {
	for len(words) > 0 {
		value, ok := staticWord(words[0])
		if !ok || !strings.HasPrefix(value, "-") {
			break
		}
		if value == "-C" || value == "-c" || value == "--git-dir" || value == "--work-tree" || value == "--namespace" || value == "--config-env" {
			if len(words) < 2 {
				return "git global option is missing its value."
			}
			words = words[2:]
		} else {
			words = words[1:]
		}
	}
	if len(words) == 0 {
		return ""
	}
	subcommand, _ := staticWord(words[0])
	if subcommand == "remote" && len(words) > 1 && literalEquals(words[1], "add") {
		return "git remote add widens repository access; ask the user first."
	}
	for _, word := range words[1:] {
		value, ok := staticWord(word)
		if !ok {
			continue
		}
		switch value {
		case "--no-verify", "--no-hooks", "--no-pre-commit-hook":
			return "git hook bypass flags are prohibited; run the hooks or surface the failing hook."
		case "-f", "--force", "--force-with-lease":
			if subcommand == "push" {
				return "force-push is prohibited unless the user explicitly authorizes it."
			}
		}
		if subcommand == "push" && strings.HasPrefix(value, "--force-with-lease=") {
			return "force-push is prohibited unless the user explicitly authorizes it."
		}
	}
	return ""
}

func staticWord(word *syntax.Word) (string, bool) {
	if word == nil {
		return "", false
	}
	var value strings.Builder
	for _, part := range word.Parts {
		switch typed := part.(type) {
		case *syntax.Lit:
			value.WriteString(typed.Value)
		case *syntax.SglQuoted:
			value.WriteString(typed.Value)
		case *syntax.DblQuoted:
			for _, quotedPart := range typed.Parts {
				literal, ok := quotedPart.(*syntax.Lit)
				if !ok {
					return "", false
				}
				value.WriteString(literal.Value)
			}
		default:
			return "", false
		}
	}
	return value.String(), true
}

func literalEquals(word *syntax.Word, expected string) bool {
	value, ok := staticWord(word)
	return ok && value == expected
}

func operatorEnvrcWord(word *syntax.Word) bool {
	if word == nil || len(word.Parts) != 1 {
		return false
	}
	if literal, ok := word.Parts[0].(*syntax.Lit); ok {
		return literal.Value == "~/.envrc"
	}
	quoted, ok := word.Parts[0].(*syntax.DblQuoted)
	if !ok || len(quoted.Parts) != 2 {
		return false
	}
	home, ok := quoted.Parts[0].(*syntax.ParamExp)
	if !ok || home.Param == nil || home.Param.Value != "HOME" || home.Excl || home.Length || home.Width || home.Index != nil || home.Slice != nil || home.Repl != nil || home.Names != 0 || home.Exp != nil {
		return false
	}
	suffix, ok := quoted.Parts[1].(*syntax.Lit)
	return ok && suffix.Value == "/.envrc"
}

func hasLiteral(words []*syntax.Word, expected string) bool {
	for _, word := range words {
		if literalEquals(word, expected) {
			return true
		}
	}
	return false
}

func hasAnyLiteral(words []*syntax.Word, expected ...string) bool {
	for _, candidate := range expected {
		if hasLiteral(words, candidate) {
			return true
		}
	}
	return false
}

func exactEnvFile(value string) bool {
	value = strings.TrimPrefix(value, "file://")
	value = strings.Trim(value, "<>")
	value = strings.ReplaceAll(value, "\\", "/")
	return path.Base(value) == ".env"
}

func checkEnvArguments(executable string, words []*syntax.Word) string {
	if executable == "echo" || executable == "printf" {
		return ""
	}
	if executable == "rg" || executable == "grep" {
		return checkSearchArguments(words)
	}
	for _, word := range words {
		if value, ok := staticWord(word); ok && exactEnvFile(value) {
			return "reading a file whose exact basename is .env is prohibited."
		}
	}
	return ""
}

func checkSearchArguments(words []*syntax.Word) string {
	patternSeen := false
	optionValue := ""
	for _, word := range words {
		value, ok := staticWord(word)
		if !ok {
			continue
		}
		if optionValue != "" {
			if optionValue == "file" && exactEnvFile(value) {
				return "reading a file whose exact basename is .env is prohibited."
			}
			patternSeen = true
			optionValue = ""
			continue
		}
		switch {
		case value == "-e" || value == "--regexp":
			optionValue = "pattern"
		case value == "-f" || value == "--file":
			optionValue = "file"
		case strings.HasPrefix(value, "--regexp="):
			patternSeen = true
		case strings.HasPrefix(value, "--file="):
			if exactEnvFile(strings.TrimPrefix(value, "--file=")) {
				return "reading a file whose exact basename is .env is prohibited."
			}
		case strings.HasPrefix(value, "-"):
		case !patternSeen:
			patternSeen = true
		case exactEnvFile(value):
			return "reading a file whose exact basename is .env is prohibited."
		}
	}
	return ""
}

func unwrap(wrapper string, words []*syntax.Word) ([]*syntax.Word, string, bool) {
	for len(words) > 0 {
		value, ok := staticWord(words[0])
		if !ok {
			return nil, "dynamic wrapper arguments cannot be classified safely.", false
		}
		if isAssignment(value) {
			words = words[1:]
			continue
		}
		if wrapper == "command" && (value == "-v" || value == "-V") {
			return nil, "", false
		}
		if value == "--" {
			return words[1:], "", true
		}
		if wrapperFlag(wrapper, value) {
			words = words[1:]
			continue
		}
		if wrapperValueFlag(wrapper, value) {
			if len(words) < 2 {
				return nil, "wrapper option is missing its value.", false
			}
			words = words[2:]
			continue
		}
		if wrapperEqualsFlag(wrapper, value) {
			words = words[1:]
			continue
		}
		if strings.HasPrefix(value, "-") {
			return nil, "unrecognized " + wrapper + " wrapper option prevents safe command classification.", false
		}
		return words, "", true
	}
	return nil, "", true
}

func isAssignment(value string) bool {
	separator := strings.IndexByte(value, '=')
	if separator < 1 {
		return false
	}
	for index, character := range value[:separator] {
		if character == '_' || character >= 'A' && character <= 'Z' || character >= 'a' && character <= 'z' {
			continue
		}
		if index > 0 && character >= '0' && character <= '9' {
			continue
		}
		return false
	}
	return true
}

func wrapperFlag(wrapper, value string) bool {
	switch wrapper + ":" + value {
	case "env:-i", "env:--ignore-environment", "env:-0", "env:--null", "env:--debug",
		"command:-p", "time:-p",
		"sudo:-n", "sudo:--non-interactive", "sudo:-E", "sudo:--preserve-env",
		"sudo:-H", "sudo:--set-home", "sudo:-S", "sudo:--stdin",
		"sudo:-b", "sudo:--background", "sudo:-k", "sudo:--reset-timestamp",
		"sudo:-K", "sudo:--remove-timestamp", "sudo:-v", "sudo:--validate",
		"sudo:-l", "sudo:--list", "sudo:-V", "sudo:--version",
		"sudo:-e", "sudo:--edit":
		return true
	}
	return false
}

func wrapperValueFlag(wrapper, value string) bool {
	switch wrapper + ":" + value {
	case "env:-u", "env:--unset", "env:-C", "env:--chdir",
		"sudo:-u", "sudo:--user", "sudo:-g", "sudo:--group",
		"sudo:-h", "sudo:--host", "sudo:-p", "sudo:--prompt",
		"sudo:-C", "sudo:--close-from", "sudo:-R", "sudo:--chroot",
		"sudo:-T", "sudo:--command-timeout":
		return true
	}
	return false
}

func wrapperEqualsFlag(wrapper, value string) bool {
	for _, option := range []string{"--unset=", "--chdir=", "--user=", "--group=", "--host=", "--prompt=", "--close-from=", "--chroot=", "--command-timeout="} {
		if strings.HasPrefix(value, option) && (wrapper == "env" || wrapper == "sudo") {
			return true
		}
	}
	return false
}
