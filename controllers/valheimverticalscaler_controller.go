package controllers

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	valheimv1beta1 "github.com/armsnyder/valheim-vertical-scaler/api/v1beta1"
	"github.com/armsnyder/valheim-vertical-scaler/genutil"
)

const deploymentNameIndexField = ".metadata.valheimDeploymentName"

// ValheimVerticalScalerReconciler reconciles a ValheimVerticalScaler object
type ValheimVerticalScalerReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=valheim.zingerweb.services,resources=valheimverticalscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=valheim.zingerweb.services,resources=valheimverticalscalers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *ValheimVerticalScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("valheimverticalscaler", req.NamespacedName)
	ctx = logr.NewContext(ctx, log)

	// First, look up the ValheimVerticalScaler that the incoming event is about.
	var vvs valheimv1beta1.ValheimVerticalScaler
	if err := r.Get(ctx, req.NamespacedName, &vvs); err != nil {
		return genutil.RequeueWithError(ctx, client.IgnoreNotFound(err))
	}

	// Then look up the referenced Deployment.
	var deployment appsv1.Deployment
	if err := r.Get(ctx, types.NamespacedName{Name: vvs.Spec.K8sDeployment.Name, Namespace: vvs.Namespace}, &deployment); err != nil {
		log.Error(err, "unable to get referenced Deployment")
		r.recorder.Event(&vvs, corev1.EventTypeWarning, "Deployment", err.Error())
		// Ignore NotFound errors since they will not resolve on their own.
		if errors.IsNotFound(err) {
			return genutil.RequeueWithError(ctx, r.updateError(ctx, &vvs, err))
		}
		return genutil.RequeueWithError(ctx, err)
	}

	log.V(1).Info("found referenced deployment")

	if err := r.updatePhase(ctx, &vvs, valheimv1beta1.PhaseReady); err != nil {
		return genutil.RequeueWithError(ctx, err)
	}

	// TODO: Add logic for scaling or not, depending on the state of the ValheimVerticalScaler and Deployment.

	return genutil.DoNotRequeue()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ValheimVerticalScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor("ValheimVerticalScaler")

	if err := r.createIndices(mgr); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&valheimv1beta1.ValheimVerticalScaler{}).
		Watches(&source.Kind{Type: &appsv1.Deployment{}}, handler.EnqueueRequestsFromMapFunc(r.mapDeploymentToRequests)).
		Complete(r)
}

func (r *ValheimVerticalScalerReconciler) createIndices(mgr ctrl.Manager) error {
	return mgr.GetFieldIndexer().IndexField(context.Background(), &valheimv1beta1.ValheimVerticalScaler{}, deploymentNameIndexField, func(object client.Object) []string {
		vvs := object.(*valheimv1beta1.ValheimVerticalScaler)

		if vvs.Spec.K8sDeployment.Name == "" {
			return nil
		}

		return []string{vvs.Spec.K8sDeployment.Name}
	})
}

func (r *ValheimVerticalScalerReconciler) mapDeploymentToRequests(object client.Object) []reconcile.Request {
	deployment := object.(*appsv1.Deployment)
	log := r.Log.WithName("mapDeploymentToRequests")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var vvsList valheimv1beta1.ValheimVerticalScalerList

	if err := r.List(ctx, &vvsList,
		client.InNamespace(deployment.Namespace),
		client.MatchingFields{deploymentNameIndexField: deployment.Name}); err != nil {
		log.Error(err, "could not list ValheimVerticalScalers. change to Deployment %s.%s will not be reconciled.",
			deployment.Name, deployment.Namespace)
		return nil
	}

	var requests []reconcile.Request

	for _, vvs := range vvsList.Items {
		if vvs.Spec.K8sDeployment.Name == deployment.Name {
			requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&vvs)})
		}
	}

	return requests
}

func (r *ValheimVerticalScalerReconciler) updatePhase(ctx context.Context, vvs *valheimv1beta1.ValheimVerticalScaler, phase valheimv1beta1.Phase) error {
	return r.updateStatus(ctx, vvs, phase, nil)
}

func (r *ValheimVerticalScalerReconciler) updateError(ctx context.Context, vvs *valheimv1beta1.ValheimVerticalScaler, err error) error {
	return r.updateStatus(ctx, vvs, valheimv1beta1.PhaseError, err)
}

func (r *ValheimVerticalScalerReconciler) updateStatus(ctx context.Context, vvs *valheimv1beta1.ValheimVerticalScaler, phase valheimv1beta1.Phase, err error) error {
	vvs.Status = valheimv1beta1.ValheimVerticalScalerStatus{
		Phase:              phase,
		ObservedGeneration: vvs.Generation,
	}

	if err != nil {
		vvs.Status.Error = err.Error()
	}

	return r.Status().Update(ctx, vvs)
}
