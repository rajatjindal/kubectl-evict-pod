package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//get pods from labelSelector
func PodsFromLabel(kubeclient kubernetes.Interface, label, namespace string) ([]string, error) {

	podList, err := kubeclient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: label})
	if err != nil {
		return nil, err
	}

	var podNames []string

	for _, pod := range podList.Items {
		podNames = append(podNames, pod.Name)
	}

	return podNames, nil
}
