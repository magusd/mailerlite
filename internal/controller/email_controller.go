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
	"errors"
	"fmt"
	"github.com/mailersend/mailersend-go"
	"github.com/mailgun/mailgun-go/v4"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	k8smagusdcomv1 "k8s.magusd.com/api/v1"
)

// EmailReconciler reconciles a Email object
type EmailReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=k8s.magusd.com.my.domain,resources=emails,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=k8s.magusd.com.my.domain,resources=emails/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=k8s.magusd.com.my.domain,resources=emails/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Email object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *EmailReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var email k8smagusdcomv1.Email
	if err := r.Get(ctx, req.NamespacedName, &email); err != nil {
		logger.Error(err, "unable to fetch email")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	switch email.Status.DeliveryStatus {
	case k8smagusdcomv1.EmailSentStatus:
		return ctrl.Result{}, nil
	case k8smagusdcomv1.EmailErrorStatus:
		//todo handle error types, retry policy, etc
		return ctrl.Result{}, nil
	case "":
		email.Status.DeliveryStatus = k8smagusdcomv1.EmailPendingStatus
		r.Status().Update(ctx, &email)
		r.Recorder.Event(&email, "Normal", "Queued", "Email queued")
	}

	emailConfig := k8smagusdcomv1.EmailSenderConfig{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      email.Spec.SenderConfigRef,
	}, &emailConfig)

	if err != nil {
		logger.Error(err, "Could not find emailSenderConfig")
		r.Recorder.Event(&email, "Warning", "Validation",
			fmt.Sprintf("can't find EmailSenderConfig %s", email.Spec.SenderConfigRef))
		return ctrl.Result{}, err
	}

	logger.Info("Sending email " + req.NamespacedName.String())
	id, err := r.SendEmail(ctx, email, emailConfig)
	if err != nil {
		logger.Info("Error sending email " + err.Error())
		email.Status.DeliveryStatus = k8smagusdcomv1.EmailErrorStatus
		email.Status.Error = err.Error()
		r.Recorder.Event(&email, "Warning", "Failed", err.Error())
		r.Status().Update(ctx, &email)
	} else {
		logger.Info("Email sent" + req.NamespacedName.String())
		email.Status.DeliveryStatus = k8smagusdcomv1.EmailSentStatus
		email.Status.MessageId = id
		r.Recorder.Event(&email, "Normal", "Send", "Email sent")
		r.Status().Update(ctx, &email)
	}
	return ctrl.Result{}, err
}

type MailgunCredentials struct {
	token  string
	domain string
}

func (r *EmailReconciler) SendEmail(ctx context.Context, email k8smagusdcomv1.Email, emailConfig k8smagusdcomv1.EmailSenderConfig) (string, error) {
	secret := corev1.Secret{}
	namespace := emailConfig.Namespace
	if namespace == "" {
		namespace = "default"
	}
	objKey := types.NamespacedName{
		Namespace: namespace,
		Name:      emailConfig.Spec.ApiTokenSecretRef,
	}

	err := r.Get(ctx, objKey, &secret)
	if err != nil {
		return "", err
	}

	provider, ok := secret.Data["provider"]
	if !ok {
		return "", errors.New("Missing required provider key in referenced secret " + email.Spec.SenderConfigRef)
	}

	switch string(provider) {
	case "mailgun":
		return r.SendMailgunEmail(email, emailConfig, secret)
	case "mailsend":

		return r.SendMailSendEmail(email, emailConfig, secret)
	default:
		return "", errors.New(fmt.Sprintf("Invalid provider, %s is not supported", provider))
	}
}

func (r *EmailReconciler) SendMailgunEmail(email k8smagusdcomv1.Email, emailConfig k8smagusdcomv1.EmailSenderConfig, secret corev1.Secret) (string, error) {
	mg := mailgun.NewMailgun(string(secret.Data["domain"]), string(secret.Data["token"]))
	sender := emailConfig.Spec.SenderEmail
	subject := email.Spec.Subject
	body := email.Spec.Body
	recipient := email.Spec.RecipientEmail

	// The message object allows you to add attachments and Bcc recipients
	message := mg.NewMessage(sender, subject, body, recipient)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message with a 10 second timeout
	_, id, err := mg.Send(ctx, message)

	if err != nil {
		return "", err
	}

	id = strings.Replace(id, "<", "", 1)
	id = strings.Replace(id, ">", "", 1)

	return id, nil
}

func (r *EmailReconciler) SendMailSendEmail(email k8smagusdcomv1.Email, emailConfig k8smagusdcomv1.EmailSenderConfig, secret corev1.Secret) (string, error) {
	var APIToken = string(secret.Data["token"])
	ms := mailersend.NewMailersend(APIToken)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	subject := email.Spec.Subject
	text := email.Spec.Body
	html := email.Spec.Body

	from := mailersend.From{
		Name:  "",
		Email: emailConfig.Spec.SenderEmail,
	}

	recipients := []mailersend.Recipient{
		{
			Name:  "",
			Email: email.Spec.RecipientEmail,
		},
	}

	message := ms.Email.NewMessage()

	message.SetFrom(from)
	message.SetRecipients(recipients)
	message.SetSubject(subject)
	message.SetHTML(html)
	message.SetText(text)

	res, err := ms.Email.Send(ctx, message)

	if err != nil {
		return "", err
	}
	return res.Header.Get("X-Message-Id"), nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EmailReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8smagusdcomv1.Email{}).
		Complete(r)
}
