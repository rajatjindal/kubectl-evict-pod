# kubectl-evict-pod

This plugin evicts pods:
- for testing pod disruption budget rules
- safely restarting applications

## Usage

- evict single pod: `kubectl evict-pod <pod-name> -n <namespace>`
- evict multiple pods: `kubectl evict-pod -n <namespace> -l app=foo`
- evict multiple pods until every one is gone: `kubectl evict-pod -n <namespace> -l app=foo --retry`
- show options: `kubectl evict-pod -h`

## Install

```
kubectl krew install evict-pod
```

## Testing

### Pod evicted successfully scenario

```bash
# before running the evict-pod command
$ kubectl get pods -n kube-system
NAME                               READY   STATUS    RESTARTS   AGE
coredns-fb8b8dccf-6wvj6            1/1     Running   0          10m
coredns-fb8b8dccf-826fh            1/1     Running   0          11m

# now lets evict 1 coredns pod
$ ./kubectl-evict-pod coredns-fb8b8dccf-6wvj6 -n kube-system
INFO[0000] pod "coredns-fb8b8dccf-6wvj6" in namespace kube-system evicted successfully 

# the pod has been evicted successfully
$ kubectl get pods -n kube-system
NAME                               READY   STATUS    RESTARTS   AGE
coredns-fb8b8dccf-7ngmk            1/1     Running   0          42s
coredns-fb8b8dccf-826fh            1/1     Running   0          11m
```

### Pod eviction prevented by pod disruption budget

- create the pod disruption budget using following spec
```yaml
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: coredns-pdb
spec:
  minAvailable: 2
  selector:
    matchLabels:
      k8s-app: kube-dns
```

- apply the pod disruption budget to the cluster

```bash
# lets apply the pod disruption budget
$ kubectl apply -f pdb.yaml -n kube-system
poddisruptionbudget.policy/coredns-pdb created
```

- Now lets try to evict the pods again

```bash
# get existing pods
$ kubectl get pods -n kube-system
NAME                               READY   STATUS    RESTARTS   AGE
coredns-fb8b8dccf-7ngmk            1/1     Running   0          10m
coredns-fb8b8dccf-826fh            1/1     Running   0          11m

# now lets try to evict the pod again
$ ./kubectl-evict-pod coredns-fb8b8dccf-826fh -n kube-system
Error: Cannot evict pod as it would violate the pod\'s disruption budget.
exit status 1

# observe pods continue to run
$ kubectl get pods -n kube-system
NAME                               READY   STATUS    RESTARTS   AGE
coredns-fb8b8dccf-7ngmk            1/1     Running   0          10m
coredns-fb8b8dccf-826fh            1/1     Running   0          11m
```

## Development

Update the code and use `go run . <args>` to test.
