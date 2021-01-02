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
	"k8s.io/apimachinery/pkg/types"
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

type contextKey string

var (
	contextKeyLog      = contextKey("log")
	contextKeyRequest  = contextKey("request")
	contextKeyRevision = contextKey("revision")
)

// +kubebuilder:rbac:groups=core.hedron.build,resources=revisions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.hedron.build,resources=revisions/status,verbs=get;update;patch

func (r *RevisionReconciler) Reconcile(request ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("revision", request.NamespacedName)
	ctx := context.WithValue(context.Background(), contextKeyLog, log)

	requestCtx := context.WithValue(ctx, contextKeyRequest, request)

	revision, err := r.fetchRevision(requestCtx)
	if err != nil {
		log.Error(err, "Failed to fetch revision")

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	revisionCtx := context.WithValue(requestCtx, contextKeyRevision, revision)

	if revision.Status.State == "Failed" || revision.Status.State == "Succeeded" {
		log.Info(fmt.Sprintf("Revision is %s", strings.ToLower(string(revision.Status.State))))
	}

	job, err := r.fetchJob(revisionCtx)
	if err != nil {
		job, err = r.createJob(revisionCtx)
		if err != nil {
			log.Error(err, "Failed to create job")
		}
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

	if err = r.Update(ctx, &revision); err != nil {
		log.Error(err, "Failed to update revision state")
	}

	return ctrl.Result{}, nil
}

func (r *RevisionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ownerKey := ".metadata.controller"
	apiGVStr := v1beta1.GroupVersion.String()
	log := r.Log.WithName("setup")

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
		Complete(r)
}

func (r *RevisionReconciler) createJob(ctx context.Context) (batchv1.Job, error) {
	revision := ctx.Value(contextKeyRevision).(v1beta1.Revision)

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-job", revision.Name),
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
	jobName := types.NamespacedName{
		Namespace: request.NamespacedName.Namespace,
		Name:      fmt.Sprintf("%s-job", request.NamespacedName.Name),
	}

	return job, r.Get(ctx, jobName, &job)
}

func (r *RevisionReconciler) fetchRevision(ctx context.Context) (v1beta1.Revision, error) {
	var revision v1beta1.Revision

	request := ctx.Value(contextKeyRequest).(ctrl.Request)
	return revision, r.Get(ctx, request.NamespacedName, &revision)
}
