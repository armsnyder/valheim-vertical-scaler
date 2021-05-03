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
	AWS AWS `json:"aws"`

	// Configuration pertaining to the Valheim server Deployment in the local Kubernetes cluster.
	K8sDeployment K8sDeployment `json:"k8sDeployment"`
}

// ValheimVerticalScalerStatus defines the observed state of ValheimVerticalScaler
type ValheimVerticalScalerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Basic state of the Valheim server.
	//+kubebuilder:validation:Enum=Ready;Error;ScalingUp;ScalingDown
	//+optional
	Phase Phase `json:"phase,omitempty"`

	// Human readable error message if the scaling has reached an error end state.
	//+optional
	Error string `json:"error,omitempty"`

	// The generation of the ValheimVerticalScaler object that this status is for.
	//+optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=vvs
//+kubebuilder:printcolumn:name="Scale",type=string,JSONPath=`.spec.scale`
//+kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`

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

type Phase string

const (
	PhaseReady       Phase = "Ready"
	PhaseError       Phase = "Error"
	PhaseScalingUp   Phase = "ScalingUp"
	PhaseScalingDown Phase = "ScalingDown"
)

type K8sDeployment struct {
	// Name of the Deployment for the Valheim server.
	Name string `json:"name"`

	// Name of the volume within the "volumes" field of the Deployment spec. (Not necessarily the
	// name of the PersistentVolumeClaim.)
	//+kubebuilder:default=gamefiles
	//+optional
	GameFilesVolumeName string `json:"gameFilesVolumeName,omitempty"`
}

type AWS struct {
	// Region of the EC2 instance.
	//+kubebuilder:validation:MinLength=1
	//+kubebuilder:validation:Pattern=`^[a-z]{2}-[a-z]{4,}-\d$`
	Region string `json:"region"`

	// Advertised domain of the server. Must live in Route53.
	//+kubebuilder:validation:MinLength=1
	//+kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
	Domain string `json:"domain"`

	// Name of a Secret containing the keys "accessKeyId" and "secretAccessKey".
	//+kubebuilder:validation:MinLength=1
	//+kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
	CredentialSecretName string `json:"credentialSecretName"`

	// Name of a Secret containing the private key for connecting to the EC2 instance.
	//+kubebuilder:validation:MinLength=1
	//+kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
	PrivateKeySecretName string `json:"privateKeySecretName"`

	// EC2 instance ID.
	//+kubebuilder:validation:MinLength=1
	//+kubebuilder:validation:Pattern=`^i-[a-z0-9]{17}$`
	InstanceID string `json:"instanceID"`
}
