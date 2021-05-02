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

package controllers

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	valheimv1beta1 "github.com/armsnyder/valheim-server/api/v1beta1"
)

const (
	labelDomain    = "beta.valheim.zingerweb.services"
	trueLabelValue = "true"

	uploadJobLabelKey     = labelDomain + "/upload-job"
	downloadJobLabelKey   = labelDomain + "/download-job"
	serverServiceLabelKey = labelDomain + "/server-service"
	serverPVCLabelKey     = labelDomain + "/server-pvc"
	serverPodLabelKey     = labelDomain + "/server-pod"
	serverNameLabelKey    = labelDomain + "/server-name"
)

var valheimServerGVK = valheimv1beta1.GroupVersion.WithKind("ValheimServer")

// ValheimServerReconciler reconciles a ValheimServer object
type ValheimServerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=valheim.zingerweb.services,resources=valheimservers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=valheim.zingerweb.services,resources=valheimservers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=valheim.zingerweb.services,resources=valheimservers/finalizers,verbs=update

//+kubebuilder:rbac:groups="",resources=pods;services;persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods/status;services/status;persistentvolumeclaims/status,verbs=get

//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get

//+kubebuilder:rbac:groups="",resources=secrets,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *ValheimServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("valheimserver", req.NamespacedName)
	ctx = logr.NewContext(ctx, log)

	log.V(1).Info("getting ValheimServer")
	var vhs valheimv1beta1.ValheimServer
	if err := r.Get(ctx, req.NamespacedName, &vhs); err != nil {
		if apierrors.IsNotFound(err) {
			log.V(1).Info("ValheimServer not found")
			// TODO: Clean up AWS resources.
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch ValheimServer")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("ensuring service")
	if err := r.ensureService(ctx, &vhs); err != nil {
		log.Error(err, "unable to ensure service")
		return ctrl.Result{}, err
	}

	log.V(1).Info("ensuring pvc")
	pvc, err := r.ensurePVC(ctx, &vhs)
	if err != nil {
		log.Error(err, "unable to ensure pvc")
		return ctrl.Result{}, err
	}

	var awsSecret corev1.Secret
	if err := r.Get(ctx, types.NamespacedName{Namespace: vhs.Namespace, Name: vhs.Spec.AWSSecretName}, &awsSecret); err != nil {
		log.Error(err, "unable to get AWS secret")
		return ctrl.Result{}, err
	}

	ec2.NewFromConfig(aws.Config{Region: vhs.Spec.AWSRegion})
	switch vhs.Spec.Location {
	case valheimv1beta1.ValheimServerLocationK8s:
		log.V(1).Info("ensuring pod")
		if err := r.ensurePod(ctx, &vhs, pvc); err != nil {
			log.Error(err, "unable to ensure pod")
			return ctrl.Result{}, err
		}
	case valheimv1beta1.ValheimServerLocationAWS:
	case valheimv1beta1.ValheimServerLocationNone:

	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ValheimServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&valheimv1beta1.ValheimServer{}).
		Owns(&corev1.Pod{}).
		Owns(&batchv1.Job{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Complete(r)
}
