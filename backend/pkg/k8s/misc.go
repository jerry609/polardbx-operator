package k8s

import (
	"context"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListPodsForPolarDBXCluster lists all pods for a PolarDBXCluster.
// Deprecated: Use ListPodsForPolarDBXClusterWithContext for better context control.
func ListPodsForPolarDBXCluster(c client.Client, namespace, clusterName string) ([]corev1.Pod, error) {
	return ListPodsForPolarDBXClusterWithContext(context.TODO(), c, namespace, clusterName)
}

func ListPodsForPolarDBXClusterWithContext(ctx context.Context, c client.Client, namespace, clusterName string) ([]corev1.Pod, error) {
	var podList corev1.PodList
	opts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels{
			"polardbx/name": clusterName,
		},
	}
	if err := c.List(ctx, &podList, opts...); err != nil {
		return nil, err
	}
	return podList.Items, nil
}

// ---- PolarDBXParameter CRUD ----

// Deprecated: Use ListPolarDBXParametersWithContext for better context control.
func ListPolarDBXParameters(c client.Client, namespace string) ([]polardbxv1.PolarDBXParameter, error) {
	return ListPolarDBXParametersWithContext(context.TODO(), c, namespace)
}

// Deprecated: Use GetPolarDBXParameterWithContext for better context control.
func GetPolarDBXParameter(c client.Client, namespace, name string) (*polardbxv1.PolarDBXParameter, error) {
	return GetPolarDBXParameterWithContext(context.TODO(), c, namespace, name)
}

// Deprecated: Use CreatePolarDBXParameterWithContext for better context control.
func CreatePolarDBXParameter(c client.Client, namespace string, param *polardbxv1.PolarDBXParameter) (*polardbxv1.PolarDBXParameter, error) {
	return CreatePolarDBXParameterWithContext(context.TODO(), c, namespace, param)
}

// Deprecated: Use UpdatePolarDBXParameterWithContext for better context control.
func UpdatePolarDBXParameter(c client.Client, namespace string, param *polardbxv1.PolarDBXParameter) (*polardbxv1.PolarDBXParameter, error) {
	return UpdatePolarDBXParameterWithContext(context.TODO(), c, namespace, param)
}

// Deprecated: Use DeletePolarDBXParameterWithContext for better context control.
func DeletePolarDBXParameter(c client.Client, namespace, name string) error {
	return DeletePolarDBXParameterWithContext(context.TODO(), c, namespace, name)
}

func ListPolarDBXParametersWithContext(ctx context.Context, c client.Client, namespace string) ([]polardbxv1.PolarDBXParameter, error) {
	var list polardbxv1.PolarDBXParameterList
	if err := c.List(ctx, &list, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	return list.Items, nil
}

func GetPolarDBXParameterWithContext(ctx context.Context, c client.Client, namespace, name string) (*polardbxv1.PolarDBXParameter, error) {
	var obj polardbxv1.PolarDBXParameter
	if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func CreatePolarDBXParameterWithContext(ctx context.Context, c client.Client, namespace string, param *polardbxv1.PolarDBXParameter) (*polardbxv1.PolarDBXParameter, error) {
	if param.Namespace == "" {
		param.Namespace = namespace
	}
	if err := c.Create(ctx, param); err != nil {
		return nil, err
	}
	return param, nil
}

func UpdatePolarDBXParameterWithContext(ctx context.Context, c client.Client, namespace string, param *polardbxv1.PolarDBXParameter) (*polardbxv1.PolarDBXParameter, error) {
	if err := c.Update(ctx, param); err != nil {
		return nil, err
	}
	return param, nil
}

func DeletePolarDBXParameterWithContext(ctx context.Context, c client.Client, namespace, name string) error {
	obj := &polardbxv1.PolarDBXParameter{}
	obj.Name = name
	obj.Namespace = namespace
	return c.Delete(ctx, obj)
}

// ---- PolarDBXParameterTemplate CRUD ----

// Deprecated: Use ListPolarDBXParameterTemplatesWithContext for better context control.
func ListPolarDBXParameterTemplates(c client.Client, namespace string) ([]polardbxv1.PolarDBXParameterTemplate, error) {
	return ListPolarDBXParameterTemplatesWithContext(context.TODO(), c, namespace)
}

// Deprecated: Use CreatePolarDBXParameterTemplateWithContext for better context control.
func CreatePolarDBXParameterTemplate(c client.Client, namespace string, template *polardbxv1.PolarDBXParameterTemplate) (*polardbxv1.PolarDBXParameterTemplate, error) {
	return CreatePolarDBXParameterTemplateWithContext(context.TODO(), c, namespace, template)
}

// Deprecated: Use GetPolarDBXParameterTemplateWithContext for better context control.
func GetPolarDBXParameterTemplate(c client.Client, namespace, name string) (*polardbxv1.PolarDBXParameterTemplate, error) {
	return GetPolarDBXParameterTemplateWithContext(context.TODO(), c, namespace, name)
}

// Deprecated: Use UpdatePolarDBXParameterTemplateWithContext for better context control.
func UpdatePolarDBXParameterTemplate(c client.Client, namespace string, template *polardbxv1.PolarDBXParameterTemplate) (*polardbxv1.PolarDBXParameterTemplate, error) {
	return UpdatePolarDBXParameterTemplateWithContext(context.TODO(), c, namespace, template)
}

// Deprecated: Use DeletePolarDBXParameterTemplateWithContext for better context control.
func DeletePolarDBXParameterTemplate(c client.Client, namespace, name string) error {
	return DeletePolarDBXParameterTemplateWithContext(context.TODO(), c, namespace, name)
}

func ListPolarDBXParameterTemplatesWithContext(ctx context.Context, c client.Client, namespace string) ([]polardbxv1.PolarDBXParameterTemplate, error) {
	var list polardbxv1.PolarDBXParameterTemplateList
	if err := c.List(ctx, &list, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	return list.Items, nil
}

func CreatePolarDBXParameterTemplateWithContext(ctx context.Context, c client.Client, namespace string, tpl *polardbxv1.PolarDBXParameterTemplate) (*polardbxv1.PolarDBXParameterTemplate, error) {
	if tpl.Namespace == "" {
		tpl.Namespace = namespace
	}
	if err := c.Create(ctx, tpl); err != nil {
		return nil, err
	}
	return tpl, nil
}

func GetPolarDBXParameterTemplateWithContext(ctx context.Context, c client.Client, namespace, name string) (*polardbxv1.PolarDBXParameterTemplate, error) {
	var obj polardbxv1.PolarDBXParameterTemplate
	if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func UpdatePolarDBXParameterTemplateWithContext(ctx context.Context, c client.Client, namespace string, tpl *polardbxv1.PolarDBXParameterTemplate) (*polardbxv1.PolarDBXParameterTemplate, error) {
	if err := c.Update(ctx, tpl); err != nil {
		return nil, err
	}
	return tpl, nil
}

func DeletePolarDBXParameterTemplateWithContext(ctx context.Context, c client.Client, namespace, name string) error {
	obj := &polardbxv1.PolarDBXParameterTemplate{}
	obj.Name = name
	obj.Namespace = namespace
	return c.Delete(ctx, obj)
}

// ---- SystemTask CRUD ----

// Deprecated: Use ListSystemTasksWithContext for better context control.
func ListSystemTasks(c client.Client, namespace string) ([]polardbxv1.SystemTask, error) {
	return ListSystemTasksWithContext(context.TODO(), c, namespace)
}

// Deprecated: Use CreateSystemTaskWithContext for better context control.
func CreateSystemTask(c client.Client, namespace string, task *polardbxv1.SystemTask) (*polardbxv1.SystemTask, error) {
	return CreateSystemTaskWithContext(context.TODO(), c, namespace, task)
}

// Deprecated: Use GetSystemTaskWithContext for better context control.
func GetSystemTask(c client.Client, namespace, name string) (*polardbxv1.SystemTask, error) {
	return GetSystemTaskWithContext(context.TODO(), c, namespace, name)
}

// Deprecated: Use UpdateSystemTaskWithContext for better context control.
func UpdateSystemTask(c client.Client, namespace string, task *polardbxv1.SystemTask) (*polardbxv1.SystemTask, error) {
	return UpdateSystemTaskWithContext(context.TODO(), c, namespace, task)
}

// Deprecated: Use DeleteSystemTaskWithContext for better context control.
func DeleteSystemTask(c client.Client, namespace, name string) error {
	return DeleteSystemTaskWithContext(context.TODO(), c, namespace, name)
}

func ListSystemTasksWithContext(ctx context.Context, c client.Client, namespace string) ([]polardbxv1.SystemTask, error) {
	var list polardbxv1.SystemTaskList
	if err := c.List(ctx, &list, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	return list.Items, nil
}

func CreateSystemTaskWithContext(ctx context.Context, c client.Client, namespace string, task *polardbxv1.SystemTask) (*polardbxv1.SystemTask, error) {
	if task.Namespace == "" {
		task.Namespace = namespace
	}
	if err := c.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func GetSystemTaskWithContext(ctx context.Context, c client.Client, namespace, name string) (*polardbxv1.SystemTask, error) {
	var obj polardbxv1.SystemTask
	if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func UpdateSystemTaskWithContext(ctx context.Context, c client.Client, namespace string, task *polardbxv1.SystemTask) (*polardbxv1.SystemTask, error) {
	if err := c.Update(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func DeleteSystemTaskWithContext(ctx context.Context, c client.Client, namespace, name string) error {
	obj := &polardbxv1.SystemTask{}
	obj.Name = name
	obj.Namespace = namespace
	return c.Delete(ctx, obj)
}

// ---- PolarDBXParameterTemplate Management Functions ----
// (Already covered above with WithContext variants and deprecated wrappers)

// ---- PolarDBXMonitor CRUD ----

// Deprecated: Use ListPolarDBXMonitorsWithContext for better context control.
func ListPolarDBXMonitors(c client.Client, namespace string) ([]polardbxv1.PolarDBXMonitor, error) {
	return ListPolarDBXMonitorsWithContext(context.TODO(), c, namespace)
}

// Deprecated: Use CreatePolarDBXMonitorWithContext for better context control.
func CreatePolarDBXMonitor(c client.Client, namespace string, monitor *polardbxv1.PolarDBXMonitor) (*polardbxv1.PolarDBXMonitor, error) {
	return CreatePolarDBXMonitorWithContext(context.TODO(), c, namespace, monitor)
}

// Deprecated: Use GetPolarDBXMonitorWithContext for better context control.
func GetPolarDBXMonitor(c client.Client, namespace, name string) (*polardbxv1.PolarDBXMonitor, error) {
	return GetPolarDBXMonitorWithContext(context.TODO(), c, namespace, name)
}

// Deprecated: Use UpdatePolarDBXMonitorWithContext for better context control.
func UpdatePolarDBXMonitor(c client.Client, namespace string, monitor *polardbxv1.PolarDBXMonitor) (*polardbxv1.PolarDBXMonitor, error) {
	return UpdatePolarDBXMonitorWithContext(context.TODO(), c, namespace, monitor)
}

// Deprecated: Use DeletePolarDBXMonitorWithContext for better context control.
func DeletePolarDBXMonitor(c client.Client, namespace, name string) error {
	return DeletePolarDBXMonitorWithContext(context.TODO(), c, namespace, name)
}

func ListPolarDBXMonitorsWithContext(ctx context.Context, c client.Client, namespace string) ([]polardbxv1.PolarDBXMonitor, error) {
	var list polardbxv1.PolarDBXMonitorList
	if err := c.List(ctx, &list, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	return list.Items, nil
}

func CreatePolarDBXMonitorWithContext(ctx context.Context, c client.Client, namespace string, m *polardbxv1.PolarDBXMonitor) (*polardbxv1.PolarDBXMonitor, error) {
	if m.Namespace == "" {
		m.Namespace = namespace
	}
	if err := c.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func GetPolarDBXMonitorWithContext(ctx context.Context, c client.Client, namespace, name string) (*polardbxv1.PolarDBXMonitor, error) {
	var out polardbxv1.PolarDBXMonitor
	if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func UpdatePolarDBXMonitorWithContext(ctx context.Context, c client.Client, namespace string, m *polardbxv1.PolarDBXMonitor) (*polardbxv1.PolarDBXMonitor, error) {
	if err := c.Update(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func DeletePolarDBXMonitorWithContext(ctx context.Context, c client.Client, namespace, name string) error {
	obj := &polardbxv1.PolarDBXMonitor{}
	obj.Name = name
	obj.Namespace = namespace
	return c.Delete(ctx, obj)
}

// ---- ClusterKnobs ----

// Deprecated: Use GetClusterKnobsListWithContext for better context control.
func GetClusterKnobsList(c client.Client) (*polardbxv1.PolarDBXClusterKnobsList, error) {
	return GetClusterKnobsListWithContext(context.TODO(), c)
}

func GetClusterKnobsListWithContext(ctx context.Context, c client.Client) (*polardbxv1.PolarDBXClusterKnobsList, error) {
	knobsList := &polardbxv1.PolarDBXClusterKnobsList{}
	err := c.List(ctx, knobsList)
	return knobsList, err
}

// Deprecated: Use CreateClusterKnobsWithContext for better context control.
func CreateClusterKnobs(c client.Client, knobs *polardbxv1.PolarDBXClusterKnobs) (*polardbxv1.PolarDBXClusterKnobs, error) {
	return CreateClusterKnobsWithContext(context.TODO(), c, knobs)
}

func CreateClusterKnobsWithContext(ctx context.Context, c client.Client, knobs *polardbxv1.PolarDBXClusterKnobs) (*polardbxv1.PolarDBXClusterKnobs, error) {
	err := c.Create(ctx, knobs)
	return knobs, err
}

// Deprecated: Use GetClusterKnobsWithContext for better context control.
func GetClusterKnobs(c client.Client, namespace, name string) (*polardbxv1.PolarDBXClusterKnobs, error) {
	return GetClusterKnobsWithContext(context.TODO(), c, namespace, name)
}

func GetClusterKnobsWithContext(ctx context.Context, c client.Client, namespace, name string) (*polardbxv1.PolarDBXClusterKnobs, error) {
	knobs := &polardbxv1.PolarDBXClusterKnobs{}
	err := c.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, knobs)
	return knobs, err
}

// Deprecated: Use UpdateClusterKnobsWithContext for better context control.
func UpdateClusterKnobs(c client.Client, knobs *polardbxv1.PolarDBXClusterKnobs) (*polardbxv1.PolarDBXClusterKnobs, error) {
	return UpdateClusterKnobsWithContext(context.TODO(), c, knobs)
}

func UpdateClusterKnobsWithContext(ctx context.Context, c client.Client, knobs *polardbxv1.PolarDBXClusterKnobs) (*polardbxv1.PolarDBXClusterKnobs, error) {
	err := c.Update(ctx, knobs)
	return knobs, err
}

// Deprecated: Use DeleteClusterKnobsWithContext for better context control.
func DeleteClusterKnobs(c client.Client, namespace, name string) error {
	return DeleteClusterKnobsWithContext(context.TODO(), c, namespace, name)
}

func DeleteClusterKnobsWithContext(ctx context.Context, c client.Client, namespace, name string) error {
	knobs := &polardbxv1.PolarDBXClusterKnobs{}
	knobs.Name = name
	knobs.Namespace = namespace
	return c.Delete(ctx, knobs)
}

// ---- PolarDBXLogCollector CRUD ----

// Deprecated: Use ListPolarDBXLogCollectorsWithContext for better context control.
func ListPolarDBXLogCollectors(c client.Client, namespace string) ([]polardbxv1.PolarDBXLogCollector, error) {
	return ListPolarDBXLogCollectorsWithContext(context.TODO(), c, namespace)
}

// Deprecated: Use CreatePolarDBXLogCollectorWithContext for better context control.
func CreatePolarDBXLogCollector(c client.Client, namespace string, obj *polardbxv1.PolarDBXLogCollector) (*polardbxv1.PolarDBXLogCollector, error) {
	return CreatePolarDBXLogCollectorWithContext(context.TODO(), c, namespace, obj)
}

// Deprecated: Use GetPolarDBXLogCollectorWithContext for better context control.
func GetPolarDBXLogCollector(c client.Client, namespace, name string) (*polardbxv1.PolarDBXLogCollector, error) {
	return GetPolarDBXLogCollectorWithContext(context.TODO(), c, namespace, name)
}

// Deprecated: Use UpdatePolarDBXLogCollectorWithContext for better context control.
func UpdatePolarDBXLogCollector(c client.Client, namespace string, obj *polardbxv1.PolarDBXLogCollector) (*polardbxv1.PolarDBXLogCollector, error) {
	return UpdatePolarDBXLogCollectorWithContext(context.TODO(), c, namespace, obj)
}

// Deprecated: Use DeletePolarDBXLogCollectorWithContext for better context control.
func DeletePolarDBXLogCollector(c client.Client, namespace, name string) error {
	return DeletePolarDBXLogCollectorWithContext(context.TODO(), c, namespace, name)
}

func ListPolarDBXLogCollectorsWithContext(ctx context.Context, c client.Client, namespace string) ([]polardbxv1.PolarDBXLogCollector, error) {
	var list polardbxv1.PolarDBXLogCollectorList
	if err := c.List(ctx, &list, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	return list.Items, nil
}

func CreatePolarDBXLogCollectorWithContext(ctx context.Context, c client.Client, namespace string, obj *polardbxv1.PolarDBXLogCollector) (*polardbxv1.PolarDBXLogCollector, error) {
	if obj.Namespace == "" {
		obj.Namespace = namespace
	}
	if err := c.Create(ctx, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func GetPolarDBXLogCollectorWithContext(ctx context.Context, c client.Client, namespace, name string) (*polardbxv1.PolarDBXLogCollector, error) {
	var out polardbxv1.PolarDBXLogCollector
	if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func UpdatePolarDBXLogCollectorWithContext(ctx context.Context, c client.Client, namespace string, obj *polardbxv1.PolarDBXLogCollector) (*polardbxv1.PolarDBXLogCollector, error) {
	if err := c.Update(ctx, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func DeletePolarDBXLogCollectorWithContext(ctx context.Context, c client.Client, namespace, name string) error {
	obj := &polardbxv1.PolarDBXLogCollector{}
	obj.Name = name
	obj.Namespace = namespace
	return c.Delete(ctx, obj)
}
