package deploy

import (
	"context"
	installv1alpha1 "github.com/bianchi2/bamboo-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BambooReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func CreateConfigMap(r *BambooReconciler, configmap *apiv1.ConfigMap, bamboo *installv1alpha1.Bamboo) error {
	setupLog := ctrl.Log.WithValues()
	err := r.Client.Create(context.TODO(), configmap)
	if err != nil && !errors.IsAlreadyExists(err) {
		setupLog.Error(err, "An error occurred")
		return err
	} else {
		if errors.IsAlreadyExists(err) {
			//setupLog.Info("Configmap "+configmap.Name+" already exists", "Namespace: "+bamboo.Namespace, "Resource: "+bamboo.Name)
		}
		return nil
	}

}

func CreateDeployment(r *BambooReconciler, deployment *appsv1.Deployment, bamboo *installv1alpha1.Bamboo) error {
	setupLog := ctrl.Log.WithValues()

	err := r.Client.Create(context.TODO(), deployment)

	if err != nil && !errors.IsAlreadyExists(err) {
		setupLog.Error(err, "An error occurred")
		return err
	} else {
		if errors.IsAlreadyExists(err) {
			//setupLog.Info("Bamboo deployment already exists", "Namespace: "+bamboo.Namespace, "Resource: "+bamboo.Name)
		}
		return nil
	}
}

func CreateRemoteAgentStatefulSet(r *BambooReconciler, statefulSet *appsv1.StatefulSet, bamboo *installv1alpha1.Bamboo) error {
	setupLog := ctrl.Log.WithValues()

	err := r.Client.Create(context.TODO(), statefulSet)

	if err != nil && !errors.IsAlreadyExists(err) {
		setupLog.Error(err, "An error occurred")
		return err
	} else {
		if errors.IsAlreadyExists(err) {
			//setupLog.Info("Bamboo remote agent StatefulSet already exists", "Namespace: "+bamboo.Namespace, "Resource: "+bamboo.Name)
		}
		return nil
	}
}

func CreatePod(r *BambooReconciler, pod *apiv1.Pod, bamboo *installv1alpha1.Bamboo) error {
	setupLog := ctrl.Log.WithValues()

	err := r.Client.Create(context.TODO(), pod)

	if err != nil && !errors.IsAlreadyExists(err) {
		setupLog.Error(err, "An error occurred")
		return err
	} else {
		if errors.IsAlreadyExists(err) {
			setupLog.Info("Backup pod already exists", "Namespace: "+bamboo.Namespace, "Resource: "+pod.Name)
		}
		return nil
	}
}

func CreateService(r *BambooReconciler, service *apiv1.Service, bamboo *installv1alpha1.Bamboo) error {
	setupLog := ctrl.Log.WithValues()

	err := r.Client.Create(context.TODO(), service)

	if err != nil && !errors.IsAlreadyExists(err) {
		setupLog.Error(err, "An error occurred")
		return err
	} else {
		if errors.IsAlreadyExists(err) {
			//setupLog.Info("Bamboo service already exists", "Namespace: "+bamboo.Namespace, "Resource: "+bamboo.Name)
		}
		return nil
	}
}

func CreateIngress(r *BambooReconciler, ingress *v1beta1.Ingress, bamboo *installv1alpha1.Bamboo) error {
	setupLog := ctrl.Log.WithValues()

	err := r.Client.Create(context.TODO(), ingress)

	if err != nil && !errors.IsAlreadyExists(err) {
		setupLog.Error(err, "An error occurred")
		return err
	} else {
		if errors.IsAlreadyExists(err) {
			//setupLog.Info("Bamboo ingress already exists", "Namespace: "+bamboo.Namespace, "Resource: "+bamboo.Name)
		}
		return nil
	}
}

func CreatePVC(r *BambooReconciler, pvc *apiv1.PersistentVolumeClaim, bamboo *installv1alpha1.Bamboo) error {
	setupLog := ctrl.Log.WithValues()

	err := r.Client.Create(context.TODO(), pvc)

	if err != nil && !errors.IsAlreadyExists(err) {
		setupLog.Error(err, "An error occurred")
		return err
	} else {
		if errors.IsAlreadyExists(err) {
			//setupLog.Info("Bamboo pvc already exists", "Namespace: "+bamboo.Namespace, "Resource: "+bamboo.Name)
		}
		return nil
	}
}

func CreateJob(r *BambooReconciler, job *batchv1.Job, bamboo *installv1alpha1.Bamboo) error {
	setupLog := ctrl.Log.WithValues()

	err := r.Client.Create(context.TODO(), job)

	if err != nil && !errors.IsAlreadyExists(err) {
		setupLog.Error(err, "An error occurred")
		return err
	} else {
		if errors.IsAlreadyExists(err) {
			//setupLog.Info("Bamboo job already exists", "Namespace: "+bamboo.Namespace, "Resource: "+bamboo.Name)
		}
		return nil
	}
}
