// Copyright 2021 NetApp, Inc. All Rights Reserved.

package tide

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	oceanv1alpha1 "github.com/spotinst/ocean-operator/api/v1alpha1"
	_ "github.com/spotinst/ocean-operator/pkg/installer/installers"
	"github.com/spotinst/ocean-operator/pkg/log"
	tiderbac "github.com/spotinst/ocean-operator/pkg/tide/rbac"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	OceanOperatorDeployment = "ocean-operator"
	OceanOperatorConfigMap  = "ocean-operator"
	OceanOperatorSecret     = "ocean-operator"
	OceanOperatorChart      = "ocean-operator"
	OceanOperatorRepository = "https://charts.spot.io"
	OceanOperatorVersion    = "" // empty string indicates the latest chart version
	OceanOperatorValues     = ""

	LegacyOceanControllerDeployment = "spotinst-kubernetes-cluster-controller"
	LegacyOceanControllerSecret     = "spotinst-kubernetes-cluster-controller"
	LegacyOceanControllerConfigMap  = "spotinst-kubernetes-cluster-controller-config"
)

var (
	//go:embed components/*
	components        embed.FS
	componentsDirName = "components"

	//go:embed crds/*
	crds        embed.FS
	crdsDirName = "crds"
)

type manager struct {
	clientSet     kubernetes.Interface
	clientGetter  genericclioptions.RESTClientGetter
	clientRuntime client.Client
	log           log.Logger
}

func NewManager(clientGetter genericclioptions.RESTClientGetter, log log.Logger) (Manager, error) {
	config, err := clientGetter.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("cannot get restconfig: %w", err)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to cluster: %w", err)
	}

	clientRuntime, err := NewControllerRuntimeClient(config, scheme)
	if err != nil {
		return nil, err
	}

	return &manager{
		clientSet:     clientSet,
		clientGetter:  clientGetter,
		clientRuntime: clientRuntime,
		log:           log,
	}, nil
}

// region Appliers

func (m *manager) ApplyEnvironment(ctx context.Context, options ...ApplyOption) error {
	opts := mutateApplyOptions(options...)

	oceanSA, oceanBinding, err := m.loadRBAC(opts)
	if err != nil {
		return err
	}
	if err = m.ApplyRBAC(ctx, oceanSA, oceanBinding, options...); err != nil {
		return err
	}

	oceanCRDs, err := m.loadCRDs(opts)
	if err != nil {
		return err
	}
	if err = m.ApplyCRDs(ctx, oceanCRDs, options...); err != nil {
		return err
	}

	oceanComponents, err := m.loadComponents(opts)
	if err != nil {
		return err
	}
	if err = m.ApplyComponents(ctx, oceanComponents, options...); err != nil {
		return err
	}

	return nil
}

func (m *manager) ApplyRBAC(ctx context.Context, sa *corev1.ServiceAccount,
	crb *rbacv1.ClusterRoleBinding, options ...ApplyOption) error {
	opts := mutateApplyOptions(options...)

	m.log.Info("applying tide rbac resources")
	if sa != nil {
		if err := m.applyRBACServiceAccount(ctx, sa, opts); err != nil {
			return err
		}
	}
	if crb != nil {
		if err := m.applyRBACClusterRoleBinding(ctx, crb, opts); err != nil {
			return err
		}
	}
	return nil
}

func (m *manager) applyRBACServiceAccount(ctx context.Context,
	sa *corev1.ServiceAccount, options *ApplyOptions) error {
	if sa.Namespace == "" {
		sa.Namespace = options.Namespace
	}
	if err := m.ensureNamespace(ctx, sa.Namespace); err != nil {
		m.log.Error(err, "unable to create namespace", "namespace", sa.Namespace)
		return err
	}

	existing := sa.DeepCopy()
	objKey := client.ObjectKeyFromObject(sa)
	err := m.clientRuntime.Get(ctx, objKey, existing)
	switch {
	case apierrors.IsNotFound(err):
		m.log.V(1).Info("creating tide service account", "name", objKey.Name)
		if err := m.clientRuntime.Create(ctx, sa); err != nil {
			return fmt.Errorf("unable to create tide service account %q: %w", objKey.Name, err)
		}
	case err != nil:
		return fmt.Errorf("unable to get tide service account %q to check if it exists: %w", objKey.Name, err)
	default:
		m.log.V(1).Info("patching tide service account", "name", objKey.Name)
		if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			if err := m.clientRuntime.Get(ctx, objKey, existing); err != nil {
				return err
			}
			sa.SetResourceVersion(existing.GetResourceVersion())
			return m.clientRuntime.Patch(ctx, sa, client.MergeFrom(existing))
		}); err != nil {
			return err
		}
	}

	return nil
}

func (m *manager) applyRBACClusterRoleBinding(ctx context.Context,
	crb *rbacv1.ClusterRoleBinding, options *ApplyOptions) error {
	existing := crb.DeepCopy()
	objKey := client.ObjectKeyFromObject(crb)
	err := m.clientRuntime.Get(ctx, objKey, existing)
	switch {
	case apierrors.IsNotFound(err):
		m.log.V(1).Info("creating tide cluster role binding", "name", objKey.Name)
		if err := m.clientRuntime.Create(ctx, crb); err != nil {
			return fmt.Errorf("unable to tide create cluster role binding %q: %w", objKey.Name, err)
		}
	case err != nil:
		return fmt.Errorf("unable to get tide cluster role binding %q to check if it exists: %w", objKey.Name, err)
	default:
		m.log.V(1).Info("patching tide cluster role binding", "name", objKey.Name)
		if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			if err := m.clientRuntime.Get(ctx, objKey, existing); err != nil {
				return err
			}
			crb.SetResourceVersion(existing.GetResourceVersion())
			return m.clientRuntime.Patch(ctx, crb, client.MergeFrom(existing))
		}); err != nil {
			return err
		}
	}

	return nil
}

func (m *manager) ApplyCRDs(ctx context.Context,
	crds []*apiextensionsv1.CustomResourceDefinition, options ...ApplyOption) error {
	opts := mutateApplyOptions(options...)

	m.log.Info("applying ocean crds")
	for _, crd := range crds {
		if err := m.applyCRD(ctx, crd, opts); err != nil {
			return err
		}
	}

	return nil
}

func (m *manager) applyCRD(ctx context.Context,
	crd *apiextensionsv1.CustomResourceDefinition, options *ApplyOptions) error {
	if crd.Namespace == "" {
		crd.Namespace = options.Namespace
	}
	if err := m.ensureNamespace(ctx, crd.Namespace); err != nil {
		m.log.Error(err, "unable to create namespace", "namespace", crd.Namespace)
		return err
	}

	existing := crd.DeepCopy()
	objKey := client.ObjectKeyFromObject(crd)
	err := m.clientRuntime.Get(ctx, objKey, existing)
	switch {
	case apierrors.IsNotFound(err):
		m.log.V(1).Info("creating ocean crd", "name", objKey.Name)
		if err := m.clientRuntime.Create(ctx, crd); err != nil {
			return fmt.Errorf("unable to create ocean crd %q: %w", objKey.Name, err)
		}
	case err != nil:
		return fmt.Errorf("unable to get ocean crd %q to check if it exists: %w", objKey.Name, err)
	default:
		m.log.V(1).Info("patching ocean crd", "name", objKey.Name)
		if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			if err := m.clientRuntime.Get(ctx, objKey, existing); err != nil {
				return err
			}
			crd.SetResourceVersion(existing.GetResourceVersion())
			return m.clientRuntime.Patch(ctx, crd, client.MergeFrom(existing))
		}); err != nil {
			return err
		}
	}

	// wait for the crd to be available
	if err = m.waitForCRD(ctx, crd); err != nil {
		if deleteErr := m.clientRuntime.Delete(ctx, crd, nil); deleteErr != nil {
			return fmt.Errorf("unable to delete crd %s: %w "+
				"(deleting crd due: %v)", crd.Name, deleteErr, err)
		}
		return err
	}

	return nil
}

func (m *manager) ApplyComponents(ctx context.Context,
	components []*oceanv1alpha1.OceanComponent, options ...ApplyOption) error {
	opts := mutateApplyOptions(options...)

	m.log.Info("applying ocean components")
	for _, component := range components {
		if err := m.applyComponent(ctx, component, opts); err != nil {
			return err
		}
	}

	return nil
}

func (m *manager) applyComponent(ctx context.Context,
	component *oceanv1alpha1.OceanComponent, options *ApplyOptions) error {
	if component.Spec.State == oceanv1alpha1.OceanComponentStateAbsent {
		m.log.V(1).Info("skipping ocean component",
			"name", component.Name, "state", component.Spec.State)
		return nil
	}
	if component.Namespace == "" {
		component.Namespace = options.Namespace
	}
	if err := m.ensureNamespace(ctx, component.Namespace); err != nil {
		m.log.Error(err, "unable to create namespace", "namespace", component.Namespace)
		return err
	}

	existing := component.DeepCopy()
	objKey := client.ObjectKeyFromObject(component)
	err := m.clientRuntime.Get(ctx, objKey, existing)
	switch {
	case apierrors.IsNotFound(err):
		m.log.V(1).Info("creating ocean component", "name", objKey.Name)
		if err := m.clientRuntime.Create(ctx, component); err != nil {
			return fmt.Errorf("unable to create ocean component %q: %w", objKey.Name, err)
		}
	case err != nil:
		return fmt.Errorf("unable to get ocean component %q to check if it exists: %w", objKey.Name, err)
	default:
		m.log.V(1).Info("patching ocean component", "name", objKey.Name)
		if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			if err := m.clientRuntime.Get(ctx, objKey, existing); err != nil {
				return err
			}
			component.SetResourceVersion(existing.GetResourceVersion())
			if _, present := options.ComponentsFilter[component.Spec.Name]; !present {
				// remove state from patch if it was not explicitly specified
				component.Spec.State = existing.Spec.State
			}
			return m.clientRuntime.Patch(ctx, component, client.MergeFrom(existing))
		}); err != nil {
			return err
		}
	}

	return nil
}

// endregion

// region Deleters

func (m *manager) DeleteEnvironment(ctx context.Context, options ...DeleteOption) error {
	m.log.Info("deleting ocean components")

	componentList := new(oceanv1alpha1.OceanComponentList)
	if err := m.clientRuntime.List(ctx, componentList); err != nil {
		componentGone, ok := err.(*apimeta.NoKindMatchError)
		if ok {
			m.log.V(1).Info("ocean components are not present", "message", componentGone.Error())
		} else {
			return err
		}
	}
	if err := m.DeleteComponents(ctx, componentList.Items, options...); err != nil {
		return err
	}

	crdList := new(apiextensionsv1.CustomResourceDefinitionList)
	crdFieldSet := client.MatchingFields{
		"metadata.name": fmt.Sprintf("oceancomponents.%s", oceanv1alpha1.GroupVersion.Group),
	}
	if err := m.clientRuntime.List(ctx, crdList, crdFieldSet); err != nil {
		crdGone, ok := err.(*apimeta.NoKindMatchError)
		if ok {
			m.log.Info("ocean crds are not present", "message", crdGone.Error())
		} else {
			return err
		}
	}
	if err := m.DeleteCRDs(ctx, crdList.Items, options...); err != nil {
		return err
	}

	if err := m.DeleteRBAC(
		ctx,
		tiderbac.ServiceAccountName,
		tiderbac.RoleBindingName,
		options...,
	); err != nil {
		return err
	}

	return nil
}

func (m *manager) DeleteRBAC(ctx context.Context,
	serviceAccount, roleBinding string, options ...DeleteOption) error {
	opts := mutateDeleteOptions(options...)

	m.log.Info("deleting tide rbac resources")
	if roleBinding != "" {
		if err := m.deleteRBACRoleBinding(ctx, roleBinding, opts); err != nil {
			return err
		}
	}
	if serviceAccount != "" {
		if err := m.deleteRBACServiceAccount(ctx, serviceAccount, opts); err != nil {
			return err
		}
	}
	return nil
}

func (m *manager) deleteRBACServiceAccount(ctx context.Context, name string, options *DeleteOptions) error {
	m.log.V(1).Info("deleting tide service account")

	o := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: options.Namespace,
		},
	}
	if err := m.clientRuntime.Delete(ctx, o); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("could not delete tide service account: %w", err)
	}

	return nil
}

func (m *manager) deleteRBACRoleBinding(ctx context.Context, name string, options *DeleteOptions) error {
	m.log.V(1).Info("deleting tide cluster role binding")

	o := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	if err := m.clientRuntime.Delete(ctx, o); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("could not delete tide cluster role binding: %w", err)
	}

	return nil
}

func (m *manager) DeleteCRDs(ctx context.Context,
	crds []apiextensionsv1.CustomResourceDefinition, options ...DeleteOption) error {
	opts := mutateDeleteOptions(options...)

	m.log.Info("deleting ocean crds")
	for _, crd := range crds {
		if err := m.deleteCRD(ctx, &crd, opts); err != nil {
			return err
		}
	}

	m.log.Info("waiting for ocean crds to be deleted")
	return wait.Poll(5*time.Second, 300*time.Second, func() (bool, error) {
		for _, crd := range crds {
			obj := new(apiextensionsv1.CustomResourceDefinition)
			objKey := types.NamespacedName{
				Namespace: crd.Namespace,
				Name:      crd.Name,
			}
			// wait for IsNotFound on all crds
			if err := m.clientRuntime.Get(ctx, objKey, obj); err == nil {
				return false, nil
			} else if !apierrors.IsNotFound(err) {
				return false, err
			}
		}
		return true, nil
	})
}

func (m *manager) deleteCRD(ctx context.Context,
	crd *apiextensionsv1.CustomResourceDefinition, options *DeleteOptions) error {
	m.log.V(1).Info("deleting ocean crd", "name", crd.Name)

	if err := m.clientRuntime.Delete(ctx, crd); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("could not delete ocean crd: %w", err)
	}

	return nil
}

func (m *manager) DeleteComponents(ctx context.Context,
	components []oceanv1alpha1.OceanComponent, options ...DeleteOption) error {
	opts := mutateDeleteOptions(options...)

	m.log.Info("deleting ocean components")
	for _, component := range components {
		if err := m.deleteComponent(ctx, &component, opts); err != nil {
			return err
		}
	}

	m.log.Info("waiting for ocean components to be deleted")
	return wait.Poll(5*time.Second, 300*time.Second, func() (bool, error) {
		for _, component := range components {
			obj := new(oceanv1alpha1.OceanComponent)
			objKey := types.NamespacedName{
				Namespace: component.Namespace,
				Name:      component.Name,
			}
			// wait for IsNotFound on all components
			if err := m.clientRuntime.Get(ctx, objKey, obj); err == nil {
				return false, nil
			} else if !apierrors.IsNotFound(err) {
				return false, err
			}
		}
		return true, nil
	})
}

func (m *manager) deleteComponent(ctx context.Context,
	component *oceanv1alpha1.OceanComponent, options *DeleteOptions) error {
	m.log.V(1).Info("deleting ocean component", "name", component.Name)

	if err := m.clientRuntime.Delete(ctx, component); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("could not delete ocean component: %w", err)
	}

	return nil
}

// endregion

// region Loaders

func (m *manager) loadCRDs(options *ApplyOptions) ([]*apiextensionsv1.CustomResourceDefinition, error) {
	m.log.V(1).Info("loading ocean crds")

	dd, err := crds.ReadDir(crdsDirName)
	if err != nil {
		return nil, fmt.Errorf("crds in %s cannot be listed: %w", crdsDirName, err)
	}

	manifests := make([]string, len(dd))
	for i, d := range dd {
		manifests[i] = path.Join(crdsDirName, d.Name())
	}
	if len(manifests) == 0 {
		return nil, fmt.Errorf("no crd manifests found")
	}

	oceanCRDs := make([]*apiextensionsv1.CustomResourceDefinition, 0, len(manifests))
	for _, manifest := range manifests {
		crd, err := m.loadCRD(manifest)
		if err != nil {
			return nil, err
		}
		oceanCRDs = append(oceanCRDs, crd)
	}

	return oceanCRDs, nil
}

func (m *manager) loadCRD(name string) (*apiextensionsv1.CustomResourceDefinition, error) {
	crd := new(apiextensionsv1.CustomResourceDefinition)
	data, err := crds.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s: %w", name, err)
	}

	serializer := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, _, err = serializer.Decode(data, &schema.GroupVersionKind{
		Group:   "apiextensionsv1.k8s.io",
		Version: runtime.APIVersionInternal,
		Kind:    "CustomResourceDefinition",
	}, crd)
	if err != nil {
		return nil, fmt.Errorf("cannot load crd %s: %w", name, err)
	}

	m.log.V(1).V(4).Info("loaded crd", "crd", crd)
	return crd, nil
}

func (m *manager) loadComponents(options *ApplyOptions) ([]*oceanv1alpha1.OceanComponent, error) {
	m.log.V(1).Info("loading ocean components")

	dd, err := components.ReadDir(componentsDirName)
	if err != nil {
		return nil, fmt.Errorf("components in %s cannot be listed: %w", componentsDirName, err)
	}

	manifests := make([]string, len(dd))
	for i, d := range dd {
		manifests[i] = path.Join(componentsDirName, d.Name())
	}
	if len(manifests) == 0 {
		return nil, fmt.Errorf("no component manifests found")
	}

	oceanComponents := make([]*oceanv1alpha1.OceanComponent, 0, len(manifests))
	for _, manifest := range manifests {
		comp, err := m.loadComponent(manifest)
		if err != nil {
			return nil, err
		}
		if _, present := options.ComponentsFilter[comp.Spec.Name]; present { // component name
			comp.Spec.State = oceanv1alpha1.OceanComponentStatePresent
		} else if _, present = options.ComponentsFilter[oceanv1alpha1.OceanComponentName(comp.Name)]; present { // resource name
			comp.Spec.State = oceanv1alpha1.OceanComponentStatePresent
		} else {
			comp.Spec.State = oceanv1alpha1.OceanComponentStateAbsent
		}
		oceanComponents = append(oceanComponents, comp)
	}

	return oceanComponents, nil
}

func (m *manager) loadComponent(name string) (*oceanv1alpha1.OceanComponent, error) {
	comp := new(oceanv1alpha1.OceanComponent)
	data, err := components.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s: %w", name, err)
	}

	serializer := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, _, err = serializer.Decode(data, &schema.GroupVersionKind{
		Group:   "ocean.spot.io",
		Version: "v1alpha1",
		Kind:    "OceanComponent",
	}, comp)
	if err != nil {
		return nil, fmt.Errorf("cannot load component %s: %w", name, err)
	}

	m.log.V(4).Info("loaded component", "component", comp)
	return comp, nil
}

func (m *manager) loadRBAC(options *ApplyOptions) (*corev1.ServiceAccount, *rbacv1.ClusterRoleBinding, error) {
	m.log.V(1).Info("loading ocean rbac resources")

	manifests, err := tiderbac.GetRBACManifests(options.Namespace)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get manifests: %w", err)
	}

	sa := new(corev1.ServiceAccount)
	err = yamlutil.NewYAMLOrJSONDecoder(
		strings.NewReader(manifests.ServiceAccount),
		len(manifests.ServiceAccount)).Decode(sa)
	if err != nil {
		return nil, nil, fmt.Errorf("could not decode service account yaml: %w", err)
	}

	crb := new(rbacv1.ClusterRoleBinding)
	err = yamlutil.NewYAMLOrJSONDecoder(
		strings.NewReader(manifests.ClusterRoleBinding),
		len(manifests.ClusterRoleBinding)).Decode(crb)
	if err != nil {
		return nil, nil, fmt.Errorf("could not decode cluster role binding yaml: %w", err)
	}

	m.log.V(4).Info("loaded rbac", "serviceaccount", sa, "clusterrolebinding", crb)
	return sa, crb, nil
}

// endregion

// region Helpers

func (m *manager) waitForCRD(ctx context.Context, crd *apiextensionsv1.CustomResourceDefinition) error {
	return wait.PollImmediate(time.Second, time.Minute, func() (bool, error) {
		m.log.V(1).Info("waiting for ocean crd to be available", "name", crd.Name, "namespace", crd.Namespace)
		existing := crd.DeepCopy()
		objName := client.ObjectKeyFromObject(crd)
		err := m.clientRuntime.Get(ctx, objName, existing)
		if err != nil {
			return false, err
		}
		m.log.V(2).Info("crd status", "status", existing.Status)
		for _, cond := range existing.Status.Conditions {
			m.log.V(2).Info("checking ocean crd condition", "type", cond.Type, "status", cond.Status)
			switch cond.Type {
			case apiextensionsv1.Established:
				if cond.Status == apiextensionsv1.ConditionTrue {
					return true, nil
				}
			case apiextensionsv1.NamesAccepted:
				if cond.Status == apiextensionsv1.ConditionFalse {
					m.log.Error(errors.New(cond.Reason), "name conflict for ocean crd")
					return false, err
				}
			}
		}
		return false, err
	})
}

func (m *manager) ensureNamespace(ctx context.Context, namespace string) error {
	ns := new(corev1.Namespace)
	key := types.NamespacedName{Name: namespace}
	m.log.V(1).Info("checking existence", "namespace", namespace)
	err := m.clientRuntime.Get(ctx, key, ns)
	if apierrors.IsNotFound(err) {
		ns.Name = namespace
		m.log.Info("creating", "namespace", namespace)
		return m.clientRuntime.Create(ctx, ns)
	}
	return err
}

// endregion
