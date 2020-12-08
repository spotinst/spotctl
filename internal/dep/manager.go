package dep

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mholt/archiver/v3"
	"github.com/spotinst/spotctl/internal/log"
	"github.com/spotinst/spotctl/internal/survey"
)

type manager struct {
	survey survey.Interface
}

func NewManager(survey survey.Interface) Manager {
	return &manager{
		survey: survey,
	}
}

func (x *manager) Install(ctx context.Context, dep Dependency, options ...InstallOption) error {
	log.Debugf("Ensuring required dependency %s-%s", dep.Name(), dep.Version())

	opts := initDefaultOptions()
	for _, opt := range options {
		opt(opts)
	}

	path := filepath.Join(opts.BinaryDir, dep.Executable())
	_, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	present := err == nil
	if !shouldInstallDep(opts.InstallPolicy, present) {
		if present {
			log.Debugf("Dependency %q already present on machine (%s)", dep.Name(), path)
			return nil
		}
		return fmt.Errorf("dep: dependency %q is not present with "+
			"install policy of %q", dep.Name(), InstallNever)
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
	var err error
	for _, dep := range deps {
		path := filepath.Join(opts.BinaryDir, dep.Executable())
		_, err = os.Stat(path)
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		present := err == nil
		if !shouldInstallDep(opts.InstallPolicy, present) {
			if present {
				log.Debugf("Dependency %q already present on machine (%s)", dep.Name(), path)
				continue
			}
			return fmt.Errorf("dep: dependency %q is not present with "+
				"install policy of %q", dep.Name(), InstallNever)
		}

		missing = append(missing, dep)
	}

	return x.installWithSelect(ctx, missing, opts)
}

// shouldInstallDep returns whether we should install a Dependency according to
// the presence and install policy.
func shouldInstallDep(installPolicy InstallPolicy, present bool) bool {
	if installPolicy == InstallNever {
		return false
	}

	if installPolicy == InstallAlways ||
		(installPolicy == InstallIfNotPresent && (!present)) {
		return true
	}

	return false
}

// Lookup looks up for an executable named file in the directories named by the
// PATH environment variable. If found, it returns the absolute path to the binary
// executable file. Otherwise, an error is returned.
func (x *manager) lookup(ctx context.Context, dep Dependency, binaryDir string) (string, error) {
	if err := x.initPath(binaryDir); err != nil {
		return "", err
	}
	path, _ := exec.LookPath(dep.Executable())
	return path, nil
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
			log.Debugf("Aborting installation of dependency %s-%s", dep.Name(), dep.Version())
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
		log.Debugf("Would install %s-%s to %s", dep.Name(), dep.Version(), opts.BinaryDir)
		return nil
	}

	log.Debugf("Installing dependency %s-%s", dep.Name(), dep.Version())

	url, err := dep.URL()
	if err != nil {
		return fmt.Errorf("dep: unable to render url: %w", err)
	}

	executable := filepath.Join(opts.BinaryDir, dep.Executable())
	ext, archive := checkArchive(url.Path)

	p := fmt.Sprintf("spotctl_*_%s-%s%s", dep.Name(), dep.Version(), ext)
	f, err := ioutil.TempFile(os.TempDir(), p)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	intermediate := f.Name()

	log.Debugf("Downloading dependency to %s", intermediate)
	if err = download(url, intermediate); err != nil {
		return err
	}

	if archive {
		d := strings.TrimSuffix(intermediate, ext)
		log.Debugf("Unarchiving dependency to %s", d)
		if err = archiver.Unarchive(intermediate, d); err != nil {
			return fmt.Errorf("dep: unable to unarchive file: %w", err)
		}
		defer os.Remove(d)
		fi, err := os.Stat(d)
		if err != nil {
			return fmt.Errorf("dep: unable to describe file: %w", err)
		}
		if fi.Mode().IsDir() {
			intermediate = filepath.Join(d, dep.Executable())
		}
	}

	log.Debugf("Copying dependency from %s", intermediate)
	if err = copyFile(intermediate, executable); err != nil {
		return fmt.Errorf("dep: unable to copy file: %w", err)
	}

	// Make it executable.
	return os.Chmod(executable, 0755)
}

func (x *manager) confirm(dep Dependency) bool {
	input := &survey.Input{
		Message: fmt.Sprintf("Install missing required dependency: %s", dep.Name),
		Help:    "Spot CLI would like to install missing required dependency",
		Default: "true",
	}
	ok, _ := x.survey.Confirm(input)
	return ok
}

func (x *manager) selectMulti(deps []Dependency) []Dependency {
	var out []Dependency

	depOpts := make([]interface{}, len(deps))
	for i, dep := range deps {
		depOpts[i] = dep.Name()
	}

	input := &survey.Select{
		Message:  "Install missing required dependencies (deselect to avoid auto installing)",
		Help:     "Spot CLI would like to install missing required dependencies",
		Options:  depOpts,
		Defaults: depOpts,
	}

	depNames, err := x.survey.SelectMulti(input)
	if err != nil {
		return out
	}

	depMap := make(map[string]Dependency)
	for _, dep := range deps {
		depMap[dep.Name()] = dep
	}

	for _, name := range depNames {
		if dep, ok := depMap[name]; ok {
			out = append(out, dep)
		}
	}

	return out
}
