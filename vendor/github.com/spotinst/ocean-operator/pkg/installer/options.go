// Copyright 2021 NetApp, Inc. All Rights Reserved.

package installer

import (
	"github.com/spotinst/ocean-operator/pkg/log"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// region Options

type InstallerOptions struct {
	Namespace    string
	ClientGetter genericclioptions.RESTClientGetter
	DryRun       bool
	Log          log.Logger
}

// endregion

// region Interfaces

// InstallerOption is some configuration that modifies options for an Installer.
type InstallerOption interface {
	// MutateInstallerOptions applies this configuration to the given InstallerOptions.
	MutateInstallerOptions(options *InstallerOptions)
}

// endregion

// region Adapters

// InstallerOptionFunc is a convenience type like http.HandlerFunc.
type InstallerOptionFunc func(options *InstallerOptions)

// MutateInstallerOptions implements the InstallerOption interface.
func (f InstallerOptionFunc) MutateInstallerOptions(options *InstallerOptions) { f(options) }

// endregion

// region "Functional" Options

// WithNamespace sets the given namespace.
func WithNamespace(namespace string) InstallerOption {
	return InstallerOptionFunc(func(options *InstallerOptions) {
		options.Namespace = namespace
	})
}

// WithClientGetter sets the given RESTClientGetter.
func WithClientGetter(getter genericclioptions.RESTClientGetter) InstallerOption {
	return InstallerOptionFunc(func(options *InstallerOptions) {
		options.ClientGetter = getter
	})
}

// WithDryRun sets the dry run flag.
func WithDryRun(dryRun bool) InstallerOption {
	return InstallerOptionFunc(func(options *InstallerOptions) {
		options.DryRun = dryRun
	})
}

// WithLogger sets the given logger.
func WithLogger(log log.Logger) InstallerOption {
	return InstallerOptionFunc(func(options *InstallerOptions) {
		options.Log = log
	})
}

// endregion

// region Helpers

func mutateInstallerOptions(options ...InstallerOption) *InstallerOptions {
	opts := new(InstallerOptions)
	for _, opt := range options {
		opt.MutateInstallerOptions(opts)
	}
	return opts
}

// endregion
