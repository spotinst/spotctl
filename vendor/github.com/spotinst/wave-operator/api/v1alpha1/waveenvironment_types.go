package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WaveEnvironmentSpec struct {

	// cluster id
	OceanClusterId string `json:"oceanClusterId"`

	// version of Wave Operator
	OperatorVersion string `json:"operatorVersion"`

	// whether Cert Manager was installed when creating Wave
	CertManagerDeployed bool `json:"certManagerDeployed"`

	// whether the K8s cluster was provisioned when creating Wave
	K8sClusterProvisioned bool `json:"k8sClusterProvisioned"`

	// whether the Ocean cluster was provisioned when create Wave
	OceanClusterProvisioned bool `json:"oceanClusterProvisioned"`
}

type WaveEnvironmentStatus struct {
}

// +kubebuilder:object:root=true

// WaveEnvironment is the Schema for the wave environment API
type WaveEnvironment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WaveEnvironmentSpec   `json:"spec,omitempty"`
	Status WaveEnvironmentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WaveEnvironmentList contains a list of WaveEnvironment
type WaveEnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WaveEnvironment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WaveEnvironment{}, &WaveEnvironmentList{})
}
