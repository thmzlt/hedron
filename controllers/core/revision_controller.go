/*
Unlicensed
*/

package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1beta1 "github.com/thmzlt/hedron/apis/core/v1beta1"
)

// RevisionReconciler reconciles a Revision object
type RevisionReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core.hedron.build,resources=revisions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.hedron.build,resources=revisions/status,verbs=get;update;patch

func (r *RevisionReconciler) Reconcile(request ctrl.Request) (ctrl.Result, error) {
	requestCtx := context.WithValue(context.Background(), contextKeyRequest, request)

	revision, err := r.fetchRevision(requestCtx)
	if err != nil && strings.Contains(err.Error(), "not found") {
		r.Log.Info("Revision no longer exists", "revision", request.NamespacedName)

		return ctrl.Result{}, client.IgnoreNotFound(err)
	} else if err != nil {
		r.Log.Error(err, "Failed to fetch revision")

		return ctrl.Result{}, err
	}

	revisionCtx := context.WithValue(requestCtx, contextKeyRevision, revision)

	if revision.Status.State == "Failed" || revision.Status.State == "Succeeded" {
		state := strings.ToLower(string(revision.Status.State))
		r.Log.Info(fmt.Sprintf("Revision is %s", state))

		return ctrl.Result{}, nil
	}

	job, err := r.fetchJob(revisionCtx)
	if err != nil && strings.Contains(err.Error(), "not found") {
		job, err = r.createJob(revisionCtx)
		if err != nil {
			r.Log.Error(err, "Failed to create job")
		}
	} else if err != nil {
		r.Log.Error(err, "Failed to fetch job")
	}

	if job.Status.Active == 1 {
		revision.Status.State = "Pending"
	}
	if job.Status.Failed == 1 {
		revision.Status.State = "Failed"
	}
	if job.Status.Succeeded == 1 {
		revision.Status.State = "Succeeded"
	}

	if err = r.Update(revisionCtx, &revision); err != nil {
		r.Log.Error(err, "Failed to update revision state")
	}
	r.Log.Info("Updated revision state", "state", revision.Status.State)

	return ctrl.Result{}, nil
}

func (r *RevisionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ownerKey := ".metadata.controller"
	apiGVStr := v1beta1.GroupVersion.String()

	if err := mgr.GetFieldIndexer().IndexField(&batchv1.Job{}, ownerKey, func(object runtime.Object) []string {
		job := object.(*batchv1.Job)
		owner := metav1.GetControllerOf(job)

		// Not a match (not controlled at all)
		if owner == nil {
			return nil
		}
		// Not a match (not controlled by a revision)
		if owner.APIVersion != apiGVStr || owner.Kind != "Revision" {
			return nil
		}
		// Match
		return []string{owner.Name}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Revision{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}

func (r *RevisionReconciler) createJob(ctx context.Context) (batchv1.Job, error) {
	revision := ctx.Value(contextKeyRevision).(v1beta1.Revision)

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      revision.Name,
			Namespace: revision.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "hedron-runner",
							Image:   "ghcr.io/thmzlt/hedron-runner:latest",
							Command: []string{"sleep", "30"},
						},
					},
					RestartPolicy: "Never",
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(&revision, &job, r.Scheme); err != nil {
		return job, err
	}

	return job, r.Create(ctx, &job)
}

func (r *RevisionReconciler) fetchJob(ctx context.Context) (batchv1.Job, error) {
	var job batchv1.Job

	request := ctx.Value(contextKeyRequest).(ctrl.Request)
	jobName := request.NamespacedName

	return job, r.Get(ctx, jobName, &job)
}

func (r *RevisionReconciler) fetchRevision(ctx context.Context) (v1beta1.Revision, error) {
	var revision v1beta1.Revision

	request := ctx.Value(contextKeyRequest).(ctrl.Request)
	return revision, r.Get(ctx, request.NamespacedName, &revision)
}
