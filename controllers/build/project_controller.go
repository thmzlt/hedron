/*
Unlicensed
*/

package controllers

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	buildv1beta1 "github.com/thmzlt/hedron/apis/build/v1beta1"
)

// ProjectReconciler reconciles a Project object
type ProjectReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=build.hedron-ci.org,resources=projects,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=build.hedron-ci.org,resources=projects/status,verbs=get;update;patch

func (r *ProjectReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("project", req.NamespacedName)

	var project buildv1beta1.Project

	// Fetch the Project resource
	err := r.Get(ctx, req.NamespacedName, &project)
	if err != nil {
		log.Error(err, "Unable to fetch project")

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Create temporary directory
	tempDir, err := ioutil.TempDir("", fmt.Sprintf("hedron-%s-", project.Name))
	if err != nil {
		log.Error(err, "Error creating temporary directory", "tempDir", tempDir)
	}
	defer os.RemoveAll(tempDir)

	// Clone project repository
	repo, err := git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:           project.Spec.Repository.URL,
		ReferenceName: plumbing.ReferenceName(project.Spec.Repository.Ref),
	})
	if err != nil {
		log.Error(err, "Failed to clone repository")
	}

	// Fetch HEAD
	ref, err := repo.Head()
	if err != nil {
		log.Error(err, "Failed to fetch repository HEAD")
	}

	log.Info("Repository is ready", "revision", ref.Hash().String())

	return ctrl.Result{}, nil
}

func (r *ProjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&buildv1beta1.Project{}).
		Complete(r)
}
