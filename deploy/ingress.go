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
			Annotations: map[string]string{
				"kubernetes.io/ingress.class":              "nginx",
				"nginx.ingress.kubernetes.io/ssl-redirect": "true",
				"cert-manager.io/cluster-issuer":           "letsencrypt-prod",
			},
		},
		Spec: v1beta1.IngressSpec{
			TLS: []v1beta1.IngressTLS{
				{
					Hosts:      []string{bamboo.Spec.Ingress.Host},
					SecretName: "bamboo-tls",
				},
			},
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
