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

	v1beta1 "github.com/thmzlt/hedron/apis/core/v1beta1"
	k8s_corev1 "k8s.io/api/core/v1"
	k8s_metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProjectReconciler reconciles a Project object
type ProjectReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core.hedron.build,resources=projects,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.hedron.build,resources=projects/status,verbs=get;update;patch

func (r *ProjectReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("project", req.NamespacedName)

	var project v1beta1.Project
	var revision v1beta1.Revision

	// Fetch Project resource
	err := r.Get(ctx, req.NamespacedName, &project)
	if err != nil {
		log.Error(err, "Failed to fetch project")

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Create temporary directory
	tempDir, err := ioutil.TempDir("", fmt.Sprintf("hedron-%s-", project.Name))
	if err != nil {
		log.Error(err, "Failed to create temporary directory", "tempDir", tempDir)
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

	// Fetch repository HEAD
	head, err := repo.Head()
	if err != nil {
		log.Error(err, "Failed to fetch repository HEAD")
	}

	// Derive Revision name
	revisionName := fmt.Sprintf("%s-%s", project.Name, head.Hash().String())

	// Fetch Revision with matching revision
	err = r.Get(ctx, client.ObjectKey{
		Namespace: project.Namespace,
		Name:      revisionName,
	}, &revision)
	if err != nil && strings.Contains(err.Error(), "not found") {
		revision = v1beta1.Revision{
			ObjectMeta: k8s_metav1.ObjectMeta{
				Namespace: project.Namespace,
				Name:      revisionName,
			},
			Spec: v1beta1.RevisionSpec{
				ProjectRef: k8s_corev1.LocalObjectReference{Name: project.Name},
				Revision:   head.Hash().String(),
			},
			Status: v1beta1.RevisionStatus{
				State: "Pending",
			},
		}
		err = r.Create(ctx, &revision)
		if err != nil {
			log.Error(err, "Failed to create revision")
		}
		log.Info("Created revision", "revisionName", revision.Name)

		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to fetch revision")

		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *ProjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Project{}).
		Complete(r)
}
