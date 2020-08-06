package deploy

import (
	installv1alpha1 "github.com/bianchi2/bamboo-operator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
)

func GetInstallBambooJob(bamboo *installv1alpha1.Bamboo) *batchv1.Job {
	license := bamboo.Spec.Installer.License
	re := regexp.MustCompile(` +\r?\n +`)
	license = re.ReplaceAllString(license, " ")

	mode := int32(0777)
	backoffLimit := int32(1)
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "install-" + bamboo.Name,
			Namespace: bamboo.Namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		Spec: batchv1.JobSpec{
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"k8s-app": "install-" + bamboo.Name,
					},
				},
				Spec: apiv1.PodSpec{
					RestartPolicy: "Never",
					Volumes: []apiv1.Volume{
						{
							Name: "install-bamboo",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									DefaultMode: &mode,
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: "install-bamboo-py",
									},
								},
							},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:            bamboo.Name + "-job-container",
							Image:           "yivantsov/py-init-atl",
							ImagePullPolicy: "Always",
							Command:         []string{"python3"},
							Args:            []string{"-u", "/root/install-bamboo.py"},
							Env: []apiv1.EnvVar{
								{
									Name:  "PYTHONUNBUFFERED",
									Value: "0",
								},
								{
									Name:  "PROTOCOL",
									Value: "http",
								},
								{
									Name:  "BAMBOO_ENDPOINT",
									Value: bamboo.Name + ":8085",
								},
								{
									Name:  "BAMBOO_LICENSE",
									Value: license,
								},
								{
									Name:  "ADMIN_USER",
									Value: bamboo.Spec.Installer.AdminName,
								},
								{
									Name:  "ADMIN_PASSWORD",
									Value: bamboo.Spec.Installer.AdminPassword,
								},
								{
									Name:  "ADMIN_EMAIL",
									Value: bamboo.Spec.Installer.AdminEmail,
								},
								{
									Name:  "FULL_NAME",
									Value: bamboo.Spec.Installer.AdminFullName,
								},
								{
									Name:  "BAMBOO_DATABASE_HOST",
									Value: bamboo.Spec.Datasource.Host,
								},
								{
									Name:  "BAMBOO_DATABASE_PORT",
									Value: bamboo.Spec.Datasource.Port,
								},
								{
									Name:  "BAMBOO_DATABASE_NAME",
									Value: bamboo.Spec.Datasource.Database,
								},
								{
									Name:  "BAMBOO_DATABASE_USER",
									Value: bamboo.Spec.Datasource.Username,
								},
								{
									Name:  "BAMBOO_DATABASE_PASSWORD",
									Value: bamboo.Spec.Datasource.Password,
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "install-bamboo",
									MountPath: "/root/install-bamboo.py",
									SubPath:   "install-bamboo.py",
								},
							},
						},
					},
				},
			},
			BackoffLimit: &backoffLimit,
		},
	}

	return job
}
