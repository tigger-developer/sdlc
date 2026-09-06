package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var buildRelease string

const previewOwnershipMarker = `<meta name="generator" content="sdlc-preview">`

func main() {
	if status := run(os.Args[1:], os.Stdout, os.Stderr); status != 0 {
		os.Exit(status)
	}
}

func run(arguments []string, output, errorOutput io.Writer) int {
	flags := flag.NewFlagSet("sdlc-preview", flag.ContinueOnError)
	flags.SetOutput(errorOutput)
	cleanup := flags.String("cleanup", "", "internal delayed-cleanup path")
	version := flags.Bool("version", false, "print the command version")
	flags.Usage = func() {
		fmt.Fprintln(errorOutput, "usage: sdlc-preview FILE")
		fmt.Fprintln(errorOutput, "Render one Markdown or Org document with Pandoc and open it in a browser.")
	}
	if err := flags.Parse(arguments); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if *cleanup != "" {
		time.Sleep(time.Second)
		if err := removeGeneratedPreview(*cleanup); err != nil && !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(errorOutput, "sdlc-preview: cleaning preview: %v\n", err)
			return 1
		}
		return 0
	}
	if *version {
		value := buildRelease
		if value == "" {
			value = "devel"
		}
		fmt.Fprintf(output, "sdlc-preview %s\n", value)
		return 0
	}
	if flags.NArg() != 1 {
		flags.Usage()
		return 2
	}
	if err := preview(flags.Arg(0), output, errorOutput); err != nil {
		fmt.Fprintf(errorOutput, "sdlc-preview: %v\n", err)
		return 1
	}
	return 0
}

func removeGeneratedPreview(path string) error {
	name := filepath.Base(path)
	if !strings.HasPrefix(name, ".sdlc-preview-") || filepath.Ext(name) != ".html" {
		return errors.New("refusing to remove a file not named as an SDLC preview")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return errors.New("refusing to remove a non-regular preview path")
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !bytes.Contains(contents, []byte(previewOwnershipMarker)) {
		return errors.New("refusing to remove a preview without the SDLC ownership marker")
	}
	return os.Remove(path)
}

func preview(input string, output, errorOutput io.Writer) error {
	source, err := filepath.Abs(input)
	if err != nil {
		return err
	}
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("reading source %q: %w", source, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("source is not a regular file: %s", source)
	}
	format, err := inputFormat(source)
	if err != nil {
		return err
	}
	pandoc, err := exec.LookPath("pandoc")
	if err != nil {
		return errors.New("pandoc is required and was not found on PATH")
	}
	template, err := previewTemplate()
	if err != nil {
		return err
	}
	previewFile, err := os.CreateTemp(filepath.Dir(source), ".sdlc-preview-*.html")
	if err != nil {
		return fmt.Errorf("creating preview beside source: %w", err)
	}
	previewPath := previewFile.Name()
	if err := previewFile.Close(); err != nil {
		_ = os.Remove(previewPath)
		return err
	}
	keep := false
	defer func() {
		if !keep {
			_ = os.Remove(previewPath)
		}
	}()
	command := exec.Command(pandoc, "--from", format, "--standalone", "--template", template, "--metadata", "title="+filepath.Base(source), "--output", previewPath, source)
	command.Stdout = output
	command.Stderr = errorOutput
	if err := command.Run(); err != nil {
		return fmt.Errorf("rendering %s: %w", filepath.Base(source), err)
	}
	opener, openerArguments, err := browserCommand(previewPath)
	if err != nil {
		return err
	}
	open := exec.Command(opener, openerArguments...)
	open.Stdout = output
	open.Stderr = errorOutput
	if err := open.Run(); err != nil {
		return fmt.Errorf("opening browser: %w", err)
	}
	if err := startCleanup(previewPath); err != nil {
		return err
	}
	keep = true
	fmt.Fprintf(output, "Opened %s\n", filepath.Base(source))
	return nil
}

func inputFormat(path string) (string, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".markdown":
		return "gfm", nil
	case ".org":
		return "org", nil
	default:
		return "", errors.New("input must be Markdown (.md or .markdown) or Org (.org)")
	}
}

func previewTemplate() (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(userHome, ".agents", "sdlc", "templates", "preview.html")
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("preview template unavailable at %s: %w", path, err)
	}
	return path, nil
}

func browserCommand(path string) (string, []string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "open", []string{path}, nil
	case "linux":
		return "xdg-open", []string{path}, nil
	default:
		return "", nil, fmt.Errorf("unsupported platform %s; supported platforms are macOS and Linux", runtime.GOOS)
	}
}

func startCleanup(path string) error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locating cleanup helper: %w", err)
	}
	command := exec.Command(executable, "--cleanup", path)
	command.Stdin = nil
	command.Stdout = nil
	command.Stderr = nil
	if err := command.Start(); err != nil {
		return fmt.Errorf("starting delayed preview cleanup: %w", err)
	}
	if err := command.Process.Release(); err != nil {
		return fmt.Errorf("releasing delayed preview cleanup: %w", err)
	}
	return nil
}
