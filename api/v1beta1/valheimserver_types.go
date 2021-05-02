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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ValheimServerSpec defines the desired state of ValheimServer
type ValheimServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Valheim world name, matching the name of the world save data. Defaults to "Dedicated" if left
	// blank.
	//+kubebuilder:default=Dedicated
	//+optional
	WorldName string `json:"worldName,omitempty"`

	// Valheim server name, as the server appears in the community servers list. Defaults to the
	// same name as this ValheimServer if left blank.
	//+optional
	ServerName string `json:"serverName,omitempty"`

	// Secret containing the keys "awsAccessKeyId" and "awsSecretAccessKey".
	// "awsAccessKeyId" and "awsSecretAccessKey" are the AWS credentials used for managing
	// Route53 and EC2.
	// Defaults to the same name as this ValheimServer if left blank.
	//+optional
	PasswordSecretName string `json:"passwordSecretName,omitempty"`

	// Secret containing the key "serverPassword".
	// "serverPassword" is the password for joining the Valheim server.
	// Defaults to the same name as this ValheimServer if left blank.
	//+optional
	AWSSecretName string `json:"awsSecretName,omitempty"`

	// First port number. This number and the following two port numbers will be used. Defaults to
	// the default Valheim server port 2456 if left blank.
	//+kubebuilder:default=2456
	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:validation:Maximum=65535
	//+optional
	Port int32 `json:"port,omitempty"`

	// Desired server location (AWS or K8s). May be modified to trigger a server migration. Defaults
	// to "K8s" if left blank.
	//+kubebuilder:default=K8s
	//+optional
	Location ValheimServerLocation `json:"location,omitempty"`

	// AWS region for the EC2 instance. Defaults to "us-east-1" if left blank.
	//+kubebuilder:default=us-east-1
	//+optional
	AWSRegion string `json:"awsRegion,omitempty"`

	// Resource requirements for the server pod. Default is 2 CPU and 4Gi memory, for both requests
	// and limits, if left blank.
	//+optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Backup settings for when the server is running in the Kubernetes cluster. No auto-backups
	// will be performed if left blank.
	//+optional
	Backup ValheimServerBackup `json:"backup,omitempty"`

	// Externally accessible hostname or IP for traffic to the server on the Kubernetes cluster.
	// Required for automatic external DNS management.
	//+optional
	ClusterAddress string `json:"clusterAddress,omitempty"`

	// Externally accessible hostname, managed in AWS Route53, which will have its records changed
	// when the server changes locations. Required for automatic external DNS management.
	//+optional
	AdvertisedHostname string `json:"advertisedHostname,omitempty"`
}

// ValheimServerStatus defines the observed state of ValheimServer
type ValheimServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Observations of the ValheimServer's current state, including any errors.
	//+patchMergeKey=type
	//+patchStrategy=merge
	//+listType=map
	//+listMapKey=type
	//*optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=vhs

// ValheimServer is the Schema for the valheimservers API
type ValheimServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ValheimServerSpec   `json:"spec,omitempty"`
	Status ValheimServerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ValheimServerList contains a list of ValheimServer
type ValheimServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ValheimServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ValheimServer{}, &ValheimServerList{})
}

//+kubebuilder:validation:Enum=None;AWS;K8s
type ValheimServerLocation string

const (
	ValheimServerLocationNone ValheimServerLocation = "None"
	ValheimServerLocationK8s  ValheimServerLocation = "K8s"
	ValheimServerLocationAWS  ValheimServerLocation = "AWS"
)

type ValheimServerBackup struct {
	// Target of remote backup server in the form "username@ip:path"
	Target string `json:"address"`

	// Name of a Secret of type kubernetes.io/ssh-auth.
	PrivateKeySecretName string `json:"privateKeySecretName"`
}
