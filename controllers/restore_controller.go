/*
 * restore_controller.go
 *
 * This source file is part of the FoundationDB open source project
 *
 * Copyright 2020-2021 Apple Inc. and the FoundationDB project authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controllers

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/FoundationDB/fdb-kubernetes-operator/pkg/fdbadminclient"

	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	fdbtypes "github.com/FoundationDB/fdb-kubernetes-operator/api/v1beta1"
	"github.com/go-logr/logr"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FoundationDBRestoreReconciler reconciles a FoundationDBRestore object
type FoundationDBRestoreReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Log      logr.Logger

	DatabaseClientProvider DatabaseClientProvider
}

// +kubebuilder:rbac:groups=apps.foundationdb.org,resources=foundationdbrestores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.foundationdb.org,resources=foundationdbrestores/status,verbs=get;update;patch

// Reconcile runs the reconciliation logic.
func (r *FoundationDBRestoreReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	restore := &fdbtypes.FoundationDBRestore{}
	err := r.Get(ctx, request.NamespacedName, restore)

	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	restoreLog := log.WithValues("namespace", restore.Namespace, "restore", restore.Name)

	subReconcilers := []restoreSubReconciler{
		startRestore{},
	}

	for _, subReconciler := range subReconcilers {
		requeue := subReconciler.reconcile(ctx, r, restore)
		if requeue == nil {
			continue
		}

		return processRequeue(requeue, subReconciler, restore, r.Recorder, restoreLog)
	}

	restoreLog.Info("Reconciliation complete")

	return ctrl.Result{}, nil
}

// getDatabaseClientProvider gets the client provider for a reconciler.
func (r *FoundationDBRestoreReconciler) getDatabaseClientProvider() DatabaseClientProvider {
	if r.DatabaseClientProvider != nil {
		return r.DatabaseClientProvider
	}
	panic("Restore reconciler does not have a DatabaseClientProvider defined")
}

// adminClientForRestore provides an admin client for a restore reconciler.
func (r *FoundationDBRestoreReconciler) adminClientForRestore(ctx context.Context, restore *fdbtypes.FoundationDBRestore) (fdbadminclient.AdminClient, error) {
	cluster := &fdbtypes.FoundationDBCluster{}
	err := r.Get(ctx, types.NamespacedName{Namespace: restore.ObjectMeta.Namespace, Name: restore.Spec.DestinationClusterName}, cluster)
	if err != nil {
		return nil, err
	}

	adminClient, err := r.getDatabaseClientProvider().GetAdminClient(cluster, r)
	if err != nil {
		return nil, err
	}

	adminClient.SetKnobs(restore.Spec.CustomParameters.GetKnobsForCLI())

	return adminClient, nil
}

// SetupWithManager prepares a reconciler for use.
func (r *FoundationDBRestoreReconciler) SetupWithManager(mgr ctrl.Manager, maxConcurrentReconciles int, selector metav1.LabelSelector) error {
	labelSelectorPredicate, err := predicate.LabelSelectorPredicate(selector)
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: maxConcurrentReconciles},
		).
		For(&fdbtypes.FoundationDBRestore{}).
		// Only react on generation changes or annotation changes and only watch
		// resources with the provided label selector.
		WithEventFilter(
			predicate.And(
				labelSelectorPredicate,
				predicate.Or(
					predicate.GenerationChangedPredicate{},
					predicate.AnnotationChangedPredicate{},
				),
			)).
		Complete(r)
}

// restoreSubReconciler describes a class that does part of the work of
// reconciliation for a restore.
type restoreSubReconciler interface {
	/**
	reconcile runs the reconciler's work.

	If reconciliation can continue, this should return nil.

	If reconciliation encounters an error, this should return a `requeue` object
	with an `Error` field.

	If reconciliation cannot proceed, this should return a `requeue` object with
	a `Message` field.
	*/
	reconcile(ctx context.Context, r *FoundationDBRestoreReconciler, restore *fdbtypes.FoundationDBRestore) *requeue
}
