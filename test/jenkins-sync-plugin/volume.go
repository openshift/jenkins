package e2e

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
)

func (ut *Tester) SetupK8SNFSServerAndVolume() (*corev1.Pod, []*corev1.PersistentVolume) {
	privBinding := &rbacv1.ClusterRoleBinding{}
	privBinding.Name = "priv-scc-binding-nfs-pvs"
	privBinding.RoleRef = rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     "system:openshift:scc:privileged",
	}
	privBinding.Subjects = []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      "default",
			Namespace: ut.namespace,
		},
	}
	_, err := ut.client.RbacV1().ClusterRoleBindings().Create(context.Background(), privBinding, metav1.CreateOptions{})
	if err != nil && !kerrors.IsAlreadyExists(err) {
		ut.t.Fatalf("error creating cluster role binding: %s", err.Error())
	}

	image := "quay.io/redhat-developer/nfs-server:1.0"
	prefix := "nfs"
	ports := []int{2049}
	vols := map[string]string{"": "/exports"}

	serverPodPorts := make([]corev1.ContainerPort, 1)

	for i := 0; i < len(ports); i++ {
		portName := fmt.Sprintf("%s-%d", prefix, i)

		serverPodPorts[i] = corev1.ContainerPort{
			Name:          portName,
			ContainerPort: int32(ports[i]),
			Protocol:      corev1.ProtocolTCP,
		}
	}

	volumeCount := len(vols)
	volumes := make([]corev1.Volume, volumeCount)
	mounts := make([]corev1.VolumeMount, volumeCount)

	i := 0
	for src, dst := range vols {
		mountName := fmt.Sprintf("path%d", i)
		volumes[i].Name = mountName
		if src == "" {
			volumes[i].VolumeSource.EmptyDir = &corev1.EmptyDirVolumeSource{}
		} else {
			volumes[i].VolumeSource.HostPath = &corev1.HostPathVolumeSource{
				Path: src,
			}
		}

		mounts[i].Name = mountName
		mounts[i].ReadOnly = false
		mounts[i].MountPath = dst

		i++

	}

	serverPodName := fmt.Sprintf("%s-server", prefix)
	privileged := new(bool)
	*privileged = true
	restartPolicy := corev1.RestartPolicyAlways

	serverPod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: serverPodName,
			Labels: map[string]string{
				"role": serverPodName,
			},
		},

		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  serverPodName,
					Image: image,
					SecurityContext: &corev1.SecurityContext{
						Privileged: privileged,
					},
					Ports:        serverPodPorts,
					VolumeMounts: mounts,
				},
			},
			Volumes:       volumes,
			RestartPolicy: restartPolicy,
		},
	}

	serverPod, err = ut.client.CoreV1().Pods(ut.namespace).Create(context.Background(), serverPod, metav1.CreateOptions{})
	if err != nil {
		ut.t.Fatalf("error creating pod %s: %s", serverPodName, err.Error())
	}

	err = wait.PollImmediate(1*time.Second, 30*time.Second, func() (done bool, err error) {
		serverPod, err = ut.client.CoreV1().Pods(ut.namespace).Get(context.Background(), serverPod.Name, metav1.GetOptions{})
		if err != nil {
			ut.t.Logf("error on pod get for %s: %s", serverPod.Name, err.Error())
			return false, nil
		}
		if len(serverPod.Status.PodIP) == 0 {
			ut.t.Logf("ip for pod %s still not set", serverPod.Name)
			return false, nil
		}
		if serverPod.Status.Phase != corev1.PodRunning {
			ut.t.Logf("pod %s phase %s", serverPod.Name, serverPod.Status.Phase)
			return false, nil
		}
		return true, nil
	})

	pvs := []*corev1.PersistentVolume{}
	volLabel := labels.Set{"e2e-pv-pool": ut.namespace}
	for i := 0; i < 3; i++ {
		pv := &corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "nfs",
				Labels:       volLabel,
				Annotations: map[string]string{
					"pv.beta.kubernetes.io/gid": "777",
				},
			},
			Spec: corev1.PersistentVolumeSpec{
				AccessModes:                   []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
				PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
				Capacity: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("3Gi"),
				},
				PersistentVolumeSource: corev1.PersistentVolumeSource{
					NFS: &corev1.NFSVolumeSource{
						Server:   serverPod.Status.PodIP,
						Path:     fmt.Sprintf("/exports/data-%d", i),
						ReadOnly: false,
					},
				},
			},
		}
		pv, err = ut.client.CoreV1().PersistentVolumes().Create(context.Background(), pv, metav1.CreateOptions{})
		if err != nil {
			ut.t.Fatalf("error creating pv: %s", err.Error())
		}
		pvs = append(pvs, pv)
	}
	return serverPod, pvs
}
