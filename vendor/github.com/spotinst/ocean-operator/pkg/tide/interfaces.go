// Copyright 2021 NetApp, Inc. All Rights Reserved.

package tide

import (
	"context"

	oceanv1alpha1 "github.com/spotinst/ocean-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

var scheme = runtime.NewScheme()

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = apiextensionsv1.AddToScheme(scheme)
	_ = oceanv1alpha1.AddToScheme(scheme)
	//+kubebuilder:scaffold:scheme
}

// DefaultScheme returns the default runtime.Scheme.
func DefaultScheme() *runtime.Scheme { return scheme }

type (
	// Applier defines the interface used for applying cluster resources.
	Applier interface {
		// ApplyEnvironment applies all resources.
		ApplyEnvironment(
			ctx context.Context,
			options ...ApplyOption) error
		// ApplyComponents applies component resources.
		ApplyComponents(
			ctx context.Context,
			components []*oceanv1alpha1.OceanComponent,
			options ...ApplyOption) error
		// ApplyCRDs applies CRD resources.
		ApplyCRDs(
			ctx context.Context,
			crds []*apiextensionsv1.CustomResourceDefinition,
			options ...ApplyOption) error
		// ApplyRBAC applies RBAC resources.
		ApplyRBAC(
			ctx context.Context,
			serviceAccount *corev1.ServiceAccount,
			roleBinding *rbacv1.ClusterRoleBinding,
			options ...ApplyOption) error
	}

	// Deleter defines the interface used for deleting cluster resources.
	Deleter interface {
		// DeleteEnvironment deletes all resources.
		DeleteEnvironment(
			ctx context.Context,
			options ...DeleteOption) error
		// DeleteComponents deletes component resources.
		DeleteComponents(
			ctx context.Context,
			components []oceanv1alpha1.OceanComponent,
			options ...DeleteOption) error
		// DeleteCRDs deletes CRD resources.
		DeleteCRDs(
			ctx context.Context,
			crds []apiextensionsv1.CustomResourceDefinition,
			options ...DeleteOption) error
		// DeleteRBAC deletes RBAC resources.
		DeleteRBAC(
			ctx context.Context,
			serviceAccount, roleBinding string,
			options ...DeleteOption) error
	}

	// Manager defines the interface used for managing cluster resources.
	Manager interface {
		Applier
		Deleter
	}
)
