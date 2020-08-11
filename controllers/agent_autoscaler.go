package controllers

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	installv1alpha1 "github.com/bianchi2/bamboo-operator/api/v1alpha1"
	"github.com/bianchi2/bamboo-operator/deploy"
	"github.com/bianchi2/bamboo-operator/rest"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"strconv"
)

func ScaleDownRemoteAgents(r *deploy.BambooReconciler, bamboo *installv1alpha1.Bamboo, bambooAPI deploy.BambooAPI, runningAgentCount int64) (err error) {

	base64Creds := b64.StdEncoding.EncodeToString([]byte(bamboo.Spec.Installer.AdminName + ":" + bamboo.Spec.Installer.AdminPassword))

	// delete agents from Bamboo before scaling down agents statefulset

	desiredReplicas := fmt.Sprint(bamboo.Spec.RemoteAgents.Replicas)
	runningReplicas := fmt.Sprint(runningAgentCount)
	setupLog.Info("Desired replicas: " + desiredReplicas + ". Running replicas " + runningReplicas)
	replicasToRemove := runningAgentCount - int64(bamboo.Spec.RemoteAgents.Replicas)
	setupLog.Info("Removing " + strconv.FormatInt(int64(replicasToRemove), 10) + " agent deployments")

	_, _, deploymentNames, _ := GetRunningRemoteAgentDeployments((*BambooReconciler)(r), bamboo)

	err, agentsIds := rest.GetAgentIdByName("/agent/remote.json?online=true", deploymentNames, bamboo, base64Creds)
	if err != nil {
		fmt.Printf("Unable to get agent IDs. Error: %s\n", err)
		return err
	}
	var agentsToDelete = []string{}
	var deploymentsToDelete = []string{}

	for i := range agentsIds {

		_, busy := rest.GetAgentStatus("/agent/remote.json?online=true", agentsIds[i], base64Creds)
		if busy {
			setupLog.Info("Agent " + agentsIds[i] + " is busy. Cannot delete")
		} else {
			if len(agentsToDelete) >= int(replicasToRemove) {
				// remove only desired number of replicas
				break
			}
				agentsToDelete = append(agentsToDelete, agentsIds[i])
				_, deploymentToDelete := rest.GetAgentNameById("/agent/remote.json?online=true", []string{agentsIds[i]}, bamboo, base64Creds)
				deploymentsToDelete = append(deploymentsToDelete, deploymentToDelete[0])
		}
	}
  
	// delete deployments and PVCs

	for i := range deploymentsToDelete {
		deployment := appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: deploymentsToDelete[i], Namespace: bamboo.Namespace}}
		pvc := apiv1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: deploymentsToDelete[i], Namespace: bamboo.Namespace}}
		setupLog.Info("Deleting deployment: " + deploymentsToDelete[i])
		err := r.Client.Delete(context.TODO(), &deployment)
		if err != nil {
			setupLog.Error(err, "unable to delete deployment " + deployment.Name)
		}
		setupLog.Info("Deleting pvc: " + deploymentNames[i])
		err = r.Client.Delete(context.TODO(), &pvc)
		if err != nil {
			setupLog.Error(err, "unable to delete persistent volume claim " + pvc.Name)
		}
	}
	// delete agents
	err = rest.DeleteAgentsById("/agent", agentsToDelete, base64Creds)
	if err != nil {
		return err
	}


	return nil
}

// ManageAgentPool updates Bamboo spec with the required number of replicas
func ManageAgentPool(r *deploy.BambooReconciler, bamboo *installv1alpha1.Bamboo) (err error) {

	base64Creds := b64.StdEncoding.EncodeToString([]byte(bamboo.Spec.Installer.AdminName + ":" + bamboo.Spec.Installer.AdminPassword))

	setupLog := ctrl.Log.WithName("bamboo-operator")

	err, queuesize := rest.GetQueueSize("/queue.json?expand=queuedBuilds", base64Creds)
	if err != nil {
		fmt.Printf("Failed to get build queue size. Error: %s\n", err)
		return err
	}
	err, agentsNumber := rest.GetOnlineAgents("/agent/remote.json?online=true", base64Creds, false)
	if err != nil {
		fmt.Printf("Failed to get online agents number. Error: %s\n", err)
		return err
	}

	err, idleAgentsNumber := rest.GetOnlineAgents("/agent/remote.json?online=true", base64Creds, true)
	if err != nil {
		fmt.Printf("Failed to get online idle agents number. Error: %s\n", err)
		return err
	}
	err, idleAgents := rest.GetOnlineIdleAgents("/agent/remote.json?online=true", base64Creds)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if idleAgents > int64(bamboo.Spec.RemoteAgents.AutoManagement.MaxIdleAgents) && queuesize == 0 {
		setupLog.Info("Idle agents: " + strconv.FormatInt(idleAgents, 10))
		setupLog.Info("Max idle agents allowed: " + strconv.FormatInt(int64(bamboo.Spec.RemoteAgents.AutoManagement.MaxIdleAgents), 10))
		setupLog.Info("Removing " + strconv.FormatInt(int64(bamboo.Spec.RemoteAgents.AutoManagement.ReplicasToRemove), 10) + " agents")
	}

	if agentsNumber != int64(bamboo.Spec.RemoteAgents.Replicas) {
		setupLog.Info("Registered agents: " + strconv.FormatInt(agentsNumber, 10))
		setupLog.Info("Agents in Bamboo spec: " + strconv.FormatInt(int64(bamboo.Spec.RemoteAgents.Replicas), 10))
		setupLog.Info("Agent auto-scaling (adding agents) won't be triggered until new agents register themselves with the server")
	} else {
		if queuesize > int64(bamboo.Spec.RemoteAgents.AutoManagement.MaxBuildInQueue) {
			setupLog.Info("Automatic scaling of remote agents triggered. Current build queue size is: " + strconv.FormatInt(queuesize, 10))
			replicasToAdd := bamboo.Spec.RemoteAgents.AutoManagement.ReplicasToAdd
			setupLog.Info("Adding: " + strconv.FormatInt(int64(replicasToAdd), 10) + " agent replicas")
			if bamboo.Spec.RemoteAgents.Replicas+1 > bamboo.Spec.RemoteAgents.AutoManagement.MaxReplicas {
				setupLog.Info("Can't add " + strconv.FormatInt(int64(bamboo.Spec.RemoteAgents.Replicas+1), 10) + " agent replicas because Max agent pool size is " + strconv.FormatInt(int64(bamboo.Spec.RemoteAgents.AutoManagement.MaxReplicas), 10))
			} else {
				setupLog.Info("Total number of replicas will be: " + strconv.FormatInt(int64(bamboo.Spec.RemoteAgents.Replicas)+1, 10))
				err := r.Client.Get(context.TODO(), types.NamespacedName{Name: bamboo.Name, Namespace: bamboo.Namespace}, bamboo)
				if err != nil {
					setupLog.Error(err, "unable to get Bamboo CR")
					return err
				}
				bamboo.Spec.RemoteAgents.Replicas = bamboo.Spec.RemoteAgents.Replicas + replicasToAdd
				err = r.Client.Update(context.TODO(), bamboo)
				if err != nil {
					setupLog.Error(err, "unable to update Bamboo CR "+bamboo.Name)
					return err
				}
			}
			// remove agents if build queue is empty and maxIdleAgents threshold is reached
		} else if queuesize == 0 && idleAgentsNumber > int64(bamboo.Spec.RemoteAgents.AutoManagement.MaxIdleAgents) {

			setupLog.Info("Build queue is empty and the number of idle agents exceeds the limit")
			setupLog.Info("Current idle agents: " + strconv.FormatInt(int64(idleAgentsNumber), 10) + " , maxIdleAgent: " + strconv.FormatInt(int64(bamboo.Spec.RemoteAgents.AutoManagement.MaxIdleAgents), 10))
			bamboo.Spec.RemoteAgents.Replicas = bamboo.Spec.RemoteAgents.AutoManagement.MaxIdleAgents
			err = r.Client.Update(context.TODO(), bamboo)
			if err != nil {
				setupLog.Error(err, "unable to update Bamboo CR "+bamboo.Name)
				return err
			}
		}
	}
	return nil
}
