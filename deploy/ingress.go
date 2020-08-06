package deploy

import (
	"fmt"
	installv1alpha1 "github.com/bianchi2/bamboo-operator/api/v1alpha1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func GetBambooIngress(bamboo *installv1alpha1.Bamboo, bambooAPI BambooAPI) *v1beta1.Ingress {
	annotations := bamboo.Spec.Ingress.Annotations
	var ingressTLS []v1beta1.IngressTLS
	if bamboo.Spec.Ingress.Tls {
		ingressTLS = []v1beta1.IngressTLS{
			{
				Hosts:      []string{bamboo.Spec.Ingress.Host},
				SecretName: bamboo.Spec.Ingress.TlsSecretName,
			},
		}
	}
	ingress := &v1beta1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: v1beta1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      bamboo.Name,
			Namespace: bamboo.Namespace,
			Labels: map[string]string{
				"k8s-app": bamboo.Name,
			},
			Annotations: annotations,
		},
		Spec: v1beta1.IngressSpec{
			TLS: ingressTLS,
			Rules: []v1beta1.IngressRule{
				{
					Host: bamboo.Spec.Ingress.Host,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{

							Paths: []v1beta1.HTTPIngressPath{
								{
									Backend: v1beta1.IngressBackend{
										ServiceName: bamboo.Name,
										ServicePort: intstr.FromInt(8085),
									},
									Path: "/",
								},
							},
						},
					},
				},
			},
		},
	}
	err := controllerutil.SetControllerReference(bamboo, ingress, bambooAPI.Scheme)
	if err != nil {
		fmt.Printf("An error occurred when setting controller reference: %s", err)
	}
	return ingress
}
