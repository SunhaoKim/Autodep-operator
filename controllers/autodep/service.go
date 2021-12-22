package autodep

import (
	"context"
	appsv1alpha1 "init_rollout_operator/api/v1alpha1"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *AutodepReconciler) CreateBackendService(ctx context.Context, svc *appsv1alpha1.Autodep) error {
	svcname := GetBackendName(svc)
	log.WithField("service name", svcname).WithField("Namespace", svc.Namespace).Info("Create new backend Service")
	service, err := r.ServiceForBackend(svc)
	if err != nil {
		log.WithField("service name", svcname).Error(err, "Failed get backend service resource")
		return err
	}
	err = r.Create(ctx, service)
	if err != nil {
		log.WithField("service name", svcname).Error(err, "Failed Create backend service ")
		return err
	}
	log.WithField("service name", svcname).WithField("Namespace", svc.Namespace).Info("create  backend service success")
	return nil
}

func (r *AutodepReconciler) UpdateBackendService(ctx context.Context, svc *appsv1alpha1.Autodep) error {
	svcname := GetBackendName(svc)
	log.WithField("service name", svcname).WithField("Namespace", svc.Namespace).Info("Get backend service just update")
	service, err := r.ServiceForBackend(svc)
	if err != nil {
		log.WithField("service name", svcname).Error(err, "Failed get  backend service resource")
		return err
	}
	err = r.Patch(ctx, service, client.Merge)
	if err != nil {
		log.WithField("service name", svcname).Error(err, "failed update backend service ")
		return err
	}
	log.WithField("service name", svcname).WithField("Namespace", svc.Namespace).Info("update backend service success")
	return nil
}

func (r *AutodepReconciler) ServiceForBackend(autodepsvc *appsv1alpha1.Autodep) (*corev1.Service, error) {
	svcname := GetBackendName(autodepsvc)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcname,
			Namespace: autodepsvc.Namespace,
			Labels: map[string]string{
				"app": svcname,
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:     "backend",
					Protocol: corev1.ProtocolTCP,
					Port:     autodepsvc.Spec.SvcPort,
				},
			},
			Selector: map[string]string{
				"app": autodepsvc.Name,
			},
		},
	}
	err := controllerutil.SetControllerReference(autodepsvc, service, r.Scheme)
	if err != nil {
		return nil, err
	}
	return service, err
}
