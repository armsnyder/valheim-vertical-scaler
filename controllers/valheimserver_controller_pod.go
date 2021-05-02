package controllers

import (
	"context"
	"errors"
	"path"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	valheimv1beta1 "github.com/armsnyder/valheim-server/api/v1beta1"
)

func (r *ValheimServerReconciler) ensurePod(ctx context.Context, vhs *valheimv1beta1.ValheimServer, pvc metav1.Object) error {
	log := logf.FromContext(ctx).WithName("ensurePod")

	log.V(1).Info("checking for existing pods")
	var podList corev1.PodList
	if err := r.List(ctx, &podList, client.InNamespace(vhs.Namespace), client.MatchingLabels(r.podMatchLabels(vhs))); err != nil {
		return err
	}

	if len(podList.Items) == 1 {
		log.V(1).Info("found one pod")
		return nil
	}

	if len(podList.Items) > 1 {
		log.V(1).Info("multiple pods found. cleaning up existing pods.")
		if err := r.cleanUpPod(ctx, vhs); err != nil {
			return err
		}
		return errors.New("multiple existing pods found and cleaned up")
	}

	log.V(1).Info("creating pod for server")
	return r.createPod(ctx, vhs, pvc)
}

func (r *ValheimServerReconciler) createPod(ctx context.Context, vhs *valheimv1beta1.ValheimServer, pvc metav1.Object) error {
	pod := r.defaultPod(vhs, pvc)

	if vhs.Spec.Backup != (valheimv1beta1.ValheimServerBackup{}) {
		pod.Spec.Containers[0].Env = append(
			pod.Spec.Containers[0].Env,
			corev1.EnvVar{
				Name:  "BACKUPS",
				Value: "true",
			},
			corev1.EnvVar{
				Name:  "BACKUPS_CRON",
				Value: "0 * * * *",
			},
			corev1.EnvVar{
				Name:  "BACKUPS_MAX_AGE",
				Value: "3",
			},
			corev1.EnvVar{
				Name:  "POST_BACKUP_HOOK",
				Value: "timeout 300 scp -i /extraVolumes/backup-ssh-key/ssh-privatekey -o StrictHostKeyChecking=no @BACKUP_FILE@ " + path.Join(vhs.Spec.Backup.Target, "$(basename @BACKUP_FILE@)"),
			},
		)
		pod.Spec.Containers[0].VolumeMounts = append(pod.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      "backup-ssh-key",
			ReadOnly:  true,
			MountPath: "/extraVolumes/backup-ssh-key",
		})
		pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
			Name: "backup-ssh-key",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: vhs.Spec.Backup.PrivateKeySecretName,
				},
			},
		})
	}

	return r.Create(ctx, pod)
}

func (r *ValheimServerReconciler) defaultPod(vhs *valheimv1beta1.ValheimServer, pvc metav1.Object) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName:    vhs.Name + "-",
			Namespace:       vhs.Namespace,
			Labels:          r.podMatchLabels(vhs),
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(vhs, valheimServerGVK)},
		},
		Spec: corev1.PodSpec{
			NodeSelector: map[string]string{
				"kubernetes.io/arch": "amd64",
				"kubernetes.io/os":   "linux",
			},
			Volumes: []corev1.Volume{
				{
					Name: "gamefiles",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvc.GetName(),
						},
					},
				},
			},
			Containers: []corev1.Container{
				{
					Name:            "valheim-server",
					Image:           "lloesche/valheim-server:latest",
					ImagePullPolicy: corev1.PullAlways,
					Env: []corev1.EnvVar{
						{
							Name:  "SERVER_NAME",
							Value: vhs.Spec.ServerName,
						},
						{
							Name:  "WORLD_NAME",
							Value: vhs.Spec.WorldName,
						},
						{
							Name: "SERVER_PASS",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: vhs.Spec.PasswordSecretName,
									},
									Key: "serverPassword",
								},
							},
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							MountPath: "/config",
							Name:      "gamefiles",
						},
					},
				},
			},
		},
	}
}

func (r *ValheimServerReconciler) cleanUpPod(ctx context.Context, vhs *valheimv1beta1.ValheimServer) error {
	return r.DeleteAllOf(ctx, &corev1.Pod{}, client.InNamespace(vhs.Namespace), client.MatchingLabels(r.podMatchLabels(vhs)))
}

func (r *ValheimServerReconciler) podMatchLabels(vhs *valheimv1beta1.ValheimServer) map[string]string {
	return map[string]string{
		serverNameLabelKey: vhs.Name,
		serverPodLabelKey:  trueLabelValue,
	}
}
