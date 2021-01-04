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

type SparkHeritage string

const (
	SparkHeritageSubmit   SparkHeritage = "spark-submit"
	SparkHeritageOperator SparkHeritage = "spark-operator"
	SparkHeritageJupyter  SparkHeritage = "jupyter-notebook"
)

// SparkApplicationSpec defines the desired state of SparkApplication
type SparkApplicationSpec struct {

	//uniquely identifies the spark application, and is shared as a label on all driver and executor pods
	ApplicationId string `json:"applicationId"`

	//the name of the spark application
	ApplicationName string `json:"applicationName"`

	//specifies whether the application originates from spark-operator, from a jupyter notebook, or from spark-submit directly
	Heritage SparkHeritage `json:"heritage"`
}

// SparkApplicationStatus defines the observed state of SparkApplication
type SparkApplicationStatus struct {
	//summarizes information about the spark application
	Data SparkApplicationData `json:"data"`
}

//SparkApplicationData
type SparkApplicationData struct {

	//the runtime configuration of the spark application
	SparkProperties map[string]string `json:"sparkProperties"`

	//collects statistics of the application run
	RunStatistics Statistics `json:"runStatistics"`

	//a reference to the driver pod
	Driver Pod `json:"driver"`

	//a list of references to the executor pods
	Executors []Pod `json:"executors"`
}

type Statistics struct {

	//the list of attempts to run the application
	Attempts []Attempt `json:"attempts"`

	//the network traffic read into the pods
	TotalInputBytes int64 `json:"totalInputBytes"`

	//the network traffic written from the pods
	TotalOutputBytes int64 `json:"totalOutputBytes"`

	//the total executor time in the attempt
	TotalExecutorCpuTime int64 `json:"totalExecutorCpuTime"`
}

type Attempt struct {

	//the unix timestamp of application start
	StartTimeEpoch int64 `json:"startTimeEpoch"`

	//the unix timestamp of application end
	EndTimeEpoch int64 `json:"endTimeEpoch"`

	//the unix timestamp of application update
	LastUpdatedEpoch int64 `json:"lastUpdatedEpoch"`

	//indicates success or failure
	Completed bool `json:"completed"`

	//the application spark version
	AppSparkVersion string `json:"appSparkVersion"`
}

type Pod struct {
	//the name of the pod
	Name string `json:"podName"`
	//the namespace of the pod
	Namespace string `json:"podNamespace"`
	//the kubernetes object UID
	UID string `json:"podUid"`
	//the phase of the pod
	Phase v1.PodPhase `json:"phase"`
	//the set of container statuses
	Statuses []v1.ContainerStatus `json:"containerStatuses"`
	//has the pod been marked as deleted
	Deleted bool `json:"deleted"`
}

// +kubebuilder:object:root=true

// SparkApplication is the Schema for the SparkApplications API
type SparkApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SparkApplicationSpec   `json:"spec,omitempty"`
	Status SparkApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SparkApplicationList contains a list of SparkApplication
type SparkApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SparkApplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SparkApplication{}, &SparkApplicationList{})
}
