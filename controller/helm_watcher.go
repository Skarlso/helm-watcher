package controller

import (
	"context"

	"k8s.io/client-go/tools/reference"

	"helm-watcher/cache"

	"github.com/fluxcd/pkg/runtime/events"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HelmWatcherReconciler runs the reconcile loop for the watcher.
type HelmWatcherReconciler struct {
	client.Client
	Scheme                *runtime.Scheme
	Cache                 *cache.HelmCache // this would be an interface ofc.
	ExternalEventRecorder *events.Recorder
}

// +kubebuilder:rbac:groups=helm.watcher,resources=helmrepositories,verbs=get;list;watch
// +kubebuilder:rbac:groups=helm.watcher,resources=helmrepositories/status,verbs=get

func (r *HelmWatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log, err := logr.FromContext(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	// get source object
	var repository sourcev1.HelmRepository
	if err := r.Get(ctx, req.NamespacedName, &repository); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("found the repository: ", "name", repository.Name)
	if url := r.Cache.Get(repository.Name); url != "" {
		log.Info("already seen this one", "url", url)
		return ctrl.Result{}, nil
	}
	r.Cache.Add(repository.Name, repository.Status.URL)
	log.Info("added this new one", "url", repository.Status.URL)
	r.event(ctx, &repository, "bla", "info", "THIS IS A MESSAGE BWUHAHAHA!")
	return ctrl.Result{}, nil
}

func (r *HelmWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sourcev1.HelmRepository{}, builder.WithPredicates(HelmWatcherReconcilerPredicate{})).
		Complete(r)
}

// event emits a Kubernetes event and forwards the event to notification controller if configured.
func (r *HelmWatcherReconciler) event(ctx context.Context, hr *sourcev1.HelmRepository, revision, severity, msg string) {
	if r.ExternalEventRecorder == nil {
		return
	}

	objRef, err := reference.GetReference(r.Scheme, hr)
	if err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "unable to send event")
		return
	}

	var meta map[string]string
	if revision != "" {
		meta = map[string]string{"revision": revision}
	}
	if err := r.ExternalEventRecorder.Eventf(*objRef, meta, severity, severity, msg); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "unable to send event")
		return
	}
}
