package deploy

import (
	"fmt"
	installv1alpha1 "github.com/bianchi2/bamboo-operator/api/v1alpha1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type BambooAPI struct {
	Client client.Client
	Scheme *runtime.Scheme
}

func GetBambooService(bamboo *installv1alpha1.Bamboo, bambooAPI BambooAPI) *apiv1.Service {
	service := &apiv1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      bamboo.Name,
			Namespace: bamboo.Namespace,
			Labels: map[string]string{
				"k8s-app": bamboo.Name,
			},
		},
		Spec: apiv1.ServiceSpec{
			Type: "NodePort",
			Ports: []apiv1.ServicePort{
				{
					Name: "http",
					Port: 8085,
				},
				{
					Name: "debug",
					Port: 7896,
				},
				{
					Name: "jms",
					Port: 54663,
				},
			},
			Selector: map[string]string{
				"k8s-app": bamboo.Name,
			},
		},
	}
	err := controllerutil.SetControllerReference(bamboo, service, bambooAPI.Scheme)
	if err != nil {
		fmt.Printf("An error occurred when setting controller reference: %s", err)
	}
	return service
}

func GetPostgresService(bamboo *installv1alpha1.Bamboo, bambooAPI BambooAPI) *apiv1.Service {
	service := &apiv1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgres",
			Namespace: bamboo.Namespace,
			Labels: map[string]string{
				"k8s-app": "postgres",
			},
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name:     "http",
					Port:     5432,
					Protocol: "TCP",
				},
			},
			Selector: map[string]string{
				"k8s-app": "postgres",
			},
		},
	}

	err := controllerutil.SetControllerReference(bamboo, service, bambooAPI.Scheme)
	if err != nil {
		fmt.Printf("An error occurred when setting controller reference: %s", err)
	}
	return service
}
