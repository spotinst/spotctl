package kubernetes

import (
	"context"
	"encoding/json"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

const Owner = "spotctl"

var decUnstructured = k8syaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

func DoServerSideApply(ctx context.Context, cfg *rest.Config, content string, logger logr.Logger) error {

	// RESTMapper will find group-version-resource
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return err

	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	// Dynamic client
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}

	// YAML manifest -> unstructured.Unstructured
	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode([]byte(content), nil, obj)
	if err != nil {
		return err
	}

	// mapping of group-version-resource (api server destination) from group-version-kind (this object)
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}

	// REST corresponding to group-version-resource
	var dri dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		dri = dyn.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		// for cluster-wide resources
		dri = dyn.Resource(mapping.Resource)
	}

	// to JSON
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	//  Create or Update the object with SSA
	//  types.ApplyPatchType indicates SSA.
	//  FieldManager specifies the field owner ID.
	logger.Info("applying object", "name", obj.GetName())
	force := true
	_, err = dri.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
		FieldManager: Owner,
		Force:        &(force),
	})

	return err
}
