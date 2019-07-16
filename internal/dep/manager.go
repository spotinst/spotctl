package dep

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/spotinst/spotinst-cli/internal/log"
	"github.com/spotinst/spotinst-cli/internal/survey"
)

type manager struct {
	survey survey.Interface
}

func New(survey survey.Interface) Interface {
	return &manager{
		survey: survey,
	}
}

func (x *manager) Install(ctx context.Context, dep Dependency, options ...InstallOption) error {
	log.Debugf("Ensuring required dependency %s-%s", dep.Name, dep.Version)

	opts := initDefaultOptions()
	for _, opt := range options {
		opt(opts)
	}

	if path, err := x.lookup(ctx, dep, opts.BinaryDir); err == nil && path != "" {
		log.Debugf("Dependency already installed: %s (%s)", dep.Name, path)
		return nil
	}

	return x.installWithConfirm(ctx, dep, opts)
}

func (x *manager) InstallBulk(ctx context.Context, deps []Dependency, options ...InstallOption) error {
	log.Debugf("Ensuring required dependencies...")

	opts := initDefaultOptions()
	for _, opt := range options {
		opt(opts)
	}

	var missing []Dependency
	for _, dep := range deps {
		if path, _ := x.lookup(ctx, dep, opts.BinaryDir); path != "" {
			log.Debugf("Dependency already installed: %s (%s)", dep.Name, path)
			continue
		}

		missing = append(missing, dep)
	}

	return x.installWithSelect(ctx, missing, opts)
}

// Lookup looks up for an executable named file in the directories named by the
// PATH environment variable. If found, it returns the absolute path to the binary
// executable file. Otherwise, an error is returned.
func (x *manager) lookup(ctx context.Context, dep Dependency, binaryDir string) (string, error) {
	if err := x.initPath(binaryDir); err != nil {
		return "", err
	}

	return exec.LookPath(x.depFilename(dep))
}

var initPathOnce sync.Once

func (x *manager) initPath(binaryDir string) error {
	var err error

	initPathOnce.Do(func() {
		log.Debugf("Initializing PATH by adding %s", binaryDir)

		path := os.Getenv("PATH")
		dirs := filepath.SplitList(path)

		for _, dir := range dirs {
			if dir == binaryDir {
				return // already exists
			}
		}

		err = os.Setenv("PATH", strings.Join(append([]string{binaryDir}, dirs...), ":"))
	})

	return err
}

func (x *manager) installWithConfirm(ctx context.Context, dep Dependency, opts *InstallOptions) error {
	if !opts.Noninteractive {
		if ok := x.confirm(dep); !ok {
			log.Debugf("Aborting installation of dependency %s-%s", dep.Name, dep.Version)
			return nil
		}
	}

	return x.install(ctx, dep, opts)
}

func (x *manager) installWithSelect(ctx context.Context, deps []Dependency, opts *InstallOptions) error {
	if !opts.Noninteractive {
		deps = x.selectMulti(deps)
	}

	for _, dep := range deps {
		if err := x.install(ctx, dep, opts); err != nil {
			return err
		}
	}

	return nil
}

func (x *manager) install(ctx context.Context, dep Dependency, opts *InstallOptions) error {
	if opts.DryRun {
		log.Debugf("Would install %s-%s to %s", dep.Name, dep.Version, opts.BinaryDir)
		return nil
	}

	log.Debugf("Installing dependency %s-%s", dep.Name, dep.Version)

	url, err := x.depUrl(dep)
	if err != nil {
		return err
	}

	return download(url, filepath.Join(opts.BinaryDir, x.depFilename(dep)))
}

func (x *manager) confirm(dep Dependency) bool {
	input := &survey.Input{
		Message: fmt.Sprintf("Install missing required dependency: %s", dep.Name),
		Help:    "Sspotinst CLI would like to install missing required dependency",
		Default: "true",
	}

	ok, _ := x.survey.Confirm(input)
	return ok
}

func (x *manager) selectMulti(deps []Dependency) []Dependency {
	var out []Dependency

	depOpts := make([]interface{}, len(deps))
	for i, dep := range deps {
		depOpts[i] = dep.Name
	}

	input := &survey.Select{
		Message:  "Install missing required dependencies (deselect to avoid auto installing)",
		Help:     "Sspotinst CLI would like to install missing required dependencies",
		Options:  depOpts,
		Defaults: depOpts,
	}

	depNames, err := x.survey.SelectMulti(input)
	if err != nil {
		return out
	}

	depMap := make(map[string]Dependency)
	for _, dep := range deps {
		depMap[dep.Name] = dep
	}

	for _, name := range depNames {
		if dep, ok := depMap[name]; ok {
			out = append(out, dep)
		}
	}

	return out
}

func (x *manager) depUrl(dep Dependency) (string, error) {
	tmpl, err := template.New(dep.Name).Parse(dep.URL)
	if err != nil {
		return "", err
	}

	variables := map[string]string{
		"version":   dep.Version,
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
		"extension": x.depExtension(),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (x *manager) depExtension() string {
	var extension string
	if runtime.GOOS == "windows" {
		extension = ".exe"
	}

	return extension
}

func (x *manager) depFilename(dep Dependency) string {
	return fmt.Sprintf("%s%s", dep.Name, x.depExtension())
}
