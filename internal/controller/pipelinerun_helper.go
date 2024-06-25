package controller

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/konflux-ci/operator-toolkit/metadata"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const NotificationPipelineRunFinalizer string = "konflux.io/notification"
const NotificationPipelineRunAnnotation string = "konflux.io/notified"
const NotificationPipelineRunAnnotationValue string = "True"

// AddFinalizerToPipelineRun adds the finalizer to the PipelineRun.
// If finalizer was not added successfully, a non-nil error is returned.
func AddFinalizerToPipelineRun(ctx context.Context, pipelineRun *tektonv1.PipelineRun, r *NotificationServiceReconciler, finalizer string) error {
	fmt.Printf("Adding finalizer\n")
	patch := client.MergeFrom(pipelineRun.DeepCopy())
	if ok := controllerutil.AddFinalizer(pipelineRun, finalizer); ok {
		err := r.Client.Patch(ctx, pipelineRun, patch)
		if err != nil {
			return fmt.Errorf("Error occurred while patching the updated PipelineRun after finalizer addition: %w", err)
		}
		fmt.Printf("Finalizer was added to PipelineRun %s\n", pipelineRun.Name)
	}
	return nil
}

// RemoveFinalizerFromPipelineRun removes the finalizer from the PipelineRun.
// If finalizer was not removed successfully, a non-nil error is returned.
func RemoveFinalizerFromPipelineRun(ctx context.Context, pipelineRun *tektonv1.PipelineRun, r *NotificationServiceReconciler, finalizer string) error {
	fmt.Printf("Removing finalizer\n")
	patch := client.MergeFrom(pipelineRun.DeepCopy())
	if ok := controllerutil.RemoveFinalizer(pipelineRun, finalizer); ok {
		err := r.Client.Patch(ctx, pipelineRun, patch)
		if err != nil {
			return fmt.Errorf("Error occurred while patching the updated PipelineRun after finalizer removal: %w", err)
		}
		fmt.Printf("Finalizer was removed from %s\n", pipelineRun.Name)
	}
	return nil
}

// IsFinalizerExistInPipelineRun checks if an finalizer exists in pipelineRun
// Return true if yes, otherwise return false
func IsFinalizerExistInPipelineRun(pipelineRun *tektonv1.PipelineRun, finalizer string) bool {
	if controllerutil.ContainsFinalizer(pipelineRun, finalizer) {
		return true
	}
	return false
}

// GetResultsFromPipelineRun extracts results from pipelinerun
// Return error if failed to extract results
func GetResultsFromPipelineRun(pipelineRun *tektonv1.PipelineRun) ([]byte, error) {
	results, err := json.Marshal(pipelineRun.Status.Results)
	if err != nil {
		return nil, fmt.Errorf("Failed to get results from pipelinerun %s: %w", pipelineRun.Name, err)
	}
	return results, nil
}

// IsPipelineRunEndedSuccessfully returns a boolean indicating whether the PipelineRun succeeded or not.
func IsPipelineRunEndedSuccessfully(pipelineRun *tektonv1.PipelineRun) bool {
	return pipelineRun.Status.GetCondition(apis.ConditionSucceeded).IsTrue()
}

// AddNotificationAnnotationToPipelineRun adds an annotation to the PipelineRun.
// If annotation was not added successfully, a non-nil error is returned.
func AddAnnotationToPipelineRun(ctx context.Context, pipelineRun *tektonv1.PipelineRun, r *NotificationServiceReconciler, annotation string, annotationValue string) error {
	fmt.Printf("Adding annotation\n")
	patch := client.MergeFrom(pipelineRun.DeepCopy())
	err := metadata.SetAnnotation(&pipelineRun.ObjectMeta, annotation, annotationValue)
	if err != nil {
		return fmt.Errorf("Error occurred while setting the annotation: %w", err)
	}
	err = r.Client.Patch(ctx, pipelineRun, patch)
	if err != nil {
		fmt.Printf("Error in update annotation client: %s\n", err)
		return fmt.Errorf("Error occurred while patching the updated pipelineRun after annotation addition: %w", err)
	}
	fmt.Printf("Annotation was added to %s\n", pipelineRun.Name)
	return nil
}

// IsNotificationAnnotationExist checks if an annotation exists in pipelineRun
// Return true if yes, otherwise return false
func IsAnnotationExistInPipelineRun(pipelineRun *tektonv1.PipelineRun, annotation string, annotationValue string) bool {
	if metadata.HasAnnotationWithValue(pipelineRun, annotation, annotationValue) {
		return true
	}
	return false
}
