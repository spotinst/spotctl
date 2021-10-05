// Copyright 2021 NetApp, Inc. All Rights Reserved.

package tide

import (
	oceanv1alpha1 "github.com/spotinst/ocean-operator/api/v1alpha1"
)

// region Options

// ApplyOptions contains apply options.
type ApplyOptions struct {
	Namespace        string
	ComponentsFilter map[oceanv1alpha1.OceanComponentName]struct{}
}

// DeleteOptions contains delete options.
type DeleteOptions struct {
	Namespace string
}

// ChartOptions contains Helm chart options.
type ChartOptions struct {
	Name      string
	Namespace string
	URL       string
	Version   string
	Values    string
}

// endregion

// region Interfaces

// ApplyOption is some configuration that modifies options for an apply request.
type ApplyOption interface {
	// MutateApplyOptions applies this configuration to the given ApplyOptions.
	MutateApplyOptions(options *ApplyOptions)
}

// DeleteOption is some configuration that modifies options for a delete request.
type DeleteOption interface {
	// MutateDeleteOptions applies this configuration to the given DeleteOptions.
	MutateDeleteOptions(options *DeleteOptions)
}

// ChartOption is some configuration that modifies options for a Helm chart.
type ChartOption interface {
	// MutateChartOptions applies this configuration to the given ChartOptions.
	MutateChartOptions(options *ChartOptions)
}

// endregion

// region Adapters

// ApplyOptionFunc is a convenience type like http.HandlerFunc.
type ApplyOptionFunc func(options *ApplyOptions)

// MutateApplyOptions implements the ApplyOption interface.
func (f ApplyOptionFunc) MutateApplyOptions(options *ApplyOptions) { f(options) }

// DeleteOptionFunc is a convenience type like http.HandlerFunc.
type DeleteOptionFunc func(options *DeleteOptions)

// MutateDeleteOptions implements the DeleteOption interface.
func (f DeleteOptionFunc) MutateDeleteOptions(options *DeleteOptions) { f(options) }

// ChartOptionFunc is a convenience type like http.HandlerFunc.
type ChartOptionFunc func(options *ChartOptions)

// MutateChartOptions implements the ChartOption interface.
func (f ChartOptionFunc) MutateChartOptions(options *ChartOptions) {
	f(options)
}

// endregion

// region "Functional" Options

// WithNamespace sets the given namespace.
func WithNamespace(namespace string) Namespace {
	return Namespace(namespace)
}

// Namespace determines where components should be applied or deleted.
type Namespace string

// MutateApplyOptions implements the ApplyOption interface.
func (w Namespace) MutateApplyOptions(options *ApplyOptions) {
	options.Namespace = string(w)
}

// MutateDeleteOptions implements the DeleteOption interface.
func (w Namespace) MutateDeleteOptions(options *DeleteOptions) {
	options.Namespace = string(w)
}

// Blank assignments to verify that Namespace implements both ApplyOption and DeleteOption.
var (
	_ ApplyOption  = Namespace("")
	_ DeleteOption = Namespace("")
)

// WithComponentsFilter sets the given ComponentsFilter list.
func WithComponentsFilter(components ...oceanv1alpha1.OceanComponentName) ComponentsFilter {
	return ComponentsFilter{
		components: components,
	}
}

// ComponentsFilter filters components to be applied or deleted.
type ComponentsFilter struct {
	components []oceanv1alpha1.OceanComponentName
}

// MutateApplyOptions implements the ApplyOption interface.
func (w ComponentsFilter) MutateApplyOptions(options *ApplyOptions) {
	options.ComponentsFilter = make(map[oceanv1alpha1.OceanComponentName]struct{})
	for _, component := range w.components {
		options.ComponentsFilter[component] = struct{}{}
	}
}

// Blank assignment to verify that ComponentsFilter implements ApplyOption.
var _ ApplyOption = ComponentsFilter{}

// WithChartName sets the given chart name.
func WithChartName(name string) ChartOption {
	return ChartOptionFunc(func(options *ChartOptions) {
		options.Name = name
	})
}

// WithChartNamespace sets the given chart namespace.
func WithChartNamespace(namespace string) ChartOption {
	return ChartOptionFunc(func(options *ChartOptions) {
		options.Namespace = namespace
	})
}

// WithChartURL sets the given chart URL.
func WithChartURL(url string) ChartOption {
	return ChartOptionFunc(func(options *ChartOptions) {
		options.URL = url
	})
}

// WithChartVersion sets the given chart version.
func WithChartVersion(version string) ChartOption {
	return ChartOptionFunc(func(options *ChartOptions) {
		options.Version = version
	})
}

// WithChartValues sets the given chart values.
func WithChartValues(values string) ChartOption {
	return ChartOptionFunc(func(options *ChartOptions) {
		options.Values = values
	})
}

// endregion

// region Helpers

func mutateApplyOptions(options ...ApplyOption) *ApplyOptions {
	opts := new(ApplyOptions)
	for _, opt := range options {
		opt.MutateApplyOptions(opts)
	}
	return opts
}

func mutateDeleteOptions(options ...DeleteOption) *DeleteOptions {
	opts := new(DeleteOptions)
	for _, opt := range options {
		opt.MutateDeleteOptions(opts)
	}
	return opts
}

func mutateChartOptions(options ...ChartOption) *ChartOptions {
	opts := new(ChartOptions)
	for _, opt := range options {
		opt.MutateChartOptions(opts)
	}
	return opts
}

// endregion
