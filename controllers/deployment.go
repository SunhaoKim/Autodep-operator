package controllers

import (
	"context"
	appsv1alpha1 "init_rollout_operator/api/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	requestcpu         string
	requestmemory      string
	limitcpu           string
	deprs              int32
	limitmemory        string
	DepimagePullPolicy string
)

func (r *AutodepReconciler) CreateDeploymentForAutodep(ctx context.Context, dep *appsv1alpha1.Autodep) error {
	depname := getDepName(dep)
	r.Log.Info("Create new deployment For Auto Dep", dep.Namespace, dep.Name, "Deployment Name", depname)
	deployment, err := r.DeploymentForbackend(dep)
	if err != nil {
		r.Log.Error(err, "Failed create autodep deployment resource", depname)
		return err
	}
	err = r.Create(ctx, deployment)
	if err != nil {
		r.Log.Error(err, "failed create deployment %s", depname)
		return err
	}
	r.Log.Info("create autodep deployment success", dep.Namespace, depname)
	return nil
}
func (r *AutodepReconciler) UpdateDeploymentForAutodep(ctx context.Context, dep *appsv1alpha1.Autodep) error {
	depname := getDepName(dep)
	r.Log.Info("get deployment For Auto Dep Just update", dep.Namespace, dep.Name, "Deployment Name", depname)
	deployment, err := r.DeploymentForbackend(dep)
	if err != nil {
		r.Log.Error(err, "Failed update autodep deployment resource", depname)
		return err
	}
	err = r.Update(ctx, deployment)
	if err != nil {
		r.Log.Error(err, "failed update deployment %s", depname)
		return err
	}
	r.Log.Info("update autodep deployment success", dep.Namespace, depname)
	return nil
}
func (r *AutodepReconciler) DeploymentForbackend(dep *appsv1alpha1.Autodep) (*appsv1.Deployment, error) {
	depname := getDepName(dep)
	switch dep.Spec.Depenv {
	case "dev":
		DepimagePullPolicy = "Always"
		deprs = 1
		requestcpu = "50m"
		requestmemory = "100Mi"
		limitcpu = "100m"
		limitmemory = "200Mi"
	case "prod":
		DepimagePullPolicy = "IfNotPresent"
		deprs = 3
		requestcpu = "1000m"
		requestmemory = "500Mi"
		limitcpu = "1000m"
		limitmemory = "500Mi"
	default:
		DepimagePullPolicy = "Always"
		deprs = 1
		requestcpu = "50m"
		requestmemory = "100Mi"
		limitcpu = "200m"
		limitmemory = "300Mi"
	}
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: dep.Namespace,
			Name:      depname,
			Labels: map[string]string{
				"app": depname,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &deprs,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": depname,
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: int32(5),
					},
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: int32(1),
					},
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": depname,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Env: []corev1.EnvVar{
								{
									Name:  "app_env",
									Value: dep.Spec.Depenv,
								},
							},
							Name:            depname,
							Image:           dep.Spec.Depimage,
							ImagePullPolicy: corev1.PullPolicy(DepimagePullPolicy),
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: dep.Spec.SvcPort,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									"cpu":    resource.MustParse(limitcpu),
									"memory": resource.MustParse(limitmemory),
								},
								Requests: corev1.ResourceList{
									"cpu":    resource.MustParse(requestcpu),
									"memory": resource.MustParse(requestmemory),
								},
							},
							Stdin: true,
							TTY:   true,
						},
					},
					DNSPolicy: corev1.DNSClusterFirst,
					ImagePullSecrets: []corev1.LocalObjectReference{
						{
							Name: dep.Spec.DepimagePullSecret,
						},
					},
					RestartPolicy: corev1.RestartPolicyAlways,
				},
			},
		},
	}
	if err := controllerutil.SetControllerReference(dep, deployment, r.Scheme); err != nil {
		return nil, err
	}
	return deployment, nil
}
