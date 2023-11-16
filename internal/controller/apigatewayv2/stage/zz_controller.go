// SPDX-FileCopyrightText: 2023 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

/*
Copyright 2022 Upbound Inc.
*/

// Code generated by upjet. DO NOT EDIT.

package stage

import (
	"time"

	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	xpresource "github.com/crossplane/crossplane-runtime/pkg/resource"
	tjcontroller "github.com/crossplane/upjet/pkg/controller"
	"github.com/crossplane/upjet/pkg/controller/handler"
	"github.com/crossplane/upjet/pkg/metrics"
	ctrl "sigs.k8s.io/controller-runtime"

	v1beta1 "github.com/upbound/provider-aws/apis/apigatewayv2/v1beta1"
	features "github.com/upbound/provider-aws/internal/features"
)

// Setup adds a controller that reconciles Stage managed resources.
func Setup(mgr ctrl.Manager, o tjcontroller.Options) error {
	name := managed.ControllerName(v1beta1.Stage_GroupVersionKind.String())
	var initializers managed.InitializerChain
	for _, i := range o.Provider.Resources["aws_apigatewayv2_stage"].InitializerFns {
		initializers = append(initializers, i(mgr.GetClient()))
	}
	initializers = append(initializers, managed.NewNameAsExternalName(mgr.GetClient()))
	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.SecretStoreConfigGVK != nil {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), *o.SecretStoreConfigGVK, connection.WithTLSConfig(o.ESSOptions.TLSConfig)))
	}
	eventHandler := handler.NewEventHandler(handler.WithLogger(o.Logger.WithValues("gvk", v1beta1.Stage_GroupVersionKind)))
	ac := tjcontroller.NewAPICallbacks(mgr, xpresource.ManagedKind(v1beta1.Stage_GroupVersionKind), tjcontroller.WithEventHandler(eventHandler), tjcontroller.WithStatusUpdates(false))
	opts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(
			tjcontroller.NewNoForkAsyncConnector(mgr.GetClient(), o.OperationTrackerStore, o.SetupFn, o.Provider.Resources["aws_apigatewayv2_stage"],
				tjcontroller.WithNoForkAsyncLogger(o.Logger),
				tjcontroller.WithNoForkAsyncConnectorEventHandler(eventHandler),
				tjcontroller.WithNoForkAsyncCallbackProvider(ac),
				tjcontroller.WithNoForkAsyncMetricRecorder(metrics.NewMetricRecorder(v1beta1.Stage_GroupVersionKind, mgr, o.PollInterval)),
				tjcontroller.WithNoForkAsyncManagementPolicies(o.Features.Enabled(features.EnableBetaManagementPolicies)))),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithFinalizer(tjcontroller.NewNoForkFinalizer(o.OperationTrackerStore, xpresource.NewAPIFinalizer(mgr.GetClient(), managed.FinalizerName))),
		managed.WithTimeout(3 * time.Minute),
		managed.WithInitializers(initializers),
		managed.WithConnectionPublishers(cps...),
		managed.WithPollInterval(o.PollInterval),
	}
	if o.PollJitter != 0 {
		opts = append(opts, managed.WithPollJitterHook(o.PollJitter))
	}
	if o.Features.Enabled(features.EnableBetaManagementPolicies) {
		opts = append(opts, managed.WithManagementPolicies())
	}
	r := managed.NewReconciler(mgr, xpresource.ManagedKind(v1beta1.Stage_GroupVersionKind), opts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(xpresource.DesiredStateChanged()).
		Watches(&v1beta1.Stage{}, eventHandler).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}
