package k8s

import (
	policy "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//Evict try to evicts a pod
func Evict(kubeclient kubernetes.Interface, podname, namespace string) error {
	return kubeclient.PolicyV1beta1().Evictions(namespace).Evict(&policy.Eviction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podname,
			Namespace: namespace,
		},
	})
}
