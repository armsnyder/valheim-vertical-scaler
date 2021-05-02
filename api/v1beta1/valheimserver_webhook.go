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
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var valheimserverlog = logf.Log.WithName("valheimserver-resource")

func (r *ValheimServer) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-valheim-zingerweb-services-v1beta1-valheimserver,mutating=true,failurePolicy=fail,sideEffects=None,groups=valheim.zingerweb.services,resources=valheimservers,verbs=create;update,versions=v1beta1,name=mvalheimserver.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &ValheimServer{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ValheimServer) Default() {
	valheimserverlog.Info("default", "name", r.Name)

	if r.Spec.ServerName == "" {
		valheimserverlog.V(1).Info("configuring serverName default")
		r.Spec.ServerName = r.Name
	}

	if r.Spec.AWSSecretName == "" {
		valheimserverlog.V(1).Info("configuring awsSecretName default")
		r.Spec.AWSSecretName = r.Name
	}

	if r.Spec.PasswordSecretName == "" {
		valheimserverlog.V(1).Info("configuring passwordSecretName default")
		r.Spec.PasswordSecretName = r.Name
	}

	if reflect.DeepEqual(r.Spec.Resources, corev1.ResourceRequirements{}) {
		valheimserverlog.V(1).Info("configuring resources default")
		r.Spec.Resources = corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("4Gi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("4Gi"),
			},
		}
	}
}
