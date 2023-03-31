/*
Copyright 2023.

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

package controllers

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	samiov1alpha1 "github.com/foreversunyao/simple-k8s-operator/api/v1alpha1"
)

// SamReconciler reconciles a Sam object
type SamReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=sam.io,resources=sams,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=sam.io,resources=sams/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=sam.io,resources=sams/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Sam object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *SamReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	cr := &samiov1alpha1.Sam{}
	err := r.Client.Get(ctx, req.NamespacedName, cr)
	if err != nil {
		return ctrl.Result{}, err
	}

	deployment, err := r.createDeployment(cr, r.samDeployment(cr))
	if err != nil {
		return reconcile.Result{}, err
	}

	// If the spec.Replicas in the CR changes, update the deployment number of replicas
	if deployment.Spec.Replicas != &cr.Spec.Replicas {
		controllerutil.CreateOrUpdate(context.TODO(), r.Client, deployment, func() error {
			deployment.Spec.Replicas = &cr.Spec.Replicas
			return nil
		})
	}

	err = r.createService(cr, r.samService(cr))
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.createRoute(cr, r.samRoute(cr))
	if err != nil {
		return reconcile.Result{}, err
	}
	return ctrl.Result{}, nil
}

func labels(cr *samiov1alpha1.Sam, tier string) map[string]string {
	// Fetches and sets labels

	return map[string]string{
		"app":    "Sam",
		"sam_cr": cr.Name,
		"tier":   tier,
	}
}

// This is the equivalent of creating a deployment yaml and returning it
// It doesn't create anything on cluster
func (r *SamReconciler) samDeployment(cr *samiov1alpha1.Sam) *appsv1.Deployment {
	// Build a Deployment
	labels := labels(cr, "backend-sam")
	size := cr.Spec.Replicas
	samDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sam",
			Namespace: cr.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &size,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           cr.Spec.Image,
						ImagePullPolicy: corev1.PullAlways,
						Name:            "sam-pod",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "sam",
						}},
					}},
				},
			},
		},
	}

	// sets the this controller as owner
	controllerutil.SetControllerReference(cr, samDeployment, r.Scheme)
	return samDeployment
}

// check for a deployment if it doesn't exist it creates one on cluster using the deployment created in deployment
func (r SamReconciler) createDeployment(cr *samiov1alpha1.Sam, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	// check for a deployment in the namespace
	found := &appsv1.Deployment{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: deployment.Name, Namespace: cr.Namespace}, found)
	if err != nil {
		log.Log.Info("Creating Deployment")
		err = r.Client.Create(context.TODO(), deployment)
		if err != nil {
			log.Log.Error(err, "Failed to create deployment")
			return found, err
		}
	}
	return found, nil
}

func (r SamReconciler) samService(cr *samiov1alpha1.Sam) *corev1.Service {
	labels := labels(cr, "backend-sam")

	samService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sam-service",
			Namespace: cr.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       8080,
				TargetPort: intstr.FromInt(8080),
			}},
		},
	}

	controllerutil.SetControllerReference(cr, samService, r.Scheme)
	return samService
}

// check for a service if it doesn't exist it creates one on cluster using the service created in samService
func (r SamReconciler) createService(cr *samiov1alpha1.Sam, samServcie *corev1.Service) error {
	// check for a service in the namespace
	found := &corev1.Service{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: samServcie.Name, Namespace: cr.Namespace}, found)
	if err != nil {
		log.Log.Info("Creating Service")
		err = r.Client.Create(context.TODO(), samServcie)
		if err != nil {
			log.Log.Error(err, "Failed to create Service")
			return err
		}
	}
	return nil
}

// This is the equivalent of creating a route yaml file and returning it
// It doesn't create anything on cluster
func (r SamReconciler) samRoute(cr *samiov1alpha1.Sam) *routev1.Route {
	labels := labels(cr, "backend-sam")

	samRoute := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sam-route",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: "sam-service",
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromInt(8080),
			},
		},
	}
	controllerutil.SetControllerReference(cr, samRoute, r.Scheme)
	return samRoute
}

// check for a route if it doesn't exist it creates one on cluster using the route created in samRoute
func (r SamReconciler) createRoute(cr *samiov1alpha1.Sam, samRoute *routev1.Route) error {
	// check for a route in the namespace
	found := &routev1.Route{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: samRoute.Name, Namespace: cr.Namespace}, found)
	if err != nil {
		log.Log.Info("Creating Route")
		err = r.Client.Create(context.TODO(), samRoute)
		if err != nil {
			log.Log.Error(err, "Failed to create Route")
			return err
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SamReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&samiov1alpha1.Sam{}).
		Complete(r)
}
