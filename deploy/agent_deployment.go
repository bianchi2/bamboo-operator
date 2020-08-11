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

func GetBambooAgentDeployment(bamboo *installv1alpha1.Bamboo, bambooAPI BambooAPI, id string, uid string) *appsv1.Deployment {
	mode := int32(0777)
	agentEnv := []apiv1.EnvVar{
		{
			Name:  "WRAPPER_JAVA_INITMEMORY",
			Value: bamboo.Spec.RemoteAgents.WrapperJavaInitMemory,
		},
		{
			Name:  "WRAPPER_JAVA_MAXMEMORY",
			Value: bamboo.Spec.RemoteAgents.WrapperJavaMaxMemory,
		},
		{
			Name:  "BAMBOO_SERVER",
			Value: "http://" + bamboo.Name + ":8085/agentServer/",
		},
		{
			Name:  "DOCKER_HOST",
			Value: "tcp://127.0.0.1:2375",
		},
	}
	// add security token if set
	if len(bamboo.Spec.RemoteAgents.SecurityToken) > 0 {
		agentEnv = append(agentEnv, apiv1.EnvVar{Name: "SECURITY_TOKEN", Value: bamboo.Spec.RemoteAgents.SecurityToken})
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "remote-agent-" + id,
			Namespace: bamboo.Namespace,
			Labels: map[string]string{
				"k8s-app": bamboo.Name + "-agent",
				"id":      "remote-agent-" + id,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			}, Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"k8s-app": bamboo.Name + "-agent",
					"id":      "remote-agent-" + id,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"k8s-app": bamboo.Name + "-agent",
						"id":      "remote-agent-" + id,
					},
				},
				Spec: apiv1.PodSpec{
					InitContainers: []apiv1.Container{
						{
							Name:  "create-config",
							Image: bamboo.Spec.ImageRepo + ":" + bamboo.Spec.ImageTag,
							Command: []string{
								"/bin/sh",
							},
							Args: []string{"-c", "/tmp/create-agent-config.sh"},
							Env: []apiv1.EnvVar{
								{
									Name:  "ID",
									Value: id,
								},
								{
									Name:  "UID",
									Value: uid,
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "create-agent-config-sh",
									MountPath: "/tmp/create-agent-config.sh",
									SubPath:   "create-agent-config.sh",
								},
								{
									Name:      "bamboo-agent-cfg-xml",
									MountPath: "/tmp/bamboo-agent.cfg.xml",
									SubPath:   "bamboo-agent.cfg.xml",
								},
								{
									Name:      "bamboo-agent-data",
									MountPath: "/var/atlassian/application-data/bamboo-agent",
								},
							},
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "bamboo-agent-data",
							VolumeSource: apiv1.VolumeSource{
								PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
									ClaimName: "remote-agent-" + id,
								},
							},
						},
						{
							Name: "create-agent-config-sh",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									DefaultMode: &mode,
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: "create-agent-config-sh",
									},
								},
							},
						},
						{
							Name: "bamboo-agent-cfg-xml",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: "bamboo-agent-cfg-xml",
									},
								},
							},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:  bamboo.Name + "-agent",
							Image: bamboo.Spec.RemoteAgents.ImageRepo + ":" + bamboo.Spec.RemoteAgents.ImageTag,
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceMemory: resource.MustParse(bamboo.Spec.RemoteAgents.ContainerMemRequest),
									apiv1.ResourceCPU:    resource.MustParse(bamboo.Spec.RemoteAgents.ContainerCPURequest),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceMemory: resource.MustParse(bamboo.Spec.RemoteAgents.ContainerMemLimit),
									apiv1.ResourceCPU:    resource.MustParse(bamboo.Spec.RemoteAgents.ContainerCPULimit),
								},
							},
							Env: agentEnv,
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "bamboo-agent-data",
									MountPath: "/var/atlassian/application-data/bamboo-agent",
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
