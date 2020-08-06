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

func GetBambooAgentStatefulSet(bamboo *installv1alpha1.Bamboo, bambooAPI BambooAPI) *appsv1.StatefulSet {
	var privileged bool = true
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bamboo.Name + "-agent",
			Namespace: bamboo.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: "bamboo-agent",
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"k8s-app": bamboo.Name + "-agent",
				},
			},
			Replicas: &bamboo.Spec.RemoteAgents.Replicas,
			VolumeClaimTemplates: []apiv1.PersistentVolumeClaim{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "PersistentVolumeClaim",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      bamboo.Name + "-agent-data",
						Namespace: bamboo.Namespace,
					},
					Spec: apiv1.PersistentVolumeClaimSpec{
						AccessModes: []apiv1.PersistentVolumeAccessMode{
							apiv1.ReadWriteOnce,
						},
						Resources: apiv1.ResourceRequirements{
							Requests: apiv1.ResourceList{
								apiv1.ResourceStorage: resource.MustParse("5Gi"),
							},
						},
					},
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"k8s-app": bamboo.Name + "-agent",
					},
				},
				Spec: apiv1.PodSpec{
					Volumes: []apiv1.Volume{
						{
							Name: bamboo.Name + "-agent-data",
							VolumeSource: apiv1.VolumeSource{
								PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
									ClaimName: bamboo.Name + "-agent-data",
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
							Env: []apiv1.EnvVar{
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
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      bamboo.Name + "-agent-data",
									MountPath: "/var/atlassian/application-data/bamboo-agent",
								},
							},
						},
						{
							Name:  "dind",
							Image: "docker:18-dind",
							SecurityContext: &apiv1.SecurityContext{
								Privileged: &privileged,
							},
						},
					},
				},
			},
		},
	}
	err := controllerutil.SetControllerReference(bamboo, statefulSet, bambooAPI.Scheme)
	if err != nil {
		fmt.Printf("An error occurred when setting controller reference: %s", err)
	}
	return statefulSet
}
