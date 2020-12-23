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

	corev1beta1 "github.com/thmzlt/hedron/apis/core/v1beta1"
)

// JobReconciler reconciles a Job object
type JobReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core.hedron.build,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.hedron.build,resources=jobs/status,verbs=get;update;patch

func (r *JobReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	log := r.Log.WithValues("job", req.NamespacedName)

	log.Info("Is the job running?")

	return ctrl.Result{}, nil
}

func (r *JobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1beta1.Job{}).
		Complete(r)
}
