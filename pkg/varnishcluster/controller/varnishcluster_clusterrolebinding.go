package controller

import (
	"context"
	icmapiv1alpha1 "icm-varnish-k8s-operator/api/v1alpha1"
	"icm-varnish-k8s-operator/pkg/labels"
	"icm-varnish-k8s-operator/pkg/logger"
	"icm-varnish-k8s-operator/pkg/names"
	"icm-varnish-k8s-operator/pkg/varnishcluster/compare"

	"github.com/pkg/errors"

	rbac "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileVarnishCluster) reconcileClusterRoleBinding(ctx context.Context, instance *icmapiv1alpha1.VarnishCluster) error {
	clusterRoleBinding := &rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   names.ClusterRoleBinding(instance.Name, instance.Namespace),
			Labels: labels.CombinedComponentLabels(instance, icmapiv1alpha1.VarnishComponentClusterRoleBinding),
			Annotations: map[string]string{
				annotationVarnishClusterNamespace: instance.Namespace,
				annotationVarnishClusterName:      instance.Name,
			},
		},
		Subjects: []rbac.Subject{
			{
				Kind:      rbac.ServiceAccountKind,
				Name:      names.ServiceAccount(instance.Name),
				Namespace: instance.Namespace,
			},
		},
		RoleRef: rbac.RoleRef{
			Kind:     "ClusterRole",
			Name:     names.ClusterRole(instance.Name, instance.Namespace),
			APIGroup: rbac.GroupName,
		},
	}

	logr := logger.FromContext(ctx).With(logger.FieldComponent, icmapiv1alpha1.VarnishComponentClusterRoleBinding)
	logr = logr.With(logger.FieldComponentName, clusterRoleBinding.Name)

	found := &rbac.ClusterRoleBinding{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: clusterRoleBinding.Name, Namespace: clusterRoleBinding.Namespace}, found)
	// If the role does not exist, create it
	// Else if there was a problem doing the GET, just return
	// Else if the clusterRoleBinding exists, and it is different, update
	// Else no changes, do nothing
	if err != nil && kerrors.IsNotFound(err) {
		logr.Infoc("Creating ClusterRoleBinding", "new", clusterRoleBinding)
		if err = r.Create(ctx, clusterRoleBinding); err != nil {
			return errors.Wrap(err, "Unable to create ClusterRoleBinding")
		}
	} else if err != nil {
		return errors.Wrap(err, "Could not Get ClusterRoleBinding")
	} else if !compare.EqualClusterRoleBinding(found, clusterRoleBinding) {
		logr.Infoc("Updating ClusterRoleBinding", "diff", compare.DiffClusterRoleBinding(found, clusterRoleBinding))
		found.Subjects = clusterRoleBinding.Subjects
		found.RoleRef = clusterRoleBinding.RoleRef
		found.Labels = clusterRoleBinding.Labels
		found.Annotations = clusterRoleBinding.Annotations
		if err = r.Update(ctx, found); err != nil {
			return errors.Wrap(err, "Could not Update ClusterRoleBinding")
		}
	} else {
		logr.Debugw("No updates for ClusterRoleBinding")
	}
	return nil
}
