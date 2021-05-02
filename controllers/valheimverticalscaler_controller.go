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

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	valheimv1beta1 "github.com/armsnyder/valheim-server/api/v1beta1"
)

// ValheimVerticalScalerReconciler reconciles a ValheimVerticalScaler object
type ValheimVerticalScalerReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=valheim.zingerweb.services,resources=valheimverticalscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=valheim.zingerweb.services,resources=valheimverticalscalers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=valheim.zingerweb.services,resources=valheimverticalscalers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ValheimVerticalScaler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *ValheimVerticalScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("valheimverticalscaler", req.NamespacedName)

	// First, look up the ValheimVerticalScaler that the incoming event is about.
	var vvs valheimv1beta1.ValheimVerticalScaler
	if err := r.Get(ctx, req.NamespacedName, &vvs); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, "unable to get ValheimVerticalScaler")
		return ctrl.Result{}, err
	}

	// Then look up the referenced Deployment.
	var deployment appsv1.Deployment
	if err := r.Get(ctx, types.NamespacedName{Name: vvs.Spec.K8sDeployment.Name, Namespace: vvs.Namespace}, &deployment); err != nil {
		r.Log.Error(err, "unable to get referenced Deployment")
		r.recorder.Event(&vvs, corev1.EventTypeWarning, "GetDeployment", err.Error())
		// Ignore NotFound errors since they will not resolve on their own.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.V(1).Info("found referenced deployment")

	// TODO: Add logic for scaling or not, depending on the state of the ValheimVerticalScaler and Deployment.

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ValheimVerticalScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor("ValheimVerticalScaler")
	return ctrl.NewControllerManagedBy(mgr).
		For(&valheimv1beta1.ValheimVerticalScaler{}).
		Complete(r)
}
