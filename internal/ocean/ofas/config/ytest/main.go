package main

import (
	"context"
	"github.com/spotinst/spotctl/internal/ocean/ofas/config"
	"io"
	corev1 "k8s.io/api/core/v1"
	rbakv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
)

//
//  Testing unstuctured objects to create resources
//

func main() {
	kubeconfig := "/Users/ragnar/.kube/config"
	conf, _ := clientcmd.BuildConfigFromFlags("", kubeconfig)
	clientset, _ := kubernetes.NewForConfig(conf)

	// Sample YAML, you would read this from a file in a real use case
	largeYAML, _ := config.GetDeploymentFile()
	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(largeYAML), 4096)

	for {
		// Decode an individual YAML document into an unstructured object
		var obj unstructured.Unstructured
		if err := decoder.Decode(&obj); err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		// Handle each resource type
		switch obj.GetKind() {
		case "Namespace":
			ns := &corev1.Namespace{}
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, ns); err != nil {
				panic(err)
			}
			_, _ = clientset.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
		case "ServiceAccount":
			sa := &corev1.ServiceAccount{}
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, sa); err != nil {
				panic(err)
			}
			_, _ = clientset.CoreV1().ServiceAccounts(sa.Namespace).Create(context.TODO(), sa, metav1.CreateOptions{})
		case "RoleBinding":
			rb := &rbakv1.RoleBinding{}
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, rb); err != nil {
				panic(err)
			}
			_, _ = clientset.RbacV1().RoleBindings(rb.Namespace).Create(context.TODO(), rb, metav1.CreateOptions{})
		case "ClusterRole":
			cr := &rbakv1.ClusterRole{}
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, cr); err != nil {
				panic(err)
			}
			_, _ = clientset.RbacV1().ClusterRoles().Create(context.TODO(), cr, metav1.CreateOptions{})
		case "ClusterRoleBinding":
			crb := &rbakv1.ClusterRoleBinding{}
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, crb); err != nil {
				panic(err)
			}
			_, _ = clientset.RbacV1().ClusterRoleBindings().Create(context.TODO(), crb, metav1.CreateOptions{})
		}
	}
}
