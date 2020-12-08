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
	"bytes"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SparkHeritage string
type SparkApplicationState string

const (
	SparkHeritageSubmit   SparkHeritage = "spark-submit"
	SparkHeritageOperator SparkHeritage = "spark-operator"
	SparkHeritageJupyter  SparkHeritage = "jupyter-notebook"

	SparkStateSubmitted SparkApplicationState = "SUBMITTED"
	SparkStateRunning   SparkApplicationState = "RUNNING"
	SparkStateCompleted SparkApplicationState = "COMPLETED"
	SparkStateFailed    SparkApplicationState = "FAILED"
	SparkStateUnknown   SparkApplicationState = "UNKNOWN"
)

// SparkApplicationSpec defines the desired state of SparkApplication
type SparkApplicationSpec struct {

	//uniquely identifies the spark application, and is shared as a label on all driver and executor pods
	ApplicationId string `json:"applicationId"`

	//identity of the Ocean cluster in which the application is running
	ClusterIdentifier string `json:"clusterIdentifier"`

	//specifies whether the application originates from spark-operator, from a jupyter notebook, or from spark-submit directly
	Heritage SparkHeritage `json:"heritage"`
}

// SparkApplicationStatus defines the observed state of SparkApplication
type SparkApplicationStatus struct {
	//the current state of the spark application
	State SparkApplicationState `json:"state"`

	//summarizes the history of the spark application
	Data SparkApplicationData `json:"data"`
}

//SparkApplicationData
type SparkApplicationData struct {

	//the runtime configuration of the driver and executors
	SparkProperties Properties `json:"sparkProperties"`

	//Rcollects statistics of the application runtoime
	RunStatistics Statistics `json:"runStatistics"`

	// a reference to the driver pod
	Driver Pod `json:"driver"`

	//a list of references to the executor pods
	Executors []Pod `json:"executors"`
}

type MemoryMB int64

func (m MemoryMB) MarshalText() (text []byte, err error) { // TODO remove, this is extra just to get the "m" into json
	var b bytes.Buffer
	fmt.Fprintf(&b, "%dm", m)
	return b.Bytes(), nil
}

type Properties struct { // TODO replace with map[string]interface{} or map[string]string
	//the count of executors
	ExecutorInstances int `json:"spark.executor.instances"`

	//the number of cores in the executor pods
	ExecutorCores int `json:"spark.executor.cores"`

	//the executor memory in MB
	ExecutorMemory MemoryMB `json:"spark.executor.memory"`

	//the number of cores in the driver pod
	DriverCores int `json:"spark.driver.cores"`

	//the driver memory in MB
	DriverMemory MemoryMB `json:"spark.driver.memory"`
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
	//the set of container statues
	Statuses []v1.ContainerStatus `json:"containerStatuses"`
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
