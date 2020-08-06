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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BambooSpec defines the desired state of Bamboo

type AutoManagement struct {
	Enabled          bool  `json:"enabled,omitempty"`
	MinReplicas      int32 `json:"minReplicas,omitempty"`
	MaxReplicas      int32 `json:"maxReplicas,omitempty"`
	MaxBuildInQueue  int32 `json:"maxBuildInQueue,omitempty"`
	ReplicasToAdd    int32 `json:"replicasToAdd,omitempty"`
	ReplicasToRemove int32 `json:"replicasToRemove,omitempty"`
	MaxIdleAgents    int32 `json:"maxIdleAgents,omitempty"`
}

type RemoteAgents struct {
	Enabled               bool   `json:"enabled,omitempty"`
	ImageRepo             string `json:"imageRepo,omitempty"`
	ImageTag              string `json:"imageTag,omitempty"`
	WrapperJavaInitMemory string `json:"wrapperJavaInitMemory,omitempty"`
	WrapperJavaMaxMemory  string `json:"wrapperJavaMaxMemory,omitempty"`
	ContainerMemRequest   string `json:"containerMemRequest,omitempty"`
	ContainerMemLimit     string `json:"containerMemLimit,omitempty"`
	ContainerCPURequest   string `json:"containerCPURequest,omitempty"`
	ContainerCPULimit     string `json:"containerCPULimit,omitempty"`
	Replicas              int32  `json:"replicas,omitempty"`
	AutoManagement        `json:"autoManagement,omitempty"`
}

type Installer struct {
	AdminName     string `json:"adminName,omitempty"`
	AdminPassword string `json:"adminPassword,omitempty"`
	AdminEmail    string `json:"adminEmail,omitempty"`
	AdminFullName string `json:"adminFullName,omitempty"`
	License       string `json:"license,omitempty"`
}

type Ingress struct {
	Enabled       bool   `json:"enabled,omitempty"`
	Host          string `json:"host,omitempty"`
	Tls           bool   `json:"tls,omitempty"`
	TlsSecretName string `json:"tlsSecretName,omitempty"`
}

type Datasource struct {
	Host     string `json:"host,omitempty"`
	Port     string `json:"port,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Database string `json:"database,omitempty"`
}
type BambooSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ImageRepo                 string `json:"imageRepo,omitempty"`
	ImageTag                  string `json:"imageTag,omitempty"`
	JvmMinimumMemory          string `json:"jvmMinimumMemory,omitempty"`
	JvmMaximumMemory          string `json:"jvmMaximumMemory,omitempty"`
	JvmSupportRecommendedArgs string `json:"jvmSupportRecommendedArgs,omitempty"`
	AtlProxyName              string `json:"atlProxyName,omitempty"`
	AtlProxyPort              string `json:"atlProxyPort,omitempty"`
	AtlProxyScheme            string `json:"atlProxyScheme,omitempty"`
	ContainerMemRequest       string `json:"containerMemRequest,omitempty"`
	ContainerMemLimit         string `json:"containerMemLimit,omitempty"`
	ContainerCPURequest       string `json:"containerCPURequest,omitempty"`
	ContainerCPULimit         string `json:"containerCPULimit,omitempty"`
	Datasource                `json:"datasource,omitempty"`
	Ingress                   `json:"ingress,omitempty"`
	RemoteAgents              `json:"remoteagents,omitempty"`
	Installer                 `json:"installer,omitempty"`
}

// BambooStatus defines the observed state of Bamboo
type BambooStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Ready     bool   `json:"ready,omitempty"`
	Installed bool   `json:"installed,omitempty"`
	URL       string `json:"url,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Bamboo is the Schema for the bambooes API
type Bamboo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BambooSpec   `json:"spec,omitempty"`
	Status BambooStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BambooList contains a list of Bamboo
type BambooList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Bamboo `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Bamboo{}, &BambooList{})
}
