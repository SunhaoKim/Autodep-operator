/*
Copyright 2021.

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

package autodep

import (
	"context"
	"fmt"
	appsv1alpha1 "init_rollout_operator/api/v1alpha1"

	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AutodepReconciler reconciles a Autodep object
type AutodepReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

/*
const (
	deploymentfinalizer = "initrolloutoperator"
)
*/
//+kubebuilder:rbac:groups=apps.autodep.com,resources=autodeps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.autodep.com,resources=autodeps/status,verbs=create;get;update;patch
//+kubebuilder:rbac:groups=apps.autodep.com,resources=autodeps/finalizers,verbs=delete;update
//+kubebuilder:rbac:groups=apps.autodep.com,,resources=deployments;statefulsets,verbs=list;watch
// +kubebuilder:rbac:groups=apps.autodep.com,resources=pods;services;services;secrets;external,verbs=get;list;watch
// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Autodep object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *AutodepReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	autodep := &appsv1alpha1.Autodep{}
	//??????autodep??????
	err := r.Get(ctx, req.NamespacedName, autodep)
	if err != nil {
		//????????? not-found ????????????????????????????????????????????????????????????????????????
		//??????????????????????????????????????????????????????????????????kub
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "failed get autodep")
		return ctrl.Result{}, err
	}
	//????????????????????? ????????????????????? ?????????????????????????????????????????????????????????????????? ???????????? ?????????
	/*if autodep.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(autodep, deploymentfinalizer) {
			controllerutil.AddFinalizer(autodep, deploymentfinalizer)
			if err := r.Update(ctx, autodep); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		//?????????deployment
		if controllerutil.ContainsFinalizer(autodep, deploymentfinalizer) {
			if err := r.Delete(ctx, dep); err != nil {
				return ctrl.Result{}, err
			}
			//??????????????????????????????
			controllerutil.RemoveFinalizer(autodep, deploymentfinalizer)
			if err := r.Update(ctx, autodep); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}
	*/
	//??????deployment,??????????????????
	err = r.ensureDEPForAutodepExists(ctx, autodep)
	if err != nil {
		return ctrl.Result{}, err
	}
	err = r.ensureSVCForAutodepExists(ctx, autodep)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}
func GetBackendName(autodep *appsv1alpha1.Autodep) string {
	return fmt.Sprintf("autodep-backend-%s", autodep.Name)
}

func (r *AutodepReconciler) ensureDEPForAutodepExists(ctx context.Context, autodep *appsv1alpha1.Autodep) error {
	depname := GetBackendName(autodep)
	founddeployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Namespace: autodep.Namespace, Name: depname}, founddeployment)
	if err != nil && errors.IsNotFound(err) {
		switch autodep.Spec.Deptype {
		case "backend":
			err = r.CreateBackendDeployment(ctx, autodep)
		}
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		log.Error(err, "failed get deployment for autodep")
		return err
	}
	// get deployment lets update
	switch autodep.Spec.Deptype {
	case "backend":
		err = r.UpdateBackendDeployment(ctx, autodep)
	}
	if err != nil {
		log.Error(err, "failed update deployment for autodep")
		return err
	}
	return nil
}
func (r *AutodepReconciler) ensureSVCForAutodepExists(ctx context.Context, autodep *appsv1alpha1.Autodep) error {
	svcname := GetBackendName(autodep)
	foundservice := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Namespace: autodep.Namespace, Name: svcname}, foundservice)
	if err != nil && errors.IsNotFound(err) {
		switch autodep.Spec.Deptype {
		case "backend":
			err = r.CreateBackendService(ctx, autodep)
		}
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		log.Error(err, "failed get service for autodep")
		return err
	}
	// get deployment lets update
	switch autodep.Spec.Deptype {
	case "backend":
		err = r.UpdateBackendService(ctx, autodep)
	}
	if err != nil {
		log.Error(err, "failed update service for autodep")
		return err
	}
	return nil
}

// your logic here

// SetupWithManager sets up the controller with the Manager.
func (r *AutodepReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.Autodep{}).
		Complete(r)
}
