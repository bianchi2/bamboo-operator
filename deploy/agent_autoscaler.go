package deploy

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	installv1alpha1 "github.com/bianchi2/bamboo-operator/api/v1alpha1"
	"github.com/bianchi2/bamboo-operator/rest"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"strconv"
)

var (
	setupLog = ctrl.Log.WithValues()
)

func ScaleStatefulSet(r *BambooReconciler, bamboo *installv1alpha1.Bamboo, remoteAgentStatefulSet appsv1.StatefulSet, bambooAPI BambooAPI) (err error) {

	base64Creds := b64.StdEncoding.EncodeToString([]byte(bamboo.Spec.Installer.AdminName + ":" + bamboo.Spec.Installer.AdminPassword))

	// delete agents from Bamboo before scaling down agents statefulset
	if bamboo.Spec.RemoteAgents.Replicas < *remoteAgentStatefulSet.Spec.Replicas {
		desiredReplicas := fmt.Sprint(bamboo.Spec.RemoteAgents.Replicas)
		runningReplicas := fmt.Sprint(*remoteAgentStatefulSet.Spec.Replicas)
		setupLog.Info("Desired replicas: " + desiredReplicas + ". Running replicas " + runningReplicas)
		setupLog.Info("Scaling down " + bamboo.Name + "-agent StatefulSet")
		// pod names in a stateful set start with 0
		orderedReplicas := *remoteAgentStatefulSet.Spec.Replicas - 1
		err, lastStatefulSetAgentPod := rest.GetAgentIdByName("/agent/remote.json?online=true", []string{fmt.Sprint(orderedReplicas)}, bamboo, base64Creds)
		if err != nil {
			fmt.Println("Unable to get the last statefulset agent. Error: %s", err)
			return err
		}
		stringOrderedReplicas := fmt.Sprint(orderedReplicas)

		err, agentsIdToDelete := rest.GetAgentIdByName("/agent/remote.json?online=true", []string{stringOrderedReplicas}, bamboo, base64Creds)
		if err != nil {
			fmt.Println("Unable to get agent IDs. Error: %s", err)
			return err
		}
		_, lastStatefulSetAgentStatus := rest.GetAgentStatus("/agent/remote.json?online=true", lastStatefulSetAgentPod[0], base64Creds)
		if lastStatefulSetAgentStatus {
			setupLog.Info("The last agent in agent Statefulset is busy. Cannot scale down")
		} else {
			// delete agents
			err = rest.DeleteAgentsById("/agent", agentsIdToDelete, base64Creds)
			if err != nil {
				return err
			}
			remoteAgentStatefulSet := GetBambooAgentStatefulSet(bamboo, bambooAPI)
			remoteAgentStatefulSet.Spec.Replicas = &orderedReplicas
			err = r.Client.Update(context.TODO(), remoteAgentStatefulSet)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}
	} else {
		remoteAgentStatefulSet := GetBambooAgentStatefulSet(bamboo, bambooAPI)
		setupLog.Info("Scaling up " + bamboo.Name + "-agent StatefulSet to " + strconv.FormatInt(int64(bamboo.Spec.RemoteAgents.Replicas), 10))
		_ = r.Client.Update(context.TODO(), remoteAgentStatefulSet)
		return nil
	}
	return nil
}

// ManageAgentPool updates Bamboo spec with the required number of replicas
func ManageAgentPool(r *BambooReconciler, bamboo *installv1alpha1.Bamboo) (err error) {

	base64Creds := b64.StdEncoding.EncodeToString([]byte(bamboo.Spec.Installer.AdminName + ":" + bamboo.Spec.Installer.AdminPassword))

	setupLog := ctrl.Log.WithName("bamboo-operator")

	err, queuesize := rest.GetQueueSize("/queue.json?expand=queuedBuilds", base64Creds)
	if err != nil {
		fmt.Println("Failed to get build queue size. Error: %s", err)
		return err
	}
	err, agentsNumber := rest.GetOnlineAgents("/agent/remote.json?online=true", base64Creds, false)
	if err != nil {
		fmt.Println("Failed to get online agents number. Error: %s", err)
		return err
	}

	err, idleAgentsNumber := rest.GetOnlineAgents("/agent/remote.json?online=true", base64Creds, true)
	if err != nil {
		fmt.Println("Failed to get online idle agents number. Error: %s", err)
		return err
	}
	err, idleAgents := rest.GetOnlineIdleAgents("/agent/remote.json?online=true", base64Creds)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if idleAgents > int64(bamboo.Spec.RemoteAgents.AutoManagement.MaxIdleAgents) {
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
