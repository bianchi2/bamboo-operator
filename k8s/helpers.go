package k8s

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	"time"
)

//GetDeploymentPod queries all pods is a selected namespace by LabelSelector
func (cl *k8s) GetDeploymentPod(name string, ns string) (podName string, err error) {
	api := cl.clientset.CoreV1()
	listOptions := metav1.ListOptions{
		LabelSelector: "k8s-app=" + name,
	}
	podList, _ := api.Pods(ns).List(context.TODO(), listOptions)
	podListItems := podList.Items
	if len(podListItems) == 0 {
		fmt.Printf("Failed to find pod for k8s-app %s. List of pods: %v\n", name, podListItems)
		return "", err
	}
	// expecting only one pod to be there so, taking the first one
	// todo maybe add a unique label to deployments?
	podName = podListItems[0].Name
	return podName, nil
}

func GeneratePasswd(stringLength int) (passwd string) {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("abcdefghijklmnopqrstuvwxyz" + "0123456789")
	length := stringLength
	buf := make([]rune, length)
	for i := range buf {
		buf[i] = chars[rand.Intn(len(chars))]
	}
	passwd = string(buf)
	return passwd
}

func GenerateId(stringLength int) (passwd string) {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("0123456789")
	length := stringLength
	buf := make([]rune, length)
	for i := range buf {
		buf[i] = chars[rand.Intn(len(chars))]
	}
	passwd = string(buf)
	return passwd
}
