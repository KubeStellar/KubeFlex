# Debugging Kubeflex

## Useful Debugging Hacks
### How to open a psql command in-cluster

```shell
kubectl run -i --tty --rm debug -n vks-system --image=postgres --restart=Never -- bash
psql -h mypsql-postgresql.vks-system.svc -U postgres
```

### How to create a secret with all certs

```shell
kubectl create secret generic k8s-certs-test --from-file=apiserver-kubelet-client.crt=/path/to/.ssh/id_rsa
```

### How to view certs info

```shell
openssl x509 -noout -text -in certs/apiserver.crt 
```

### Get decoded value from secret

```shell
kubectl get secret -n cp3-system admin-kubeconfig -o jsonpath='{.data.kubeconfig}' | base64 -d
kubectl get secrets -n cp1-system k8s-certs -o jsonpath='{.data.apiserver\.crt}' | base64 -d
```

### How to attach a ephemeral container to debug

```shell
kubectl debug -n cp1-system -it kube-controller-manager-676c565f96-r952b --image=busybox:1.28 --target=kube-controller-manager
```

### Getting all the command args for a process

```
cat /proc/<pid>/cmdline | sed -e "s/\x00/ /g"; echo
```

### Manually install the postgres helm chart

```
helm install postgres oci://registry-1.docker.io/bitnamicharts/postgresql --set primary.extendedConfiguration="max_connections = 1000" -n kflex-system
```

### Getting resident memory for processes 

```shell
[root@lsfmx2 kubeflex]# hack/get-rss-info.sh kube-apiserver
PID: 291353, Resident Memory: VmRSS:      320868 kB
PID: 297828, Resident Memory: VmRSS:      247040 kB
PID: 298746, Resident Memory: VmRSS:      246248 kB
[root@lsfmx2 kubeflex]# hack/get-rss-info.sh kine
PID: 297767, Resident Memory: VmRSS:       43536 kB
PID: 298623, Resident Memory: VmRSS:       42784 kB
[root@lsfmx2 kubeflex]# hack/get-rss-info.sh kube-controller-manager
PID: 291369, Resident Memory: VmRSS:       83348 kB
PID: 297144, Resident Memory: VmRSS:       51200 kB
PID: 298672, Resident Memory: VmRSS:       52364 kB
```

For ocm single binary (which includes embedded etcd):

```shell
[root@lsfmx2 kubeflex]# hack/get-rss-info.sh server
PID: 357032, Resident Memory: VmRSS:      253188 kB
```

For kcp (with default workspaces)

```shell
[root@lsfmx2 kubeflex]# hack/get-rss-info.sh kcp
PID: 360147, Resident Memory: VmRSS:      360492 kB
```