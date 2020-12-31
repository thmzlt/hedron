/*
Unlicensed
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"
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

func (r *RevisionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("revision", req.NamespacedName)

	var revision v1beta1.Revision

	// Fetch Revision resource
	err := r.Get(ctx, req.NamespacedName, &revision)
	if err != nil {
		log.Info("Failed to fetch revision")

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	switch revision.Status.State {
	case "Pending":
		return r.ReconcilePending(log, revision)
	case "Building":
		log.Info("Revision is building")
	case "Ready":
		log.Info("Revision is ready")
	case "Failed":
		log.Info("Revision is failed")
	}

	return ctrl.Result{}, nil
}

func (r *RevisionReconciler) ReconcilePending(log logr.Logger, revision v1beta1.Revision) (ctrl.Result, error) {
	log.Info("Revision is pending")

	return ctrl.Result{}, nil
}

func (r *RevisionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Revision{}).
		Complete(r)
}
