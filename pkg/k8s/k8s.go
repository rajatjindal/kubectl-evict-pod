package k8s

import (
	"context"
	"k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Evict tries to evict a pod
func Evict(kubeclient kubernetes.Interface, podname, namespace string) error {
	return kubeclient.PolicyV1beta1().Evictions(namespace).Evict(context.TODO(), &v1beta1.Eviction{
		ObjectMeta: v1.ObjectMeta{
			Name:      podname,
			Namespace: namespace,
		},
	})
}

// ListPods get pods with labelSelector
func ListPods(kubeclient kubernetes.Interface, namespace string, options v1.ListOptions) (podNames []string, err error) {
	podList, err := kubeclient.CoreV1().Pods(namespace).List(context.TODO(), options)
	if err != nil {
		return
	}
	for _, pod := range podList.Items {
		podNames = append(podNames, pod.Name)
	}
	return
}
