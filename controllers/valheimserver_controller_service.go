package controllers

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	valheimv1beta1 "github.com/armsnyder/valheim-server/api/v1beta1"
)

func (r *ValheimServerReconciler) ensureService(ctx context.Context, vhs *valheimv1beta1.ValheimServer) error {
	log := logf.FromContext(ctx).WithName("ensureService")

	log.V(1).Info("checking for existing services")
	var serviceList corev1.ServiceList
	if err := r.List(ctx, &serviceList, client.InNamespace(vhs.Namespace), client.MatchingLabels(r.serviceMatchLabels(vhs))); err != nil {
		return err
	}

	if len(serviceList.Items) == 1 {
		log.V(1).Info("found one service")
		return nil
	}

	if len(serviceList.Items) > 1 {
		log.V(1).Info("multiple services found. cleaning up existing services.")
		if err := r.cleanUpService(ctx, vhs); err != nil {
			return err
		}
		return errors.New("multiple existing services found and cleaned up")
	}

	log.V(1).Info("creating service for server")
	return r.createService(ctx, vhs)
}

func (r *ValheimServerReconciler) createService(ctx context.Context, vhs *valheimv1beta1.ValheimServer) error {
	return r.Create(ctx, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName:    vhs.Name + "-",
			Namespace:       vhs.Namespace,
			Labels:          r.serviceMatchLabels(vhs),
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(vhs, valheimServerGVK)},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				serverNameLabelKey: vhs.Name,
				serverPodLabelKey:  trueLabelValue,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "game1",
					Port:       vhs.Spec.Port,
					TargetPort: intstr.FromInt(2456),
					Protocol:   corev1.ProtocolUDP,
				},
				{
					Name:       "game2",
					Port:       vhs.Spec.Port + 1,
					TargetPort: intstr.FromInt(2457),
					Protocol:   corev1.ProtocolUDP,
				},
				{
					Name:       "game3",
					Port:       vhs.Spec.Port + 2,
					TargetPort: intstr.FromInt(2458),
					Protocol:   corev1.ProtocolUDP,
				},
			},
			Type: corev1.ServiceTypeLoadBalancer,
		},
	})
}

func (r *ValheimServerReconciler) cleanUpService(ctx context.Context, vhs *valheimv1beta1.ValheimServer) error {
	return r.DeleteAllOf(ctx, &corev1.Service{}, client.InNamespace(vhs.Namespace), client.MatchingLabels(r.serviceMatchLabels(vhs)))
}

func (r *ValheimServerReconciler) serviceMatchLabels(vhs *valheimv1beta1.ValheimServer) map[string]string {
	return map[string]string{
		serverNameLabelKey:    vhs.Name,
		serverServiceLabelKey: trueLabelValue,
	}
}
