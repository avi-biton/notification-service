/*
Copyright 2024.

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

	"github.com/go-logr/logr"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NotificationServiceReconciler reconciles a NotificationService object
type NotificationServiceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=konflux-ci.com,resources=notificationservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=konflux-ci.com,resources=notificationservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=konflux-ci.com,resources=notificationservices/finalizers,verbs=update
// +kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns/finalizers,verbs=update
// +kubebuilder:rbac:groups=tekton.dev,resources=taskruns,verbs=get;list;watch
// +kubebuilder:rbac:groups=tekton.dev,resources=taskruns/status,verbs=get
// +kubebuilder:rbac:groups=appstudio.redhat.com,resources=applications/finalizers,verbs=update
// +kubebuilder:rbac:groups=appstudio.redhat.com,resources=applications,verbs=get;list;watch
// +kubebuilder:rbac:groups=appstudio.redhat.com,resources=applications/status,verbs=get
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=pipelinesascode.tekton.dev,resources=repositories,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NotificationService object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.2/pkg/reconcile
func (r *NotificationServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger := r.Log.WithValues("pipelinerun", req.NamespacedName)
	pipelineRun := &tektonv1.PipelineRun{}

	err := r.Get(ctx, req.NamespacedName, pipelineRun)
	if err != nil {
		logger.Error(err, "Failed to get pipelineRun for", "req", req.NamespacedName)
		fmt.Printf("Failed to get pipelineRun for req %s\n", req.NamespacedName)
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}
	logger.Info("Reconciling PipelineRun", "Name", pipelineRun.Name)
	fmt.Printf("Reconciling PipelineRun - Name %s\n", pipelineRun.Name)

	if !IsFinalizerExistInPipelineRun(pipelineRun, NotificationPipelineRunFinalizer) &&
		!IsAnnotationExistInPipelineRun(pipelineRun, NotificationPipelineRunAnnotation, NotificationPipelineRunAnnotationValue) {
		err = AddFinalizerToPipelineRun(ctx, pipelineRun, r, NotificationPipelineRunFinalizer)
		if err != nil {
			fmt.Printf("Failed to add finalizer - %s\n", err)
		}
	}

	if IsPipelineRunEndedSuccessfully(pipelineRun) &&
		!IsAnnotationExistInPipelineRun(pipelineRun, NotificationPipelineRunAnnotation, NotificationPipelineRunAnnotationValue) {
		results, err := GetResultsFromPipelineRun(pipelineRun)
		if err != nil {
			fmt.Printf("Failed to get results - %s", err)
		} else {
			fmt.Printf("Results are: %s\n", results)
			err = AddAnnotationToPipelineRun(ctx, pipelineRun, r, NotificationPipelineRunAnnotation, NotificationPipelineRunAnnotationValue)
			if err != nil {
				fmt.Printf("Failed to add annotation - %s\n", err)
			}
		}
	}

	if IsPipelineRunEndedSuccessfully(pipelineRun) &&
		IsAnnotationExistInPipelineRun(pipelineRun, NotificationPipelineRunAnnotation, NotificationPipelineRunAnnotationValue) {
		err = RemoveFinalizerFromPipelineRun(ctx, pipelineRun, r, NotificationPipelineRunFinalizer)
		if err != nil {
			fmt.Printf("Failed to remove finalizer - %s\n", err)
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NotificationServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tektonv1.PipelineRun{}).
		// WithEventFilter(predicate.Funcs{
		// 	CreateFunc: func(e event.CreateEvent) bool {
		// 		return true
		// 	},
		// 	UpdateFunc: func(e event.UpdateEvent) bool {
		// 		return true
		// 	},
		// 	DeleteFunc: func(e event.DeleteEvent) bool {
		// 		return true
		// 	},
		// }).
		Complete(r)
}
