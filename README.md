# k8s-volume-utils
A simple tool to reuse PV. It can rename a PVC or move a PVC from one namespace to another namespace. After that we can resue bound PV without data lost

# Usage:
```
k8s-volume-utils pvc
   It is used to rename a PVC object in kubernetes cluster.
   It will delete original PVC and create new PVC referring to same PV.
   New PVC can be in another namespace.
k8s-volume-utils pv
   It can help to reuse a pv

Usage:
  k8s-volume-utils [command]

Available Commands:
  help        Help about any command
  pv          help to reuse a pv
  pvc         rename PVC object in kubernetes clusterd

```
# Example 
1. create two namespaces for test  
   `oc new-project pvc`  
   `oc new-project pvc1`  
    `oc project pvc ` 
2. create a statefulset and it will create a PV.  
   - Edit sts-pvc-test.yaml to update its `storageClassName` to available one on your cluster.  
   - `oc create -f sts-pvc-test.yaml`  
   - `oc get po www-pvc-test-0 -w ` to wait pod running. Then 
     ```
     $ oc exec pvc-test-0 cat /usr/share/nginx/html/a.txt
      a
     ```
     We can see there is data in PV.
3. delete the statefulset by `oc delete -f sts-pvc-test.yaml`. Now we have a pvc `www-pvc-test-0` to be reused by other service.
4. move pvc to namespace pvc1 by `go run main.go pvc www-pvc-test-0 "" pvc1`   
     `oc get pvc -n pvc1` to check PVC `www-pvc-test-0` is moved to namespace pvc1.  
5. create another statefulset to reuse the same PV.  
   -  Edit sts-pvc-test.yaml  
      Change `"echo a > /usr/share/nginx/html/a.txt;sleep 3600"` to `"echo b > /usr/share/nginx/html/b.txt;sleep 3600"`
   - `oc -n pvc1 create -f sts-pvc-test.yaml`
   - when pod `pvc-test-0 ` is running in namespace `pvc1`. Run `oc exec pvc-test-0 cat /usr/share/nginx/html/a.txt -n pvc1` to check data is not lost.