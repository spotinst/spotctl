// Copyright 2021 NetApp, Inc. All Rights Reserved.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceSystem is the system namespace where we place Ocean components.
const NamespaceSystem = "spot-system"

// OceanComponentType represents the type of OceanComponent.
type OceanComponentType string

// These are valid component types.
const (
	OceanComponentTypeHelm OceanComponentType = "Helm"
)

func (x OceanComponentType) String() string { return string(x) }

// OceanComponentState represents the state of OceanComponent.
type OceanComponentState string

// These are valid component states.
const (
	OceanComponentStatePresent OceanComponentState = "Present"
	OceanComponentStateAbsent  OceanComponentState = "Absent"
)

func (x OceanComponentState) String() string { return string(x) }

// OceanComponentName represents the name of OceanComponent.
type OceanComponentName string

// These are valid component names.
const (
	MetricsServerComponentName         OceanComponentName = "metrics-server"
	OceanControllerComponentName       OceanComponentName = "ocean-controller"
	LegacyOceanControllerComponentName OceanComponentName = "spotinst-kubernetes-cluster-controller"
)

func (x OceanComponentName) String() string { return string(x) }

// OceanComponentConditionType represents the type of OceanComponentCondition.
type OceanComponentConditionType string

// These are valid component conditions.
const (
	// OceanComponentConditionTypeAvailable means the application is available.
	OceanComponentConditionTypeAvailable OceanComponentConditionType = "Available"
	// OceanComponentConditionTypeProgressing means the component is progressing.
	OceanComponentConditionTypeProgressing OceanComponentConditionType = "Progressing"
	// OceanComponentConditionTypeDegraded indicates the component is in a temporary degraded state.
	OceanComponentConditionTypeDegraded OceanComponentConditionType = "Degraded"
	// OceanComponentConditionTypeFailure indicates a significant error conditions.
	OceanComponentConditionTypeFailure OceanComponentConditionType = "Failing"
)

func (x OceanComponentConditionType) String() string { return string(x) }

// OceanComponentCondition describes the state of a deployment at a certain point.
type OceanComponentCondition struct {
	// Type of deployment condition.
	Type OceanComponentConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=OceanComponentConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty" protobuf:"bytes,6,opt,name=lastUpdateTime"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,7,opt,name=lastTransitionTime"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
}

// OceanComponentSpec defines the desired state of OceanComponent.
type OceanComponentSpec struct {
	// Type is one of ["Helm"].
	Type OceanComponentType `json:"type"`
	// Name is the name of the OceanComponent.
	Name OceanComponentName `json:"name"`
	// State determines whether the component should be installed or removed.
	State OceanComponentState `json:"state"`
	// URL is the location of the OceanComponent archive file.
	URL string `json:"url"`
	// Version is a SemVer 2 conformant version string of the OceanComponent archive file.
	Version string `json:"version"`
	// Values is the set of extra values added to the OceanComponent.
	Values string `json:"values,omitempty"`
}

// OceanComponentStatus defines the observed state of OceanComponent.
type OceanComponentStatus struct {
	// A set of installation values specific to the component
	Properties map[string]string         `json:"properties,omitempty"`
	Conditions []OceanComponentCondition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=oc,path=oceancomponents

// OceanComponent is the Schema for the OceanComponent API
type OceanComponent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OceanComponentSpec   `json:"spec,omitempty"`
	Status OceanComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OceanComponentList contains a list of OceanComponent
type OceanComponentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OceanComponent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OceanComponent{}, &OceanComponentList{})
}
