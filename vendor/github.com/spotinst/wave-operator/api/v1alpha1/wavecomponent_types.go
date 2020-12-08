/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ComponentType string
type ComponentState string
type ChartName string

const (
	HelmComponentType ComponentType = "helm"

	PresentComponentState ComponentState = "present"
	AbsentComponentState  ComponentState = "absent"

	SparkHistoryChartName      ChartName = "spark-history-server"
	EnterpriseGatewayChartName ChartName = "enterprise-gateway"
	SparkOperatorChartName     ChartName = "sparkoperator"
	WaveIngressChartName       ChartName = "ingress-nginx"
)

// WaveComponentSpec defines the desired state of WaveComponent
type WaveComponentSpec struct {

	// Type is one of ["helm",]
	Type ComponentType `json:"type"`

	// Name is the name of a helm chart
	Name ChartName `json:"name"`

	// State determines whether the component should be installed or removed
	State ComponentState `json:"state"`

	// URL is the location of the helm repository
	URL string `json:"url"`

	// Version is the version of the helm chart
	Version string `json:"version"`

	// ValuesConfiguration is a set of helm values, in yaml form
	ValuesConfiguration string `json:"valuesConfiguration,omitempty"`
}

// WaveComponentStatus defines the observed state of WaveComponent
type WaveComponentStatus struct {

	// A set of installation values specific to the component
	Properties map[string]string `json:"properties,omitempty"`

	Conditions []WaveComponentCondition `json:"conditions,omitempty"`
}

type ConditionStatus string

type WaveComponentConditionType string

// These are valid conditions of a wave component.
const (
	// Available means the application is available
	WaveComponentAvailable WaveComponentConditionType = "Available"
	// Progressing means the component is progressing, including installation and upgrades
	WaveComponentProgressing WaveComponentConditionType = "Progressing"
	// Degraded indicates the component is in a temporary degraded state
	WaveComponentDegraded WaveComponentConditionType = "Degraded"
	// Failing indicates a significant error conditions
	WaveComponentFailure WaveComponentConditionType = "Failing"
)

// WaveComponentCondition describes the state of a deployment at a certain point.
type WaveComponentCondition struct {
	// Type of deployment condition.
	Type WaveComponentConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=WaveComponentConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty" protobuf:"bytes,6,opt,name=lastUpdateTime"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,7,opt,name=lastTransitionTime"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
}

// +kubebuilder:object:root=true

// WaveComponent is the Schema for the wavecomponents API
type WaveComponent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WaveComponentSpec   `json:"spec,omitempty"`
	Status WaveComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WaveComponentList contains a list of WaveComponent
type WaveComponentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WaveComponent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WaveComponent{}, &WaveComponentList{})
}
