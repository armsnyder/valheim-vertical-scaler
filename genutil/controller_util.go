package genutil

import (
	"context"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func DoNotRequeue() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func RequeueWithError(ctx context.Context, err error) (reconcile.Result, error) {
	if err != nil {
		logr.FromContext(ctx).Error(err, "re-enqueuing after error")
	}
	return reconcile.Result{}, err
}
