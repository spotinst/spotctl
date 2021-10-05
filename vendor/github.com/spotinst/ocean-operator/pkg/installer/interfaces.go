// Copyright 2021 NetApp, Inc. All Rights Reserved.

package installer

import (
	"errors"

	oceanv1alpha1 "github.com/spotinst/ocean-operator/api/v1alpha1"
)

var (
	// ErrNotImplemented is the error returned if a method is not implemented.
	ErrNotImplemented = errors.New("installer: not implemented")
	// ErrReleaseNotFound indicates that a component release is not found.
	ErrReleaseNotFound = errors.New("installer: release not found")
)

type (
	// Installer defines the interface of a component installer.
	Installer interface {
		// Get returns details of a component release by name.
		Get(name oceanv1alpha1.OceanComponentName) (*Release, error)
		// Install installs a component to a cluster.
		Install(component *oceanv1alpha1.OceanComponent) (*Release, error)
		// Uninstall uninstalls a component from a cluster.
		Uninstall(component *oceanv1alpha1.OceanComponent) error
		// Upgrade upgrades a component to a cluster.
		Upgrade(component *oceanv1alpha1.OceanComponent) (*Release, error)
		// IsUpgrade determines whether a component release is an upgrade.
		IsUpgrade(component *oceanv1alpha1.OceanComponent, release *Release) bool
	}

	// Release describes a deployment of a component. For Helm-based components,
	// it represents a Helm release (i.e. a chart installed into a cluster).
	Release struct {
		// Name is the name of the release.
		Name string `json:"name,omitempty"`
		// Version is a SemVer 2 conformant version string of the release.
		Version string `json:"version,omitempty"`
		// AppVersion is the version of the application enclosed inside of this release.
		AppVersion string `json:"appVersion,omitempty"`
		// Status is the current state of the release.
		Status ReleaseStatus `json:"status,omitempty"`
		// Description is human-friendly "log entry" about this release.
		Description string `json:"description,omitempty"`
		// Values is the set of extra values added to the release.
		Values map[string]interface{} `json:"values,omitempty"`
		// Manifest is the string representation of the rendered template.
		Manifest string `json:"manifest,omitempty"`
	}
)

// ReleaseStatus is the status of a release.
type ReleaseStatus string

// These are valid release statuses.
const (
	// ReleaseStatusUnknown indicates that a release is in an uncertain state.
	ReleaseStatusUnknown ReleaseStatus = "Unknown"
	// ReleaseStatusDeployed indicates that a release has been deployed to Kubernetes.
	ReleaseStatusDeployed ReleaseStatus = "Deployed"
	// ReleaseStatusUninstalled indicates that a release has been uninstalled from Kubernetes.
	ReleaseStatusUninstalled ReleaseStatus = "Uninstalled"
	// ReleaseStatusFailed indicates that the release was not successfully deployed.
	ReleaseStatusFailed ReleaseStatus = "Failed"
	// ReleaseStatusProgressing indicates that a release is in progress.
	ReleaseStatusProgressing ReleaseStatus = "Progressing"
)

func (x ReleaseStatus) String() string { return string(x) }

// IsNotImplemented returns true if the specified error is ErrNotImplemented.
func IsNotImplemented(err error) bool {
	return errors.Is(err, ErrNotImplemented)
}

// IsReleaseNotFound returns true if the specified error is ErrReleaseNotFound.
func IsReleaseNotFound(err error) bool {
	return errors.Is(err, ErrReleaseNotFound)
}
