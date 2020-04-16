/*
Copyright 2020 TODO(natef): assign copyright.

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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/cluster-addons/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	addonsv1alpha1 "sigs.k8s.io/cluster-addons/cert-manager/api/v1alpha1"
)

// CertManagerReconciler reconciles a CertManager object
type CertManagerReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	DiscoveryClient *discovery.DiscoveryClient
	DynamicClient   dynamic.Interface
	Datastore       []util.ApplyObject
}

// +kubebuilder:rbac:groups=addons.x-force.netapp.io,resources=certmanagers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=addons.x-force.netapp.io,resources=certmanagers/status,verbs=get;update;patch

func (r *CertManagerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("certmanager", req.NamespacedName)

	// Fetch the object
	instance := addonsv1alpha1.CertManager{}
	if err := r.Get(context.TODO(), req.NamespacedName, &instance); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		r.Log.Error(err, "error reading object")
		return reconcile.Result{}, err
	}

	return r.reconcileExists(context.TODO(), req.NamespacedName, &instance)

}

func (r *CertManagerReconciler) reconcileExists(ctx context.Context, name types.NamespacedName, instance *addonsv1alpha1.CertManager) (reconcile.Result, error) {
	log := r.Log
	log.WithValues("object", name.String()).Info("reconciling")

	// err = r.injectOwnerRef(ctx, instance, objects)
	// if err != nil {
	// 	return reconcile.Result{}, err
	// }
	// var manifestStr string

	// if err := r.kubectl.Apply(ctx, ns, manifestStr, extraArgs...); err != nil {
	// 	log.Error(err, "applying manifest")
	// 	return reconcile.Result{}, fmt.Errorf("error applying manifest: %v", err)
	// }

	for _, a := range r.Datastore {
		log.Info("Applying", "object", a.Unstruc.GetName())
		err := r.Apply(a)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	// var pause = 30 * time.Second
	// var installError error
	// for i := 0; i < 10; i++ {
	// 	installError = r.Apply(a)
	// 	if installError != nil {
	// 		log.Error(installError, "installation error", "retry after", pause)
	// 		time.Sleep(pause)
	// 	} else {
	// 		break
	// 	}
	// }
	// if installError != nil {
	// 	return reconcile.Result{}, installError
	// }
	// }

	return reconcile.Result{}, nil
}

func (r *CertManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	loader := util.NewObjectLoader(mgr.GetScheme())
	objects, err := loader.LoadObjects()
	if err != nil {
		return err
	}

	r.Datastore = objects

	r.Log.Info("loaded objects", "count", len(r.Datastore))
	return ctrl.NewControllerManagedBy(mgr).
		For(&addonsv1alpha1.CertManager{}).
		Complete(r)
}
