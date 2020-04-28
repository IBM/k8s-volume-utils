//
// Copyright 2020 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package kubeutils

import (
	"context"
	"fmt"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

var kclient *kubernetes.Clientset
var defaultNamespace string
var ctx context.Context

func RenamePVC(name string, tName string, tNS string, sNS string) error {
	srcName, srcNamespace, dstName, dstNamespace := initPVCCmdParas(name, tName, tNS, sNS)
	var err error
	var pvc *v1.PersistentVolumeClaim
	var pv *v1.PersistentVolume
	pvc, err = getSrcPVC(srcName, srcNamespace)
	if err != nil {
		klog.Errorf("fail to get source pvc %s. errors: %s", srcName, err.Error())
		return err
	}
	pv, err = getBoundPV(pvc)
	if err != nil {
		klog.Errorf("fail to get bound pv. errors: %s", err.Error())
		return err
	}

	err = retainPV(pv)
	if err != nil {
		klog.Errorf("fail to retain pv. errors: %s", err.Error())
		return err
	}
	err = deletePVC(pvc)
	if err != nil {
		klog.Errorf("fail to delete pvc %s. errors: %s", pvc.Name, err.Error())
		return err
	}
	err = reusePV(pv)
	if err != nil {
		klog.Errorf("fail to mark pv available", err.Error())
		return err
	}
	err = copyPVC(pvc, dstName, dstNamespace)
	if err != nil {
		klog.Errorf("fail to create new pvc: %s. errors: %s", dstName, err.Error())
		return err
	}
	recoverPVPolicy(pv)
	klog.Info("complete successfully")
	return nil

}

func RetainPV(name string) error {
	pv, err := kclient.CoreV1().PersistentVolumes().Get(name, metav1.GetOptions{})
	if err != nil {
		klog.Error("fail to ger pv object")
		return err
	}
	klog.Infof("pv reclaim policy of %s is changed to retain now", name)
	return retainPV(pv)

}

func ReusePV(name string) error {
	pv, err := kclient.CoreV1().PersistentVolumes().Get(name, metav1.GetOptions{})
	if err != nil {
		klog.Error("fail to ger pv object")
		return err
	}
	klog.Infof("pv %s is available to be bound", name)
	return reusePV(pv)

}
func recoverPVPolicy(pv *v1.PersistentVolume) {
	newPV, err := kclient.CoreV1().PersistentVolumes().Get(pv.Name, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("fail to get pv %s. errors: %s", pv.Name, err.Error())
	}
	newPV.Spec.PersistentVolumeReclaimPolicy = pv.Spec.PersistentVolumeReclaimPolicy
	_, err = kclient.CoreV1().PersistentVolumes().Update(newPV)
	if err != nil {
		klog.Warning("fail to recover pv reclaim policy")
	}
	time.Sleep(200)

}
func copyPVC(src *v1.PersistentVolumeClaim, dstName string, dstNamespace string) error {
	newPVC := src.DeepCopy()
	if _, ok := newPVC.ObjectMeta.Annotations["pv.kubernetes.io/bind-completed"]; ok {
		delete(newPVC.ObjectMeta.Annotations, "pv.kubernetes.io/bind-completed")
	}
	newPVC.ObjectMeta.Name = dstName
	newPVC.ObjectMeta.Namespace = dstNamespace
	newPVC.ObjectMeta.CreationTimestamp = metav1.Now()
	newPVC.ObjectMeta.ResourceVersion = ""
	newPVC.ObjectMeta.SelfLink = ""
	newPVC.ObjectMeta.UID = ""
	newPVC.Status = v1.PersistentVolumeClaimStatus{}
	if _, err := kclient.CoreV1().PersistentVolumeClaims(dstNamespace).Create(newPVC); err != nil {
		return err
	}
	time.Sleep(500)
	klog.Infof("create new pvc %s/%s", newPVC.Namespace, newPVC.Name)
	return nil

}

func reusePV(pv *v1.PersistentVolume) error {
	time.Sleep(100)
	npv, err := kclient.CoreV1().PersistentVolumes().Get(pv.Name, metav1.GetOptions{})
	if _, ok := npv.ObjectMeta.Annotations["pv.kubernetes.io/bound-by-controller"]; ok {
		delete(npv.ObjectMeta.Annotations, "pv.kubernetes.io/bound-by-controller")
	}
	npv.Spec.ClaimRef = nil
	npv, err = kclient.CoreV1().PersistentVolumes().Update(npv)
	if err != nil {
		return err
	}
	var n int
	for {
		time.Sleep(500)
		npv, err = kclient.CoreV1().PersistentVolumes().Get(pv.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("fail to get pv new state")
		}
		if npv.Status.Phase == v1.VolumeAvailable {
			return nil
		}
		if n > 600 {
			return fmt.Errorf("fail to wait pv available")
		}
		n = n + 1
	}
	klog.Infof("pv %s gets available", npv.Name)
	return nil
}
func deletePVC(pvc *v1.PersistentVolumeClaim) error {
	err := kclient.CoreV1().PersistentVolumeClaims(pvc.ObjectMeta.Namespace).Delete(pvc.Name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	time.Sleep(3000)
	klog.Infof("delete pvc: %s/%s", pvc.Namespace, pvc.Name)
	return nil

}
func retainPV(pv *v1.PersistentVolume) error {
	oPolicy := pv.Spec.PersistentVolumeReclaimPolicy
	npv := pv.DeepCopy()
	if oPolicy != v1.PersistentVolumeReclaimRetain {
		klog.Infof("update reclian policy of pv %s from %s to %s", npv.ObjectMeta.Name, oPolicy, v1.PersistentVolumeReclaimRetain)
		npv.Spec.PersistentVolumeReclaimPolicy = v1.PersistentVolumeReclaimRetain
		_, err := kclient.CoreV1().PersistentVolumes().Update(npv)
		if err != nil {
			return fmt.Errorf("fail to retain pv: %s", err.Error())
		}
		var n int
		for {
			time.Sleep(100)
			npv, err = kclient.CoreV1().PersistentVolumes().Get(pv.Name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("fail to get pv new state after updating reclaim policy to retain")
			}
			if npv.Spec.PersistentVolumeReclaimPolicy == v1.PersistentVolumeReclaimRetain {
				return nil
			}
			if n > 3 {
				return fmt.Errorf("fail to wait pv reclaim policy to update to retain")
			}
			n = n + 1
		}
		return nil
	}
	return nil
}
func getSrcPVC(srcName string, srcNamespace string) (*v1.PersistentVolumeClaim, error) {
	pvc, err := kclient.CoreV1().PersistentVolumeClaims(srcNamespace).Get(srcName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("fail to get source pvc: %s", err.Error())
	}
	return pvc, nil
}
func getBoundPV(pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolume, error) {
	ann, ok := pvc.Annotations["pv.kubernetes.io/bind-completed"]
	if !ok || ann != "yes" {
		return nil, fmt.Errorf("pvc is not bound yet. no annotation 'pv.kubernetes.io/bind-completed=yes'")
	}
	if pvc.Spec.VolumeName == "" {
		return nil, fmt.Errorf("volume name is empty")
	}
	pv, err := kclient.CoreV1().PersistentVolumes().Get(pvc.Spec.VolumeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get bound pv: %s. error mesage: %s", pvc.Spec.VolumeName, err.Error())
	}
	return pv, nil

}
func initPVCCmdParas(name string, tName string, tNS string, sNS string) (string, string, string, string) {
	if name == "" {
		klog.Errorf("source pvc name not set")
		os.Exit(1)
	}
	var srcName, dstName, srcNamespace, dstNamespace string
	srcName = name
	if tName == "" && tNS == "" {
		klog.Errorf("both target pvc name and target namespace not set")
	}
	if tName == "" {
		dstName = name
	} else {
		dstName = tName
	}
	if sNS == "" {
		srcNamespace = defaultNamespace
	} else {
		srcNamespace = sNS
	}
	if tNS == "" {
		dstNamespace = srcNamespace
	} else {
		dstNamespace = tNS

	}
	if srcName == dstName && srcNamespace == dstNamespace {
		klog.Info("target pvc is same as source pvc. no change.")
		os.Exit(0)

	}
	klog.Infof("we are about to rename pvc %s/%s to %s/%s", srcNamespace, srcName, dstNamespace, dstName)
	return srcName, srcNamespace, dstName, dstNamespace

}
func InitKube(client *kubernetes.Clientset, namespace string) {
	ctx = context.TODO()
	kclient = client
	defaultNamespace = namespace

}
