package deploy

import (
	installv1alpha1 "github.com/bianchi2/bamboo-operator/api/v1alpha1"
	"github.com/bianchi2/bamboo-operator/k8s"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func GetBackupPod(bamboo *installv1alpha1.Bamboo, version string) *apiv1.Pod {
	currentTime := time.Now()
	backupPod := &apiv1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bamboo-backup-" + currentTime.Format("2020-01-02") + "-" + version + "-" + k8s.GeneratePasswd(4),
			Namespace: bamboo.Namespace,
			Labels: map[string]string{
				"k8s-app": bamboo.Name + "-backup",
			},
		},
		Spec: apiv1.PodSpec{
			RestartPolicy: "Never",
			Volumes: []apiv1.Volume{
				{
					Name: "bamboo-data",
					VolumeSource: apiv1.VolumeSource{
						PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
							ClaimName: "bamboo-data",
						},
					},
				},
			},
			Containers: []apiv1.Container{
				{
					Name:            bamboo.Name,
					Image:           bamboo.Spec.ImageRepo + ":" + bamboo.Spec.ImageTag,
					Command:         []string{"/bin/bash"},
					Args:            []string{"-c", "mkdir -p /var/atlassian/application-data/bamboo/backups; tar --exclude='/var/atlassian/application-data/bamboo/backups' -zcvf /var/atlassian/application-data/$(date +\"%H-%M_%d-%m-%y\")_version_" + version + ".tar.gz /var/atlassian/application-data/bamboo --owner=bamboo --group=bamboo; cp /var/atlassian/application-data/*.tar.gz /var/atlassian/application-data/bamboo/backups"},
					ImagePullPolicy: "IfNotPresent",
					VolumeMounts: []apiv1.VolumeMount{
						{
							Name:      "bamboo-data",
							MountPath: "/var/atlassian/application-data/bamboo",
						},
					},
				},
			},
		},
	}
	return backupPod
}
