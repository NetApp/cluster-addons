package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/cluster-addons/util"
)

var (
	gkAPIServer = schema.GroupKind{Group: "apiregistration.k8s.io", Kind: "APIService"}
)

func (r *CertManagerReconciler) Discover(obj runtime.Object) (*schema.GroupVersionKind, string, error) {
	gvk := obj.GetObjectKind().GroupVersionKind()

	groupResources, err := restmapper.GetAPIGroupResources(r.DiscoveryClient)
	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	m, err := mapper.RESTMappings(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, "", err
	}
	if len(m) == 0 {
		return nil, "", fmt.Errorf("No resource found for %s", gvk.String())
	}
	resourceName := m[0].Resource.Resource

	return &gvk, resourceName, nil
}

func (r *CertManagerReconciler) Apply(ao util.ApplyObject) error {

	log := ctrl.Log.WithName("applier")

	gvk, resourceName, err := r.Discover(ao.Runtime)
	if err != nil {
		return err
	}

	fmt.Printf("%s   %s/%s \n", gvk.GroupKind(), ao.Unstruc.GetNamespace(), ao.Unstruc.GetName())
	// create the object using the dynamic client
	gvr := schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: resourceName,
	}

	ns := ao.Unstruc.GetNamespace()
	var patched *unstructured.Unstructured
	var iface dynamic.ResourceInterface
	if ns == "" {
		iface = r.DynamicClient.Resource(gvr)
	} else {
		iface = r.DynamicClient.Resource(gvr).Namespace(ns)
	}

	forceApply := true // https://github.com/kubernetes/kubernetes/issues/89954
	// see also https://github.com/kubernetes-sigs/structured-merge-diff/issues/130
	// :( :( https://github.com/kubernetes/kubernetes/issues/89264
	if gvk.GroupKind() == gkAPIServer {
		log.Info("Can't \"apply\" APIService %s because of https://github.com/kubernetes/kubernetes/issues/89264", "object", ao.Unstruc.GetName())
		_, err = iface.Get(context.TODO(), ao.Unstruc.GetName(), metav1.GetOptions{TypeMeta: metav1.TypeMeta{Kind: gvk.Kind, APIVersion: gvk.Version}})
		if err != nil {
			log.Error(err, "Error in Get")
		}
		if err != nil && errors.IsNotFound(err) {
			patched, err = iface.Create(context.TODO(), ao.Unstruc, metav1.CreateOptions{})
			if err != nil {
				log.Error(err, "Error in Create")
			}
		} else {
			if err != nil {
				log.Error(err, "can't get resource")
				return fmt.Errorf("cannot get resource %s %s, %w", gvr.String(), ao.Unstruc.GetName(), err)
			}
		}
	} else {
		patched, err = iface.Patch(
			context.TODO(),
			ao.Unstruc.GetName(),
			types.ApplyPatchType,
			ao.Raw,
			metav1.PatchOptions{
				Force:        &forceApply,
				FieldManager: "kube-put",
			},
		)
	}

	if err != nil {
		log.Error(err, "Error")
		return fmt.Errorf("cannot apply resource %s %s, %w", gvr.String(), ao.Unstruc.GetName(), err)
	}

	if patched == nil {
		log.Info("skipped", "object", ao.Unstruc.GetName())
	} else {
		log.Info("applied", "object", patched.GetName())
	}
	return nil
}
