package k8s

import (
	"bytes"
	"context"
	"fmt"
	"io"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

type k8s struct {
	clientset kubernetes.Interface
}

var (
	K8sclient = GetK8Client()
)

func GetK8Client() *k8s {
	setupLog := ctrl.Log.WithName("bamboo-operator")

	cfg, err := config.GetConfig()
	if err != nil {
		setupLog.Error(err, "An error occurred when getting K8s config")
	}
	client := k8s{}
	client.clientset, err = kubernetes.NewForConfig(cfg)

	if err != nil {
		setupLog.Error(err, "An error occurred when getting K8s config")
		return nil
	}
	return &client
}
func (cl *k8s) ExecIntoPod(podName string, command string, reason string, namespace string) (string, error) {
	setupLog := ctrl.Log.WithName("bamboo-operator")

	if reason != "" {
		fmt.Printf("Running exec for '%s' in the pod '%s'\n", reason, podName)
	}

	args := []string{"/bin/bash", "-c", command}
	stdout, stderr, err := cl.RunExec(args, podName, namespace)
	if err != nil {
		fmt.Printf("Error running exec: %v, command: %s\n", err, args)
		fmt.Printf("Stderr: %s\n", stderr)
		return stdout, err
	}

	if reason != "" {
		setupLog.Info("Exec successfully completed", "Pod: "+podName, "Namespace: "+namespace)

	}
	return stdout, nil
}

func (cl *k8s) RunExec(command []string, podName, namespace string) (string, string, error) {

	req := cl.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&apiv1.PodExecOptions{
		Command: command,
		Stdin:   false,
		Stdout:  true,
		Stderr:  true,
		TTY:     false,
	}, scheme.ParameterCodec)

	cfg, _ := config.GetConfig()
	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		return "", "", fmt.Errorf("error while creating executor: %v", err)
	}

	var stdout, stderr bytes.Buffer
	var stdin io.Reader
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})
	if err != nil {
		return stdout.String(), stderr.String(), err
	}

	return stdout.String(), stderr.String(), nil
}

func (cl *k8s) GetMetrics()   {
	cfg, _ := config.GetConfig()
	metrics, _ := metrics.NewForConfig(cfg)
	podMetrics, _ := metrics.MetricsV1beta1().PodMetricses("atl").List(context.TODO(), metav1.ListOptions{})
	fmt.Println(podMetrics)
}