package deploy

import (
	"fmt"
	installv1alpha1 "github.com/bianchi2/bamboo-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func GetBambooDeployment(bamboo *installv1alpha1.Bamboo, bambooAPI BambooAPI) *appsv1.Deployment {
	proxyPort := "80"
	proxyScheme := "http"
	if bamboo.Spec.Ingress.Tls {
		proxyPort = "443"
		proxyScheme = "https"
	}

	mode := int32(0777)
	replicas := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bamboo.Name,
			Namespace: bamboo.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"k8s-app": bamboo.Name,
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"k8s-app": bamboo.Name,
					},
				},
				Spec: apiv1.PodSpec{
					RestartPolicy: "Always",
					Volumes: []apiv1.Volume{
						{
							Name: "create-config-sh",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									DefaultMode: &mode,
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: "create-config-sh",
									},
								},
							},
						},
						{
							Name: "bamboo-cfg-xml",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: "bamboo-cfg-xml",
									},
								},
							},
						},
						{
							Name: "bamboo-administration-xml",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: "bamboo-administration-xml",
									},
								},
							},
						},
						{
							Name: "logging-properties",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: "bamboo-logging-properties",
									},
								},
							},
						},
						{
							Name: "bamboo-data",
							VolumeSource: apiv1.VolumeSource{
								PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
									ClaimName: "bamboo-data",
								},
							},
						},
					},
					InitContainers: []apiv1.Container{
						{
							Name:  "wait-for-db",
							Image: "postgres:9.6-alpine",
							Command: []string{
								"/bin/sh",
							},
							Args: []string{"-c", "until pg_isready -h " + bamboo.Spec.Datasource.Host + " -p " + bamboo.Spec.Datasource.Port + "; do echo waiting for database; sleep 2; done;"},
						},
						{
							Name:  "create-config",
							Image: bamboo.Spec.ImageRepo + ":" + bamboo.Spec.ImageTag,
							Command: []string{
								"/bin/sh",
							},
							Args: []string{"-c", "/tmp/create-config.sh"},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "create-config-sh",
									MountPath: "/tmp/create-config.sh",
									SubPath:   "create-config.sh",
								},
								{
									Name:      "bamboo-cfg-xml",
									MountPath: "/tmp/bamboo.cfg.xml",
									SubPath:   "bamboo.cfg.xml",
								},
								{
									Name:      "bamboo-administration-xml",
									MountPath: "/tmp/administration.xml",
									SubPath:   "administration.xml",
								},
								{
									Name:      "bamboo-data",
									MountPath: "/var/atlassian/application-data/bamboo",
								},
							},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:            bamboo.Name,
							Image:           bamboo.Spec.ImageRepo + ":" + bamboo.Spec.ImageTag,
							ImagePullPolicy: "IfNotPresent",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 8085,
								},
								{
									Name:          "debug",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 7896,
								},
								{
									Name:          "jms",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 54663,
								},
							},
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceMemory: resource.MustParse(bamboo.Spec.ContainerMemRequest),
									apiv1.ResourceCPU:    resource.MustParse(bamboo.Spec.ContainerCPURequest),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceMemory: resource.MustParse(bamboo.Spec.ContainerMemLimit),
									apiv1.ResourceCPU:    resource.MustParse(bamboo.Spec.ContainerCPULimit),
								},
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "JVM_MINIMUM_MEMORY",
									Value: bamboo.Spec.JvmMinimumMemory,
								},
								{
									Name:  "JVM_MAXIMUM_MEMORY",
									Value: bamboo.Spec.JvmMaximumMemory,
								},
								{
									Name:  "ATL_PROXY_NAME",
									Value: bamboo.Spec.Ingress.Host,
								},
								{
									Name:  "ATL_PROXY_PORT",
									Value: proxyPort,
								},
								{
									Name:  "ATL_TOMCAT_SCHEME",
									Value: proxyScheme,
								},
								{
									Name:  "JVM_SUPPORT_RECOMMENDED_ARGS",
									Value: bamboo.Spec.JvmSupportRecommendedArgs,
								},
							},
							ReadinessProbe: &apiv1.Probe{
								InitialDelaySeconds: 20,
								FailureThreshold:    10,
								PeriodSeconds:       20,
								TimeoutSeconds:      20,
								Handler: apiv1.Handler{
									HTTPGet: &apiv1.HTTPGetAction{
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: int32(8085),
										},
										Path: "/rest/api/latest/status",
									},
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "bamboo-data",
									MountPath: "/var/atlassian/application-data/bamboo",
								},
								{
									Name:      "logging-properties",
									MountPath: "/opt/atlassian/bamboo/atlassian-bamboo/WEB-INF/classes/log4j.properties",
									SubPath:   "logging.properties",
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
