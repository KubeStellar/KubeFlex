/*
Copyright 2023 The KubeStellar Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	clog "sigs.k8s.io/controller-runtime/pkg/log"

	"mcc.ibm.org/kubeflex/pkg/util"
)

const (
	APIServerDeploymentName = "kube-apiserver"
	CMDeploymentName        = "kube-controller-manager"
	SecurePort              = 9444
	cmHealthzPort           = 10257
)

func (r *ControlPlaneReconciler) ReconcileAPIServerDeployment(ctx context.Context, name string, owner *metav1.OwnerReference) error {
	_ = clog.FromContext(ctx)
	namespace := util.GenerateNamespaceFromControlPlaneName(name)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      APIServerDeploymentName,
			Namespace: namespace,
		},
	}

	err := r.Client.Get(context.TODO(), client.ObjectKeyFromObject(deployment), deployment, &client.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			deployment, err = r.generateAPIServerDeployment(namespace, name)
			if err != nil {
				return err
			}
			util.EnsureOwnerRef(deployment, owner)
			err = r.Client.Create(context.TODO(), deployment, &client.CreateOptions{})
			if err != nil {
				return err
			}
		}
		return err
	}
	return nil
}

func (r *ControlPlaneReconciler) ReconcileCMDeployment(ctx context.Context, name string, owner *metav1.OwnerReference) error {
	_ = clog.FromContext(ctx)
	namespace := util.GenerateNamespaceFromControlPlaneName(name)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CMDeploymentName,
			Namespace: namespace,
		},
	}

	err := r.Client.Get(context.TODO(), client.ObjectKeyFromObject(deployment), deployment, &client.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			deployment, err = r.generateCMDeployment(name, namespace)
			if err != nil {
				return err
			}
			util.EnsureOwnerRef(deployment, owner)
			err = r.Client.Create(context.TODO(), deployment, &client.CreateOptions{})
			if err != nil {
				return err
			}
		}
		return err
	}
	return nil
}

func (r *ControlPlaneReconciler) generateAPIServerDeployment(namespace, dbName string) (*appsv1.Deployment, error) {
	dbPassword, err := r.getDBPassword()
	if err != nil {
		return nil, err
	}
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      APIServerDeploymentName,
			Namespace: namespace,
			Labels: map[string]string{
				"component": "kube-apiserver",
				"tier":      "control-plane",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "kube-apiserver",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "kube-apiserver",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:    "kine",
							Image:   "rancher/kine:v0.9.9-amd64",
							Command: []string{"kine", "--endpoint", fmt.Sprintf("postgres://postgres:%s@%s-postgresql.%s.svc/%s?sslmode=disable", dbPassword, util.DBReleaseName, util.DBNamespace, dbName)},
							Ports: []v1.ContainerPort{{
								ContainerPort: 2379,
							}},
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									"cpu":    resource.MustParse("500m"),
									"memory": resource.MustParse("256Mi"),
								},
								Requests: v1.ResourceList{
									"cpu":    resource.MustParse("250m"),
									"memory": resource.MustParse("64Mi"),
								},
							},
						},
						{
							Name:            "kube-apiserver",
							Image:           "registry.k8s.io/kube-apiserver:v1.27.1",
							ImagePullPolicy: v1.PullIfNotPresent,
							Command: []string{
								"kube-apiserver",
								"--allow-privileged=true",
								"--authorization-mode=Node,RBAC",
								"--client-ca-file=/etc/kubernetes/pki/ca.crt",
								"--enable-admission-plugins=NodeRestriction",
								"--enable-bootstrap-token-auth=true",
								"--etcd-servers=http://127.0.0.1:2379",
								"--kubelet-client-certificate=/etc/kubernetes/pki/apiserver-kubelet-client.crt",
								"--kubelet-client-key=/etc/kubernetes/pki/apiserver-kubelet-client.key",
								"--kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname",
								"--proxy-client-cert-file=/etc/kubernetes/pki/front-proxy-client.crt",
								"--proxy-client-key-file=/etc/kubernetes/pki/front-proxy-client.key",
								"--requestheader-allowed-names=front-proxy-client",
								"--requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt",
								"--requestheader-extra-headers-prefix=X-Remote-Extra-",
								"--requestheader-group-headers=X-Remote-Group",
								"--requestheader-username-headers=X-Remote-User",
								fmt.Sprintf("--secure-port=%d", SecurePort),
								"--service-account-issuer=https://kubernetes.default.svc.cluster.local",
								"--service-account-key-file=/etc/kubernetes/pki/sa.pub",
								"--service-account-signing-key-file=/etc/kubernetes/pki/sa.key",
								"--service-cluster-ip-range=10.96.0.0/12",
								"--tls-cert-file=/etc/kubernetes/pki/apiserver.crt",
								"--tls-private-key-file=/etc/kubernetes/pki/apiserver.key",
							},
							Env: []v1.EnvVar{
								{
									Name: "POD_IP",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "status.podIP",
										},
									},
								},
							},
							Ports: []v1.ContainerPort{{
								ContainerPort: SecurePort,
							}},
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									"cpu":    resource.MustParse("1000m"),
									"memory": resource.MustParse("512Mi"),
								},
								Requests: v1.ResourceList{
									"cpu":    resource.MustParse("256m"),
									"memory": resource.MustParse("250Mi"),
								},
							},
							LivenessProbe: &v1.Probe{
								FailureThreshold: 8,
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/livez",
										Port:   intstr.FromInt(SecurePort),
										Scheme: v1.URISchemeHTTPS,
									},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:       10,
								TimeoutSeconds:      15,
								SuccessThreshold:    1,
							},
							ReadinessProbe: &v1.Probe{
								FailureThreshold: 3,
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/readyz",
										Port:   intstr.FromInt(SecurePort),
										Scheme: v1.URISchemeHTTPS,
									},
								},
								PeriodSeconds:    1,
								TimeoutSeconds:   15,
								SuccessThreshold: 1,
							},
							StartupProbe: &v1.Probe{
								FailureThreshold: 24,
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/livez",
										Port:   intstr.FromInt(SecurePort),
										Scheme: v1.URISchemeHTTPS,
									},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:       10,
								SuccessThreshold:    1,
								TimeoutSeconds:      15,
							},
							VolumeMounts: []v1.VolumeMount{{
								MountPath: "/etc/kubernetes/pki",
								Name:      "k8s-certs",
								ReadOnly:  true,
							}},
						},
					},
					PriorityClassName: "system-node-critical",
					Volumes: []v1.Volume{{
						Name: "k8s-certs",
						VolumeSource: v1.VolumeSource{
							Secret: &v1.SecretVolumeSource{
								SecretName: "k8s-certs",
							},
						},
					}},
				},
			},
		},
	}
	return deployment, nil
}

func (r *ControlPlaneReconciler) generateCMDeployment(cpName, namespace string) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CMDeploymentName,
			Namespace: namespace,
			Labels: map[string]string{
				"component": "kube-controller-manager",
				"tier":      "control-plane",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "kube-controller-manager",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "kube-controller-manager",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "kube-controller-manager",
							Image:           "registry.k8s.io/kube-controller-manager:v1.27.1",
							ImagePullPolicy: v1.PullIfNotPresent,
							Command: []string{
								"kube-controller-manager",
								fmt.Sprintf("--master=https://%s:%d", cpName, SecurePort),
								"--authentication-kubeconfig=/etc/kubernetes/kubeconfig",
								"--authorization-kubeconfig=/etc/kubernetes/kubeconfig",
								"--bind-address=0.0.0.0",
								"--client-ca-file=/etc/kubernetes/pki/ca.crt",
								"--cluster-name=kubernetes",
								"--cluster-signing-cert-file=/etc/kubernetes/pki/ca.crt",
								"--cluster-signing-key-file=/etc/kubernetes/pki/ca.key",
								"--controllers=csrapproving,csrcleaner,csrsigning,namespace,root-ca-cert-publisher,serviceaccount,serviceaccount-token,bootstrapsigner,tokencleaner",
								"--kubeconfig=/etc/kubernetes/kubeconfig",
								"--leader-elect=true",
								"--requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt",
								"--root-ca-file=/etc/kubernetes/pki/ca.crt",
								"--service-account-private-key-file=/etc/kubernetes/pki/sa.key",
								"--use-service-account-credentials=true",
							},
							Ports: []v1.ContainerPort{{
								ContainerPort: SecurePort,
							}},
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									"cpu":    resource.MustParse("250m"),
									"memory": resource.MustParse("128Mi"),
								},
								Requests: v1.ResourceList{
									"cpu":    resource.MustParse("200m"),
									"memory": resource.MustParse("64Mi"),
								},
							},
							LivenessProbe: &v1.Probe{
								FailureThreshold: 8,
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/healthz",
										Port:   intstr.FromInt(cmHealthzPort),
										Scheme: v1.URISchemeHTTPS,
									},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:       10,
								TimeoutSeconds:      15,
							},
							StartupProbe: &v1.Probe{
								FailureThreshold: 24,
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/healthz",
										Port:   intstr.FromInt(cmHealthzPort),
										Scheme: v1.URISchemeHTTPS,
									},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:       10,
								TimeoutSeconds:      15,
							},
							VolumeMounts: []v1.VolumeMount{
								{
									MountPath: "/etc/kubernetes/pki",
									Name:      "k8s-certs",
									ReadOnly:  true,
								},
								{
									MountPath: "/etc/kubernetes/",
									Name:      "cm-kubeconfig",
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "k8s-certs",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: "k8s-certs",
								},
							},
						},
						{
							Name: "cm-kubeconfig",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: "cm-kubeconfig",
								},
							},
						},
					},
				},
			},
		},
	}
	return deployment, nil
}

func (r *ControlPlaneReconciler) getDBPassword() (string, error) {
	// create certs secret object
	pSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GeneratePSecretName(util.DBReleaseName),
			Namespace: util.DBNamespace,
		},
	}

	err := r.Client.Get(context.TODO(), client.ObjectKeyFromObject(pSecret), pSecret, &client.GetOptions{})
	if err != nil {
		return "", err
	}

	return string(pSecret.Data["postgres-password"]), nil
}
