package controllers

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	valheimv1beta1 "github.com/armsnyder/valheim-server/api/v1beta1"
)

func (r *ValheimServerReconciler) ensurePVC(ctx context.Context, vhs *valheimv1beta1.ValheimServer) (metav1.Object, error) {
	log := logf.FromContext(ctx).WithName("ensurePVC")

	log.V(1).Info("checking for existing pvcs")
	var pvcList corev1.PersistentVolumeClaimList
	if err := r.List(ctx, &pvcList, client.InNamespace(vhs.Namespace), client.MatchingLabels(r.pvcMatchLabels(vhs))); err != nil {
		return nil, err
	}

	if len(pvcList.Items) == 1 {
		log.V(1).Info("found one pvc")
		return &pvcList.Items[0], nil
	}

	if len(pvcList.Items) > 1 {
		log.V(1).Info("multiple pvcs found. cleaning up existing pvcs.")
		if err := r.cleanUpPVC(ctx, vhs); err != nil {
			return nil, err
		}
		return nil, errors.New("multiple existing pvcs found and cleaned up")
	}

	log.V(1).Info("creating pvc for server")
	return r.createPVC(ctx, vhs)
}

func (r *ValheimServerReconciler) createPVC(ctx context.Context, vhs *valheimv1beta1.ValheimServer) (metav1.Object, error) {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName:    vhs.Name + "-",
			Namespace:       vhs.Namespace,
			Labels:          r.pvcMatchLabels(vhs),
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(vhs, valheimServerGVK)},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceStorage: resource.MustParse("2Gi"),
				},
			},
		},
	}
	return pvc, r.Create(ctx, pvc)
}

func (r *ValheimServerReconciler) cleanUpPVC(ctx context.Context, vhs *valheimv1beta1.ValheimServer) error {
	return r.DeleteAllOf(ctx, &corev1.PersistentVolumeClaim{}, client.InNamespace(vhs.Namespace), client.MatchingLabels(r.pvcMatchLabels(vhs)))
}

func (r *ValheimServerReconciler) pvcMatchLabels(vhs *valheimv1beta1.ValheimServer) map[string]string {
	return map[string]string{
		serverNameLabelKey: vhs.Name,
		serverPVCLabelKey:  trueLabelValue,
	}
}
