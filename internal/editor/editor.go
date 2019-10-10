package editor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spotinst/spotinst-cli/internal/child"
)

const (
	// Default Unix/Linux editor.
	defaultLinuxEditor = "vi"
	defaultLinuxShell  = "/bin/bash"

	// Default Windows editor.
	defaultWindowsEditor = "notepad"
	defaultWindowsShell  = "cmd"
)

type Editor struct {
	// In, Out, and Err represent the respective data streams that the editor
	// may act upon.
	In       io.Reader
	Out, Err io.Writer
}

func New(in io.Reader, out, err io.Writer) Editor {
	return Editor{
		In:  in,
		Out: out,
		Err: err,
	}
}

func (e Editor) args(path string) []string {
	args, shell := defaultEnvEditor()

	if shell {
		last := args[len(args)-1]
		args[len(args)-1] = fmt.Sprintf("%s %q", last, path)
	} else {
		args = append(args, path)
	}

	return args
}

func (e Editor) Open(ctx context.Context, path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	cmdArgs := e.args(abs)
	cmdOptions := []child.CommandOption{
		child.WithStdio(e.In, e.Out, e.Err),
		child.WithArgs(cmdArgs[1:]...),
	}

	cmd := child.NewCommand(cmdArgs[0], cmdOptions...)

	if err := cmd.Run(ctx); err != nil {
		if err, ok := err.(*exec.Error); ok {
			if err.Err == exec.ErrNotFound {
				return fmt.Errorf("unable to launch the editor %q", strings.Join(cmdArgs, " "))
			}
		}
		return fmt.Errorf("there was a problem with the editor %q", strings.Join(cmdArgs, " "))
	}

	return nil
}

func (e Editor) OpenTempFile(ctx context.Context, prefix, suffix string, r io.Reader) ([]byte, string, error) {
	f, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("%s_*%s", prefix, suffix))
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	path := f.Name()
	if _, err := io.Copy(f, r); err != nil {
		os.Remove(path)
		return nil, path, err
	}

	// This file descriptor needs to close so the next process (Open) can claim it.
	f.Close()
	if err := e.Open(ctx, path); err != nil {
		return nil, path, err
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, path, err
	}

	return bytes.TrimSuffix(data, []byte("\n")), path, err
}

func defaultEnvEditor() ([]string, bool) {
	editor := os.Getenv("EDITOR")
	if len(editor) == 0 {
		editor = platformize(defaultLinuxEditor, defaultWindowsEditor)
	}

	if !strings.Contains(editor, " ") {
		return []string{editor}, false
	}

	if !strings.ContainsAny(editor, "\"'\\") {
		return strings.Split(editor, " "), false
	}

	shell := defaultEnvShell()
	return append(shell, editor), true
}

func defaultEnvShell() []string {
	shell := os.Getenv("SHELL")
	if len(shell) == 0 {
		shell = platformize(defaultLinuxShell, defaultWindowsShell)
	}

	flag := "-c"
	if shell == defaultWindowsShell {
		flag = "/C"
	}

	return []string{shell, flag}
}

func platformize(linux, windows string) string {
	if runtime.GOOS == "windows" {
		return windows
	}

	return linux
}
