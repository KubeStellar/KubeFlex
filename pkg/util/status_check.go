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

package util

import (
	"context"
	"fmt"
	"log"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

func WaitForDeploymentReady(clientset kubernetes.Clientset, name, namespace string) error {
	watcher, err := clientset.AppsV1().Deployments(namespace).Watch(context.Background(), metav1.ListOptions{
		FieldSelector:   fmt.Sprintf("metadata.name=%s", name),
		ResourceVersion: "0",
	})
	if err != nil {
		return err
	}
	defer watcher.Stop()

	for {
		event, ok := <-watcher.ResultChan()
		if !ok {
			return nil
		}

		switch event.Type {
		case watch.Error:
			log.Println("Error watching deployment:", watch.Error)
		case watch.Added, watch.Modified:
			deploy := event.Object.(*v1.Deployment)
			if deploy.Status.ReadyReplicas == deploy.Status.Replicas && deploy.Status.Replicas == *deploy.Spec.Replicas {
				return nil
			}
		case watch.Deleted:
			return nil
		}
	}
}

func WaitForStatefulSetReady(clientset kubernetes.Clientset, name, namespace string) error {
	watcher, err := clientset.AppsV1().StatefulSets(namespace).Watch(context.Background(), metav1.ListOptions{
		FieldSelector:   fmt.Sprintf("metadata.name=%s", name),
		ResourceVersion: "0",
	})
	if err != nil {
		return err
	}
	defer watcher.Stop()

	for {
		event, ok := <-watcher.ResultChan()
		if !ok {
			return nil
		}

		switch event.Type {
		case watch.Error:
			log.Println("Error watching statefulset:", watch.Error)
		case watch.Added, watch.Modified:
			stset := event.Object.(*v1.StatefulSet)
			if stset.Status.ReadyReplicas == stset.Status.Replicas && stset.Status.Replicas == *stset.Spec.Replicas {
				return nil
			}
		case watch.Deleted:
			return nil
		}
	}
}

func WaitForNamespaceDeletion(clientset kubernetes.Clientset, name string) error {
	watcher, err := clientset.CoreV1().Namespaces().Watch(context.Background(), metav1.ListOptions{
		FieldSelector:   fmt.Sprintf("metadata.name=%s", name),
		ResourceVersion: "0",
	})
	if err != nil {
		return err
	}
	defer watcher.Stop()

	for {
		event, ok := <-watcher.ResultChan()
		if !ok {
			return nil
		}

		switch event.Type {
		case watch.Error:
			log.Println("Error watching namespace:", watch.Error)
		case watch.Added, watch.Modified:
			namespace := event.Object.(*corev1.Namespace)
			if namespace.Status.Phase == corev1.NamespaceTerminating {
				continue
			}
		case watch.Deleted:
			return nil
		}
	}
}
