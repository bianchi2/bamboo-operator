/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"github.com/bianchi2/bamboo-operator/deploy"
	"github.com/bianchi2/bamboo-operator/k8s"
	"github.com/bianchi2/bamboo-operator/rest"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"strconv"
	"strings"
	"time"

	//"strconv"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	installv1alpha1 "github.com/bianchi2/bamboo-operator/api/v1alpha1"
)

type BambooAPI struct {
	Client client.Client
	Scheme *runtime.Scheme
}

// BambooReconciler reconciles a Bamboo object
type BambooReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=install.atlassian.com,resources=bambooes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=install.atlassian.com,resources=bambooes/status,verbs=get;update;patch

func (r *BambooReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	bambooAPI := BambooAPI{
		Client: r.Client,
		Scheme: r.Scheme,
	}
	_ = context.Background()
	_ = r.Log.WithValues("bamboo", req.NamespacedName)
	setupLog := ctrl.Log.WithName("bamboo-operator")

	bamboo := &installv1alpha1.Bamboo{}

	err := r.Client.Get(context.TODO(), req.NamespacedName, bamboo)
	if err != nil {
		_ = r.Log.WithValues("failed to get custom resource", req.NamespacedName)
	}

	// deploy Postgres

	postgresPVC := deploy.GetPVC(bamboo, "postgres-data", deploy.BambooAPI(bambooAPI))
	err = deploy.CreatePVC((*deploy.BambooReconciler)(r), postgresPVC, bamboo)
	if err != nil {
		setupLog.Error(err, "unable to create pvc "+postgresPVC.Name)
	}

	postgresSVC := deploy.GetPostgresService(bamboo, deploy.BambooAPI(bambooAPI))
	err = deploy.CreateService((*deploy.BambooReconciler)(r), postgresSVC, bamboo)
	if err != nil {
		setupLog.Error(err, "unable to create service "+postgresSVC.Name)
	}

	postgresDeployment := deploy.GetPostgresDeployment(bamboo, deploy.BambooAPI(bambooAPI))
	err = deploy.CreateDeployment((*deploy.BambooReconciler)(r), postgresDeployment, bamboo)
	if err != nil {
		setupLog.Error(err, "unable to create deployment "+postgresDeployment.Name)
	}

	bambooPvc := deploy.GetPVC(bamboo, "bamboo-data", deploy.BambooAPI(bambooAPI))
	err = deploy.CreatePVC((*deploy.BambooReconciler)(r), bambooPvc, bamboo)
	if err != nil {
		setupLog.Error(err, "unable to create pvc "+bambooPvc.Name)
	}

	// create bamboo configmap
	administrationXmlConfigMap := deploy.GetAdministrationXMLConfigMap(bamboo, deploy.BambooAPI(bambooAPI))
	err = deploy.CreateConfigMap((*deploy.BambooReconciler)(r), administrationXmlConfigMap, bamboo)
	if err != nil {
		setupLog.Error(err, "unable to create configmap "+administrationXmlConfigMap.Name)
	}

	bambooCfgConfigMap := deploy.GetBambooCfgConfigMap(bamboo, deploy.BambooAPI(bambooAPI))
	err = deploy.CreateConfigMap((*deploy.BambooReconciler)(r), bambooCfgConfigMap, bamboo)
	if err != nil {
		setupLog.Error(err, "unable to create configmap "+bambooCfgConfigMap.Name)
	}

	createConfigConfigMap := deploy.GetBambooCreateConfigConfigMap(bamboo, deploy.BambooAPI(bambooAPI))
	err = deploy.CreateConfigMap((*deploy.BambooReconciler)(r), createConfigConfigMap, bamboo)
	if err != nil {
		setupLog.Error(err, "unable to create configmap "+createConfigConfigMap.Name)
	}

	loggingPropertiesConfigMap := deploy.GetLoggingPropertiesConfigMap(bamboo, deploy.BambooAPI(bambooAPI))
	err = deploy.CreateConfigMap((*deploy.BambooReconciler)(r), loggingPropertiesConfigMap, bamboo)
	if err != nil {
		setupLog.Error(err, "unable to create configmap "+loggingPropertiesConfigMap.Name)
	}

	// create Bamboo service

	service := deploy.GetBambooService(bamboo, deploy.BambooAPI(bambooAPI))
	err = deploy.CreateService((*deploy.BambooReconciler)(r), service, bamboo)

	if err != nil {
		setupLog.Error(err, "unable to create Bamboo service")
	}

	// create Bamboo ingress

	ingress := deploy.GetBambooIngress(bamboo, deploy.BambooAPI(bambooAPI))
	err = deploy.CreateIngress((*deploy.BambooReconciler)(r), ingress, bamboo)

	if err != nil {
		setupLog.Error(err, "unable to create Bamboo ingress")
	}

	// create bamboo deployment

	bambooDeployment := deploy.GetBambooDeployment(bamboo, deploy.BambooAPI(bambooAPI))
	err = deploy.CreateDeployment((*deploy.BambooReconciler)(r), bambooDeployment, bamboo)

	if err != nil {
		setupLog.Error(err, "unable to create Bamboo deployment")
	}

	// create Bamboo installer job

	installerConfigmap := deploy.GetInstallBambooConfigMap(bamboo)
	err = deploy.CreateConfigMap((*deploy.BambooReconciler)(r), installerConfigmap, bamboo)

	if err != nil {
		setupLog.Error(err, "unable to create configmap "+installerConfigmap.Name)
	}

	// create Bamboo install job

	installBambooJob := deploy.GetInstallBambooJob(bamboo)
	err = deploy.CreateJob((*deploy.BambooReconciler)(r), installBambooJob, bamboo)

	if err != nil {
		setupLog.Error(err, "unable to create configmap "+installBambooJob.Name)
	}

	// create remote agent deployments

	if bamboo.Spec.RemoteAgents.Enabled {
		setupLog.Info("Remote agents enabled")
		specAgentCount := bamboo.Spec.RemoteAgents.Replicas
		setupLog.Info("Agents in the spec: " +  strconv.FormatInt(int64(specAgentCount), 10))

		runningAgentCount, _, _, _ := GetRunningRemoteAgentDeployments(r, bamboo)
		setupLog.Info("Agent deployments: " +  strconv.FormatInt(runningAgentCount, 10))
		base64Creds := b64.StdEncoding.EncodeToString([]byte(bamboo.Spec.Installer.AdminName + ":" + bamboo.Spec.Installer.AdminPassword))
		err, agentsNumber := rest.GetOnlineAgents("/agent/remote.json?online=true", base64Creds, false)
		if err != nil {
			setupLog.Info("Failed to get registered agents from Bamboo API")
			agentsNumber = 0
		}
		setupLog.Info("Registered agents: " +  strconv.FormatInt(agentsNumber, 10))

		if int64(specAgentCount) > runningAgentCount {
			replicasToAdd := int64(specAgentCount) - runningAgentCount
			for i := 0; i < int(replicasToAdd); i++ {
				err := ScaleRemoteAgents(r, bamboo, bambooAPI)
				if err != nil {
					setupLog.Error(err, "unable to add agent")
				}

			}
		} else if int64(specAgentCount) < runningAgentCount {

			err = ScaleDownRemoteAgents((*deploy.BambooReconciler)(r), bamboo, deploy.BambooAPI(bambooAPI), runningAgentCount )

		}

		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: bamboo.Name, Namespace: bamboo.Namespace}, bambooDeployment)
		bambooReadyReplicas := bambooDeployment.Status.AvailableReplicas
		if bambooReadyReplicas != 1 {
			setupLog.Info("Waiting for Bamboo to load. Agent auto-scaling won't be activated")
			time.Sleep(10 * time.Second)
			err = r.Client.Get(context.TODO(), types.NamespacedName{Name: bamboo.Name, Namespace: bamboo.Namespace}, bambooDeployment)
			bambooReadyReplicas = bambooDeployment.Status.AvailableReplicas
		} else {
			_, remoteAgentImages, remoteAgentDeploymentNames, err := GetRunningRemoteAgentDeployments(r, bamboo)

			if err != nil {
				return ctrl.Result{RequeueAfter: time.Second * 120}, err
			}
			if remoteAgentImages[0] != bamboo.Spec.RemoteAgents.ImageRepo+":"+bamboo.Spec.RemoteAgents.ImageTag {
				setupLog.Info("Updating remote agent deployments with new image:tag. Agents will be stopped.")
				for i := range remoteAgentDeploymentNames {
					agentDeployment := &appsv1.Deployment{}
					err = r.Client.Get(context.TODO(), types.NamespacedName{Name: remoteAgentDeploymentNames[i], Namespace: bamboo.Namespace}, agentDeployment)
					agentDeployment.Spec.Template.Spec.Containers[0].Image = bamboo.Spec.RemoteAgents.ImageRepo + ":" + bamboo.Spec.RemoteAgents.ImageTag
					setupLog.Info("Updating remote agent " + agentDeployment.Name + " with image " + bamboo.Spec.RemoteAgents.ImageRepo + ":" + bamboo.Spec.RemoteAgents.ImageTag)
					err = r.Client.Update(context.TODO(), agentDeployment)
					if err != nil {
						setupLog.Error(err, "unable to update deployment "+agentDeployment.Name)
					}
				}
			}
		}
	}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: bamboo.Name, Namespace: bamboo.Namespace}, bambooDeployment)
	currentBambooVersion := strings.SplitAfter(bambooDeployment.Spec.Template.Spec.Containers[0].Image, ":")[1]
	if len(currentBambooVersion) < 1 {
		setupLog.Info("No tag is set for Bamboo deployment. Assuming latest")
		currentBambooVersion = "latest"
	}

	if bambooDeployment.Spec.Template.Spec.Containers[0].Image != bamboo.Spec.ImageRepo+":"+bamboo.Spec.ImageTag {
		setupLog.Info("New Bamboo image detected:", "image", bamboo.Spec.ImageRepo+":"+bamboo.Spec.ImageTag)
		setupLog.Info("Scaling Bamboo to 0 before making a backup")
		zeroReplicas := int32(0)
		bambooDeployment := deploy.GetBambooDeployment(bamboo, deploy.BambooAPI(bambooAPI))
		bambooDeployment.Spec.Replicas = &zeroReplicas
		err = r.Client.Update(context.TODO(), bambooDeployment)
		if err != nil {
			setupLog.Error(err, "unable to update deployment"+bambooDeployment.Name)
		}

		// wait until Bamboo pod is terminated
		bambooStatus := bambooDeployment.Status.AvailableReplicas
		for bambooStatus != 0 {
			setupLog.Info("Waiting for Bamboo to shut down")
			time.Sleep(5 * time.Second)
			err = r.Client.Get(context.TODO(), types.NamespacedName{Name: bamboo.Name, Namespace: bamboo.Namespace}, bambooDeployment)
			bambooStatus = bambooDeployment.Status.AvailableReplicas
		}
		// wait a bit for the volume to be released
		time.Sleep(15 * time.Second)

		// backup Postgres database
		var databaseBackupSucceeded = true
		setupLog.Info("Backup up database before upgrading Bamboo")
		postgresPodName, _ := k8s.K8sclient.GetDeploymentPod("postgres", bamboo.Namespace)
		postgresBackupCommand := k8s.GetPostgresBackupCommand(bamboo, currentBambooVersion)
		stdout, err := k8s.K8sclient.ExecIntoPod(postgresPodName, postgresBackupCommand, "backup", bamboo.Namespace)
		if len(stdout) < 1 {
			stdout = "Empty response"
		}
		fmt.Printf("Exec received: %s\n", stdout)
		if err != nil {
			setupLog.Error(err, "unable to backup database "+bamboo.Spec.Datasource.Database+". Bamboo deployment will not be upgraded")
			databaseBackupSucceeded = false
		}

		// create Bamboo home backup
		var bambooHomeBackupSucceeded = true
		bambooHomeBackupPod := deploy.GetBackupPod(bamboo, currentBambooVersion)
		err = deploy.CreatePod((*deploy.BambooReconciler)(r), bambooHomeBackupPod, bamboo)
		if err != nil {
			setupLog.Error(err, "unable to create backup pod "+bambooHomeBackupPod.Name)
			bambooHomeBackupSucceeded = false
		} else {

			// wait for backup pod to complete its job
			err = r.Client.Get(context.TODO(), types.NamespacedName{Name: bambooHomeBackupPod.Name, Namespace: bamboo.Namespace}, bambooHomeBackupPod)
			bambooBackupPodStatus := bambooHomeBackupPod.Status.Phase

			for bambooBackupPodStatus != "Succeeded" {
				if bambooBackupPodStatus == "Failed" {
					setupLog.Error(err, "Bamboo home directory backup pod "+bambooHomeBackupPod.Name+" failed. Check pod logs")
					bambooHomeBackupSucceeded = false
					break
				}
				setupLog.Info("Waiting for Bamboo backup pod to complete its job")
				time.Sleep(5 * time.Second)
				err = r.Client.Get(context.TODO(), types.NamespacedName{Name: bambooHomeBackupPod.Name, Namespace: bamboo.Namespace}, bambooHomeBackupPod)
				bambooBackupPodStatus = bambooHomeBackupPod.Status.Phase
			}

		}

		// update Bamboo deployment with new image tag
		if databaseBackupSucceeded && bambooHomeBackupSucceeded {
			setupLog.Info("Update Bamboo with image " + bamboo.Spec.ImageRepo + ":" + bamboo.Spec.ImageTag)
			bambooDeployment = deploy.GetBambooDeployment(bamboo, deploy.BambooAPI(bambooAPI))
			err = r.Client.Update(context.TODO(), bambooDeployment)
			if err != nil {
				setupLog.Error(err, "unable to update deployment"+bambooDeployment.Name)
			}
		}
	}
	if bamboo.Spec.RemoteAgents.Enabled {
		// reconcile agents based on build queue and idle agents
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: bamboo.Name, Namespace: bamboo.Namespace}, bambooDeployment)
		bambooReadyReplicas := bambooDeployment.Status.AvailableReplicas
		if bambooReadyReplicas != 1 {
			setupLog.Info("Waiting for Bamboo to load. Agent auto-scaling won't be activated")
			time.Sleep(10 * time.Second)
			err = r.Client.Get(context.TODO(), types.NamespacedName{Name: bamboo.Name, Namespace: bamboo.Namespace}, bambooDeployment)
			bambooReadyReplicas = bambooDeployment.Status.AvailableReplicas
		} else {
			if bamboo.Spec.RemoteAgents.AutoManagement.Enabled {
				err = ManageAgentPool((*deploy.BambooReconciler)(r), bamboo)
				if err != nil {
					setupLog.Info("unable to use autoscaling based on build queue due to an error. Bamboo maybe starting up or is otherwise unavailable")
				}
			}

		}
	}

	return ctrl.Result{RequeueAfter: time.Second * 30}, nil
}

func (r *BambooReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&installv1alpha1.Bamboo{}).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&apiv1.Service{}).
		Owns(&apiv1.ConfigMap{}).
		Owns(&apiv1.PersistentVolumeClaim{}).
		Owns(&v1beta1.Ingress{}).
		Complete(r)
}
