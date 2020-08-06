package deploy

import (
	"fmt"
	installv1alpha1 "github.com/bianchi2/bamboo-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func GetPostgresDeployment(bamboo *installv1alpha1.Bamboo, bambooAPI BambooAPI) *appsv1.Deployment {

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgres",
			Namespace: bamboo.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"k8s-app": "postgres",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"k8s-app": "postgres",
					},
				},
				Spec: apiv1.PodSpec{
					Volumes: []apiv1.Volume{
						{
							Name: "postgres-data",
							VolumeSource: apiv1.VolumeSource{
								PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
									ClaimName: "postgres-data",
								},
							},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:  "postgres",
							Image: "postgres:9.6-alpine",
							Args:  []string{"-c", "max_connections=1000"},
							Ports: []apiv1.ContainerPort{
								{
									Name:          "postgres",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 5432,
								},
							},
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceMemory: resource.MustParse("265Mi"),
									apiv1.ResourceCPU:    resource.MustParse("100m"),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceMemory: resource.MustParse("2Gi"),
									apiv1.ResourceCPU:    resource.MustParse("1"),
								},
							},
							Env: []apiv1.EnvVar{

								{
									Name:  "POSTGRES_DB",
									Value: bamboo.Spec.Datasource.Database,
								},
								{
									Name:  "POSTGRES_USER",
									Value: bamboo.Spec.Datasource.Username,
								},
								{
									Name:  "POSTGRES_PASSWORD",
									Value: bamboo.Spec.Datasource.Password,
								},
								{
									Name:  "PGDATA",
									Value: "/var/lib/postgresql/data/pgdata",
								},
							},
							ReadinessProbe: &apiv1.Probe{
								Handler: apiv1.Handler{
									Exec: &apiv1.ExecAction{
										Command: []string{
											"/bin/sh",
											"-c",
											"psql -h 127.0.0.1 -U $POSTGRES_USER -d $POSTGRES_DB -c 'SELECT 1'",
										},
									},
								},
								InitialDelaySeconds: 10,
								FailureThreshold:    10,
								SuccessThreshold:    1,
								PeriodSeconds:       5,
								TimeoutSeconds:      5,
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "postgres-data",
									MountPath: "/var/lib/postgresql/data",
								},
							},
						},
					},
				},
			},
		},
	}
	err := controllerutil.SetControllerReference(bamboo, deployment, bambooAPI.Scheme)
	if err != nil {
		fmt.Printf("An error occurred when setting controller reference: %s", err)
	}
	return deployment
}
