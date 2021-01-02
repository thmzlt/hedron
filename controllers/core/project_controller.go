/*
Unlicensed
*/

package controllers

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/thmzlt/hedron/apis/core/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProjectReconciler reconciles a Project object
type ProjectReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core.hedron.build,resources=projects,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.hedron.build,resources=projects/status,verbs=get;update;patch

func (r *ProjectReconciler) Reconcile(request ctrl.Request) (ctrl.Result, error) {
	requestCtx := context.WithValue(context.Background(), contextKeyRequest, request)

	project, err := r.fetchProject(requestCtx)
	if err != nil && strings.Contains(err.Error(), "not found") {
		r.Log.Info("Project no longer exists", "project", request.NamespacedName)
	} else if err != nil {
		r.Log.Error(err, "Failed to fetch project")

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	projectCtx := context.WithValue(requestCtx, contextKeyProject, project)

	head, err := r.getRepoHead(projectCtx)
	if err != nil {
		r.Log.Error(err, "Failed to get repository HEAD")
	}

	_, err = r.fetchRevision(projectCtx, head.Hash())
	if err != nil && strings.Contains(err.Error(), "not found") {
		_, err = r.createRevision(projectCtx, head.Hash())
		if err != nil {
			r.Log.Error(err, "Failed to create revision")
		}
	} else if err != nil {
		r.Log.Error(err, "Failed to fetch revision")
	}

	return ctrl.Result{}, nil
}

func (r *ProjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ownerKey := ".metadata.controller"
	apiGVStr := v1beta1.GroupVersion.String()

	if err := mgr.GetFieldIndexer().IndexField(&v1beta1.Revision{}, ownerKey, func(object runtime.Object) []string {
		revision := object.(*v1beta1.Revision)
		owner := metav1.GetControllerOf(revision)

		// Not a match (not controlled at all)
		if owner == nil {
			return nil
		}
		// Not a match (not controlled by a project)
		if owner.APIVersion != apiGVStr || owner.Kind != "Project" {
			return nil
		}
		// Match
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Project{}).
		Owns(&v1beta1.Revision{}).
		Complete(r)
}

func (r *ProjectReconciler) createRevision(ctx context.Context, gitHash plumbing.Hash) (v1beta1.Revision, error) {
	project := ctx.Value(contextKeyProject).(v1beta1.Project)
	revisionName := fmt.Sprintf("%s-%s", project.Name, gitHash.String())

	revision := v1beta1.Revision{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: project.Namespace,
			Name:      revisionName,
		},
		Spec: v1beta1.RevisionSpec{
			ProjectRef: corev1.LocalObjectReference{Name: project.Name},
			Revision:   gitHash.String(),
		},
		Status: v1beta1.RevisionStatus{
			State: "Pending",
		},
	}

	if err := ctrl.SetControllerReference(&project, &revision, r.Scheme); err != nil {
		return revision, err
	}

	return revision, r.Create(ctx, &revision)
}

func (r *ProjectReconciler) fetchProject(ctx context.Context) (v1beta1.Project, error) {
	var project v1beta1.Project

	request := ctx.Value(contextKeyRequest).(ctrl.Request)

	return project, r.Get(ctx, request.NamespacedName, &project)
}

func (r *ProjectReconciler) fetchRevision(ctx context.Context, gitHash plumbing.Hash) (v1beta1.Revision, error) {
	var revision v1beta1.Revision

	project := ctx.Value(contextKeyProject).(v1beta1.Project)
	revisionName := fmt.Sprintf("%s-%s", project.Name, gitHash.String())

	return revision, r.Get(ctx, client.ObjectKey{
		Namespace: project.Namespace,
		Name:      revisionName,
	}, &revision)
}

func (r *ProjectReconciler) getRepoHead(ctx context.Context) (*plumbing.Reference, error) {
	project := ctx.Value(contextKeyProject).(v1beta1.Project)

	// Create temporary directory
	tempDir, err := ioutil.TempDir("", fmt.Sprintf("hedron-%s-", project.Name))
	if err != nil {
		r.Log.Error(err, "Failed to create temporary directory", "tempDir", tempDir)
	}
	defer os.RemoveAll(tempDir)

	// Clone project repository
	repo, err := git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:           project.Spec.Repository.URL,
		ReferenceName: plumbing.ReferenceName(project.Spec.Repository.Ref),
	})
	if err != nil {
		r.Log.Error(err, "Failed to clone repository")
	}

	return repo.Head()
}
