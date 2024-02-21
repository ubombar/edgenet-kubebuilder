/*
Copyright 2024 EdgeNet.

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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	multitenancyedgenetiov1 "github.com/ubombar/edgenet-kubebuilder/api/v1"
	v1 "github.com/ubombar/edgenet-kubebuilder/api/v1"
	util "github.com/ubombar/edgenet-kubebuilder/internal"
)

// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Boilerplate code.
// This is used to retrieve the object with finalizers. In case of any error the error != nil. In case of a requeue request, bool=true. Else
// returns the obj.
// return tuple -> (tenant, isDeleted, requeque, error)
// Note that, if isDeleted is true you need to remove the finalizer from the object to release it.
func (r *TenantReconciler) GetResourceWithFinalizer(ctx context.Context, namespacedName types.NamespacedName) (*v1.Tenant, bool, bool, error) {
	objOld := &v1.Tenant{}
	if err := r.Get(ctx, namespacedName, objOld); err != nil {
		if errors.IsNotFound(err) {
			return nil, false, false, nil
		}
		return nil, false, true, err
	}

	obj := objOld.DeepCopy()

	// If the object is not marked for deletion
	if obj.DeletionTimestamp.IsZero() && !util.ContainsFinalizer(obj.ObjectMeta.Finalizers, "edge-net.io/controller") {
		obj.ObjectMeta.Finalizers = append(obj.ObjectMeta.Finalizers, "edge-net.io/controller")

		if err := r.Update(ctx, obj); err != nil {
			return nil, false, true, err
		}
	}

	return obj, !obj.DeletionTimestamp.IsZero(), false, nil
}

func (r *TenantReconciler) ReleaseResource(ctx context.Context, obj *v1.Tenant) error {
	objCopy := obj.DeepCopy()

	objCopy.ObjectMeta.Finalizers = util.RemoveFinalizer(objCopy.ObjectMeta.Finalizers, "edge-net.io/controller")

	if err := r.Update(ctx, objCopy.DeepCopy()); err != nil {
		return err
	}

	return nil
}

//+kubebuilder:rbac:groups=multitenancy.edge-net.io,resources=tenants,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=multitenancy.edge-net.io,resources=tenants/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=multitenancy.edge-net.io,resources=tenants/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Tenant object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Get the object, isDeleted, requeque, and error
	tenant, isDeleted, requeque, err := r.GetResourceWithFinalizer(ctx, req.NamespacedName)

	if tenant == nil {
		return reconcile.Result{Requeue: requeque}, err
	}

	fmt.Printf("(tenant == nil)=%v, isDeleted=%v\n", tenant == nil, isDeleted)

	// You need to release the resource if it is marked for deletion
	if isDeleted {
		err := r.ReleaseResource(ctx, tenant)
		return reconcile.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *TenantReconciler) OnDeletion(t *v1.Tenant) (ctrl.Result, error) {
	return reconcile.Result{}, nil
}

func (r *TenantReconciler) OnUpdate(t *v1.Tenant) (ctrl.Result, error) {
	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&multitenancyedgenetiov1.Tenant{}).
		Complete(r)
}
