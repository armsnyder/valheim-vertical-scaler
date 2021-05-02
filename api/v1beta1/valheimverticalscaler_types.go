/*
Copyright 2021.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ValheimVerticalScalerSpec defines the desired state of ValheimVerticalScaler
type ValheimVerticalScalerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Desired state of the vertical scaler.
	//+kubebuilder:validation:Enum=Down;Up
	//+kubebuilder:default=Down
	//+optional
	Scale string `json:"scale,omitempty"`

	// Configuration pertaining to AWS.
	AWS AWSConfig `json:"aws"`

	// Configuration pertaining to the Valheim server Deployment in the local Kubernetes cluster.
	K8sDeployment K8sDeploymentConfig `json:"k8sDeployment"`
}

// ValheimVerticalScalerStatus defines the observed state of ValheimVerticalScaler
type ValheimVerticalScalerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Basic state of the Valheim server.
	//+kubebuilder:validation:Enum=Unknown;Error;ScalingUp;Up;ScalingDown;Down
	//+kubebuilder:default=Unknown
	//+optional
	State State `json:"state"`

	// Human readable error message if the scaling has reached an error end state.
	//+optional
	Error string `json:"error,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=vvs

// ValheimVerticalScaler is the Schema for the valheimverticalscalers API
type ValheimVerticalScaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ValheimVerticalScalerSpec   `json:"spec,omitempty"`
	Status ValheimVerticalScalerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ValheimVerticalScalerList contains a list of ValheimVerticalScaler
type ValheimVerticalScalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ValheimVerticalScaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ValheimVerticalScaler{}, &ValheimVerticalScalerList{})
}

type State string

const (
	StateUnknown     State = "Unknown"
	StateError       State = "Error"
	StateScalingUp   State = "ScalingUp"
	StateUp          State = "Up"
	StateScalingDown State = "ScalingDown"
	StateDown        State = "Down"
)

type K8sDeploymentConfig struct {
	// Name of the Deployment for the Valheim server.
	Name string `json:"name"`

	// Name of the volume within the "volumes" field of the Deployment spec. (Not necessarily the
	// name of the PersistentVolumeClaim.)
	//+kubebuilder:default=gamefiles
	//+optional
	GameFilesVolumeName string `json:"gameFilesVolumeName,omitempty"`
}

type AWSConfig struct {
	// Region of the EC2 instance.
	Region string `json:"region"`

	// Advertised domain of the server. Must live in Route53.
	Domain string `json:"domain"`

	// Name of a Secret containing the keys "accessKeyId" and "secretAccessKey".
	CredentialSecretName string `json:"credentialSecretName"`

	// Name of a Secret containing the private key for connecting to the EC2 instance.
	PrivateKeySecretName string `json:"privateKeySecretName"`

	// EC2 instance ID.
	InstanceID string `json:"instanceID"`
}
