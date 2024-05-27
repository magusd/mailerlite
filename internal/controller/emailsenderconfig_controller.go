/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	k8smagusdcomv1 "k8s.magusd.com/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// EmailSenderConfigReconciler reconciles a EmailSenderConfig object
type EmailSenderConfigReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=k8s.magusd.com.my.domain,resources=emailsenderconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=k8s.magusd.com.my.domain,resources=emailsenderconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=k8s.magusd.com.my.domain,resources=emailsenderconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the EmailSenderConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *EmailSenderConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Reconciling sender config")

	var emailConfig k8smagusdcomv1.EmailSenderConfig
	if err := r.Get(ctx, req.NamespacedName, &emailConfig); err != nil {
		logger.Error(err, "unable to fetch sender config")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if emailConfig.Generation == 1 {
		logger.Info("sender config created")
	} else {
		logger.Info("sender config updated")
	}

	//secretConfig := &corev1.Secret{}
	//secretRef := types.NamespacedName{
	//	Namespace: req.NamespacedName.Namespace,
	//	Name:      emailConfig.Spec.ApiTokenSecretRef,
	//}

	//if err := r.Client.Get(ctx, secretRef, secretConfig); err != nil {
	//	logger.Error(err, "unable to find secret apiTokenSecretRef in this namespace")
	//	r.Recorder.Event(&emailConfig, "Warning", "NotFound",
	//		fmt.Sprintf("can't find secret %s", emailConfig.Spec.ApiTokenSecretRef))
	//	return ctrl.Result{}, client.IgnoreNotFound(err)
	//}
	//
	//for key, value := range secretConfig.Data {
	//	fmt.Printf("Key: %s, Value: %s\n", key, string(value))
	//}

	//validate credentials

	//ctrl.SetControllerReference(job, pod, r.Scheme)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EmailSenderConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8smagusdcomv1.EmailSenderConfig{}).
		Complete(r)
}
