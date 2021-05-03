package controllers

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	valheimv1beta1 "github.com/armsnyder/valheim-vertical-scaler/api/v1beta1"
)

var _ = Describe("ValheimVerticalScaler controller", func() {
	ctx := context.Background()

	var namespace corev1.Namespace

	BeforeEach(func() {
		By("creating a namespace")
		namespace = corev1.Namespace{ObjectMeta: metav1.ObjectMeta{GenerateName: "valheim-server-ns-"}}
		Expect(k8sClient.Create(ctx, &namespace)).Should(Succeed())
	})

	Context("all requisite objects exist", func() {
		const deploymentName = "valheim-server-deploy"
		const vvsName = "valheim-server-vvs"

		deployment := func() *appsv1.Deployment {
			return &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deploymentName,
					Namespace: namespace.Name,
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "valheim-server",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": "valheim-server",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "valheim-server",
									Image: "lloesche/valheim-server",
								},
							},
						},
					},
				},
			}
		}

		defaultVVS := func() *valheimv1beta1.ValheimVerticalScaler {
			return &valheimv1beta1.ValheimVerticalScaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      vvsName,
					Namespace: namespace.Name,
				},
				Spec: valheimv1beta1.ValheimVerticalScalerSpec{
					K8sDeployment: valheimv1beta1.K8sDeployment{
						Name: deploymentName,
					},
					AWS: valheimv1beta1.AWS{
						Region:               "us-west-2",
						Domain:               "valheim-server-aws-creds",
						CredentialSecretName: "valheim.example.com",
						PrivateKeySecretName: "valheim-server-aws-ssh",
						InstanceID:           "i-1234567890abcdef0",
					},
				},
			}
		}

		BeforeEach(func() {
			By("creating a Valheim server deployment")
			Expect(k8sClient.Create(ctx, deployment())).Should(Succeed())
		})

		It("should create successfully", func() {
			By("creating ValheimVerticalScaler")
			Expect(k8sClient.Create(ctx, defaultVVS())).Should(Succeed())

			By("expecting Ready phase")
			Eventually(func() valheimv1beta1.Phase {
				var vss valheimv1beta1.ValheimVerticalScaler
				_ = k8sClient.Get(ctx, types.NamespacedName{Name: vvsName, Namespace: namespace.Name}, &vss)
				return vss.Status.Phase
			}).Should(Equal(valheimv1beta1.PhaseReady))
		})

		It("should error when referencing nonexistent Deployment", func() {
			By("creating ValheimVerticalScaler")
			vvs := defaultVVS()
			vvs.Spec.K8sDeployment.Name = "i-do-not-exist"
			Expect(k8sClient.Create(ctx, vvs)).Should(Succeed())

			By("expecting Error phase")
			Eventually(func() valheimv1beta1.ValheimVerticalScalerStatus {
				var vss valheimv1beta1.ValheimVerticalScaler
				_ = k8sClient.Get(ctx, types.NamespacedName{Name: vvsName, Namespace: namespace.Name}, &vss)
				return vss.Status
			}).Should(MatchFields(IgnoreExtras, Fields{
				"Phase": Equal(valheimv1beta1.PhaseError),
				"Error": ContainSubstring("i-do-not-exist"),
			}))

			By("expecting warning-type event")
			Eventually(func() []corev1.Event {
				var eventList corev1.EventList
				_ = k8sClient.List(ctx, &eventList, client.MatchingFields{
					"involvedObject.name":      vvsName,
					"involvedObject.namespace": namespace.Name,
				})
				return eventList.Items
			}).Should(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Reason": Equal("Deployment"),
				"Type":   Equal(corev1.EventTypeWarning),
			})))
		})
	})
})
