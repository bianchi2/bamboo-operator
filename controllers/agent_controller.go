package controllers

import (
	"context"
	installv1alpha1 "github.com/bianchi2/bamboo-operator/api/v1alpha1"
	"github.com/bianchi2/bamboo-operator/deploy"
	"github.com/bianchi2/bamboo-operator/k8s"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sclient struct {
	clientset kubernetes.Interface
}

var (
	setupLog = ctrl.Log.WithName("bamboo-operator")
)

func GetRunningRemoteAgentDeployments(r *BambooReconciler, bamboo *installv1alpha1.Bamboo) (count int64, images []string, deploymentNames []string, err error) {
	labelSelector := labels.SelectorFromSet(map[string]string{"k8s-app": "k8s-bamboo-agent"})
	listOps := client.ListOptions{LabelSelector: labelSelector, Namespace: bamboo.Namespace}
	deploymentList := appsv1.DeploymentList{}
	err = r.Client.List(context.TODO(), &deploymentList, &listOps)
	if err != nil {
		setupLog.Error(err, "unable to get agent deployments")
		return -1, nil, nil, err
	} else {
		replicas := deploymentList.Items
		count = int64(len(replicas))
		for i := range replicas {
			images = append(images, replicas[i].Spec.Template.Spec.Containers[0].Image)
			deploymentNames = append(deploymentNames, replicas[i].Name)
		}
	}
	return count, images, deploymentNames, nil

}
func ScaleRemoteAgents(r *BambooReconciler, bamboo *installv1alpha1.Bamboo, bambooAPI BambooAPI) (err error) {
	id := k8s.GenerateId(6)
	uid := k8s.GenerateId(8) + "-" + k8s.GenerateId(4) + "-" + k8s.GenerateId(4) + "-" + k8s.GenerateId(4) + "-" + k8s.GenerateId(12)

	createAgentConfigConfigMap := deploy.GetBambooCreateAgentConfigConfigMap(bamboo, deploy.BambooAPI(bambooAPI))

	err = deploy.CreateConfigMap((*deploy.BambooReconciler)(r), createAgentConfigConfigMap, bamboo)
	if err != nil {
		setupLog.Error(err, "unable to create configmap "+createAgentConfigConfigMap.Name)
		return err

	}

	bambooAgentCfgConfigMap := deploy.GetBambooAgentCfgConfigMap(bamboo, deploy.BambooAPI(bambooAPI))
	err = deploy.CreateConfigMap((*deploy.BambooReconciler)(r), bambooAgentCfgConfigMap, bamboo)
	if err != nil {
		setupLog.Error(err, "unable to create configmap "+bambooAgentCfgConfigMap.Name)
		return err

	}

	agentPVC := deploy.GetPVC(bamboo, "remote-agent-"+id, deploy.BambooAPI(bambooAPI))
	err = deploy.CreatePVC((*deploy.BambooReconciler)(r), agentPVC, bamboo)
	if err != nil {
		setupLog.Error(err, "unable to create pvc "+agentPVC.Name)
		return err

	}

	bambooAgentDeployment := deploy.GetBambooAgentDeployment(bamboo, deploy.BambooAPI(bambooAPI), id, uid)
	setupLog.Info("Starting remote agent " + bambooAgentDeployment.Name)

	err = deploy.CreateDeployment((*deploy.BambooReconciler)(r), bambooAgentDeployment, bamboo)

	if err != nil {
		setupLog.Error(err, "unable to create Bamboo agent deployment "+id)
		return err
	}
	return nil
}
