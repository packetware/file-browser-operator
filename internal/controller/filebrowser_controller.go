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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	//filebrowserv1 "github.com/packetware/file-browser-operator/api/v1"
)

// FileBrowserReconciler reconciles a FileBrowser object
type FileBrowserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=filebrowser.packetware.net,resources=filebrowsers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=filebrowser.packetware.net,resources=filebrowsers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=filebrowser.packetware.net,resources=filebrowsers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=persistentvolumes,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the FileBrowser object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *FileBrowserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the PVC instance
	pvc := &corev1.PersistentVolumeClaim{}
	if err := r.Get(ctx, req.NamespacedName, pvc); err != nil {
		if errors.IsNotFound(err) {
			// PVC not found, might be deleted
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch PersistentVolumeClaim")
		return ctrl.Result{}, err
	}

	// Define a new Deployment object
	dep := r.deploymentForPVC(pvc)

	// Check if the Deployment already exists
	found := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Define a new Service object
	svc := r.serviceForPVC(pvc)

	// Check if the Service already exists
	foundSvc := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, foundSvc)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
		err = r.Create(ctx, svc)
		if err != nil {
			log.Error(err, "Failed to create new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
			return ctrl.Result{}, err
		}
		// Service created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Service")
		return ctrl.Result{}, err
	}

	/*
		// Ensure Deployment for Filebrowser
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      filebrowser.Name + "-deployment",
				Namespace: req.Namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(1),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": filebrowser.Name,
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": filebrowser.Name,
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "filesapi",
								Image: "ultrasive/filesapi:v1.0.0", // Example image
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "data",
										MountPath: "/srv/filebrowser",
									},
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "data",
								VolumeSource: corev1.VolumeSource{
									PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
										ClaimName: filebrowser.Spec.PersistentVolumeClaimName,
									},
								},
							},
						},
					},
				},
			},
		}
		if err := controllerutil.SetControllerReference(filebrowser, deployment, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		// Check if Deployment exists, create if not
		foundDeployment := &appsv1.Deployment{}
		err = r.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, foundDeployment)
		if err != nil && errors.IsNotFound(err) {
			log.Info("Creating Deployment", "Namespace", deployment.Namespace, "Name", deployment.Name)
			err = r.Create(ctx, deployment)
			if err != nil {
				return ctrl.Result{}, err
			}
		} else if err != nil {
			return ctrl.Result{}, err
		}

		// Ensure NodePort Service for Filebrowser
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      filebrowser.Name + "-service",
				Namespace: req.Namespace,
			},
			Spec: corev1.ServiceSpec{
				Selector: deployment.Spec.Selector.MatchLabels,
				Ports: []corev1.ServicePort{
					{
						Name:       "http",
						Protocol:   corev1.ProtocolTCP,
						Port:       80,
						TargetPort: intstr.FromInt(80),
						NodePort:   30000, // Adjust NodePort as needed
					},
				},
				Type: corev1.ServiceTypeNodePort,
			},
		}
		if err := controllerutil.SetControllerReference(filebrowser, service, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		// Check if Service exists, create if not
		foundService := &corev1.Service{}
		err = r.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, foundService)
		if err != nil && errors.IsNotFound(err) {
			log.Info("Creating Service", "Namespace", service.Namespace, "Name", service.Name)
			err = r.Create(ctx, service)
			if err != nil {
				return ctrl.Result{}, err
			}
		} else if err != nil {
			return ctrl.Result{}, err
		}*/

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FileBrowserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		//For(&filebrowserv1.FileBrowser{}).
		For(&corev1.PersistentVolumeClaim{}).
		//Owns(&corev1.PersistentVolumeClaim{}).
		//Owns(&appsv1.Deployment{}).
		//Owns(&corev1.Service{}).
		Complete(r)
}

// Helpers
func (r *FileBrowserReconciler) deploymentForPVC(pvc *corev1.PersistentVolumeClaim) *appsv1.Deployment {
	labels := map[string]string{
		"app": "filebrowser",
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvc.Name + "-filebrowser",
			Namespace: pvc.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "filesapi",
						Image: "ultrasive/filesapi:v1.0.0",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "data",
							MountPath: "/srv",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: "data",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: pvc.Name,
							},
						},
					}},
				},
			},
		},
	}
	// Set PVC instance as the owner and controller
	ctrl.SetControllerReference(pvc, dep, r.Scheme)
	return dep
}

func (r *FileBrowserReconciler) serviceForPVC(pvc *corev1.PersistentVolumeClaim) *corev1.Service {
	labels := map[string]string{
		"app": "filebrowser",
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvc.Name + "-filebrowser",
			Namespace: pvc.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Port: 8080,
				//TargetPort: intstr.FromInt(8080),
				Protocol: corev1.ProtocolTCP,
			}},
			Selector: labels,
			Type:     corev1.ServiceTypeNodePort,
		},
	}
	// Set PVC instance as the owner and controller
	ctrl.SetControllerReference(pvc, svc, r.Scheme)
	return svc
}

func int32Ptr(i int32) *int32 { return &i }
