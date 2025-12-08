package k8s

import (
	"context"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ---- PolarDBXBackup (and dry-run) ----

// Deprecated: Use ListPolarDBXBackupsWithContext for better context control.
func ListPolarDBXBackups(c client.Client, namespace, clusterName string) ([]polardbxv1.PolarDBXBackup, error) {
	return ListPolarDBXBackupsWithContext(context.TODO(), c, namespace, clusterName)
}

// Deprecated: Use CreatePolarDBXBackupWithContext for better context control.
func CreatePolarDBXBackup(c client.Client, namespace string, backup *polardbxv1.PolarDBXBackup) (*polardbxv1.PolarDBXBackup, error) {
	return CreatePolarDBXBackupWithContext(context.TODO(), c, namespace, backup)
}

// Deprecated: Use CreatePolarDBXBackupDryRunWithContext for better context control.
func CreatePolarDBXBackupDryRun(c client.Client, namespace string, backup *polardbxv1.PolarDBXBackup) (*polardbxv1.PolarDBXBackup, error) {
	return CreatePolarDBXBackupDryRunWithContext(context.TODO(), c, namespace, backup)
}

// Deprecated: Use DeletePolarDBXBackupWithContext for better context control.
func DeletePolarDBXBackup(c client.Client, namespace, name string) error {
	return DeletePolarDBXBackupWithContext(context.TODO(), c, namespace, name)
}

func ListPolarDBXBackupsWithContext(ctx context.Context, c client.Client, namespace, clusterName string) ([]polardbxv1.PolarDBXBackup, error) {
	var list polardbxv1.PolarDBXBackupList
	opts := []client.ListOption{client.InNamespace(namespace), client.MatchingLabels{"polardbx/name": clusterName}}
	if err := c.List(ctx, &list, opts...); err != nil {
		return nil, err
	}
	return list.Items, nil
}

func CreatePolarDBXBackupWithContext(ctx context.Context, c client.Client, namespace string, backup *polardbxv1.PolarDBXBackup) (*polardbxv1.PolarDBXBackup, error) {
	if backup.Namespace == "" {
		backup.Namespace = namespace
	}
	if err := c.Create(ctx, backup); err != nil {
		return nil, err
	}
	return backup, nil
}

func CreatePolarDBXBackupDryRunWithContext(ctx context.Context, c client.Client, namespace string, backup *polardbxv1.PolarDBXBackup) (*polardbxv1.PolarDBXBackup, error) {
	if backup.Namespace == "" {
		backup.Namespace = namespace
	}
	opts := &client.CreateOptions{DryRun: []string{metav1.DryRunAll}}
	if err := c.Create(ctx, backup, opts); err != nil {
		return nil, err
	}
	return backup, nil
}

func DeletePolarDBXBackupWithContext(ctx context.Context, c client.Client, namespace, name string) error {
	obj := &polardbxv1.PolarDBXBackup{}
	obj.Name = name
	obj.Namespace = namespace
	return c.Delete(ctx, obj)
}

// ---- PolarDBXBackupSchedule ----

// Deprecated: Use ListPolarDBXBackupSchedulesWithContext for better context control.
func ListPolarDBXBackupSchedules(c client.Client, namespace string) ([]polardbxv1.PolarDBXBackupSchedule, error) {
	return ListPolarDBXBackupSchedulesWithContext(context.TODO(), c, namespace)
}

// Deprecated: Use CreatePolarDBXBackupScheduleWithContext for better context control.
func CreatePolarDBXBackupSchedule(c client.Client, namespace string, schedule *polardbxv1.PolarDBXBackupSchedule) (*polardbxv1.PolarDBXBackupSchedule, error) {
	return CreatePolarDBXBackupScheduleWithContext(context.TODO(), c, namespace, schedule)
}

// Deprecated: Use GetPolarDBXBackupScheduleWithContext for better context control.
func GetPolarDBXBackupSchedule(c client.Client, namespace, name string) (*polardbxv1.PolarDBXBackupSchedule, error) {
	return GetPolarDBXBackupScheduleWithContext(context.TODO(), c, namespace, name)
}

// Deprecated: Use UpdatePolarDBXBackupScheduleWithContext for better context control.
func UpdatePolarDBXBackupSchedule(c client.Client, namespace string, schedule *polardbxv1.PolarDBXBackupSchedule) (*polardbxv1.PolarDBXBackupSchedule, error) {
	return UpdatePolarDBXBackupScheduleWithContext(context.TODO(), c, namespace, schedule)
}

// Deprecated: Use DeletePolarDBXBackupScheduleWithContext for better context control.
func DeletePolarDBXBackupSchedule(c client.Client, namespace, name string) error {
	return DeletePolarDBXBackupScheduleWithContext(context.TODO(), c, namespace, name)
}

func ListPolarDBXBackupSchedulesWithContext(ctx context.Context, c client.Client, namespace string) ([]polardbxv1.PolarDBXBackupSchedule, error) {
	var list polardbxv1.PolarDBXBackupScheduleList
	if err := c.List(ctx, &list, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	return list.Items, nil
}

func CreatePolarDBXBackupScheduleWithContext(ctx context.Context, c client.Client, namespace string, schedule *polardbxv1.PolarDBXBackupSchedule) (*polardbxv1.PolarDBXBackupSchedule, error) {
	if schedule.Namespace == "" {
		schedule.Namespace = namespace
	}
	if err := c.Create(ctx, schedule); err != nil {
		return nil, err
	}
	return schedule, nil
}

func GetPolarDBXBackupScheduleWithContext(ctx context.Context, c client.Client, namespace, name string) (*polardbxv1.PolarDBXBackupSchedule, error) {
	var item polardbxv1.PolarDBXBackupSchedule
	if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func UpdatePolarDBXBackupScheduleWithContext(ctx context.Context, c client.Client, namespace string, schedule *polardbxv1.PolarDBXBackupSchedule) (*polardbxv1.PolarDBXBackupSchedule, error) {
	if err := c.Update(ctx, schedule); err != nil {
		return nil, err
	}
	return schedule, nil
}

func DeletePolarDBXBackupScheduleWithContext(ctx context.Context, c client.Client, namespace, name string) error {
	obj := &polardbxv1.PolarDBXBackupSchedule{}
	obj.Name = name
	obj.Namespace = namespace
	return c.Delete(ctx, obj)
}

// ---- PolarDBXBackupBinlog ----

// Deprecated: Use ListPolarDBXBackupBinlogsWithContext for better context control.
func ListPolarDBXBackupBinlogs(c client.Client, namespace string) ([]polardbxv1.PolarDBXBackupBinlog, error) {
	return ListPolarDBXBackupBinlogsWithContext(context.TODO(), c, namespace)
}

// Deprecated: Use CreatePolarDBXBackupBinlogWithContext for better context control.
func CreatePolarDBXBackupBinlog(c client.Client, namespace string, binlog *polardbxv1.PolarDBXBackupBinlog) (*polardbxv1.PolarDBXBackupBinlog, error) {
	return CreatePolarDBXBackupBinlogWithContext(context.TODO(), c, namespace, binlog)
}

// Deprecated: Use GetPolarDBXBackupBinlogWithContext for better context control.
func GetPolarDBXBackupBinlog(c client.Client, namespace, name string) (*polardbxv1.PolarDBXBackupBinlog, error) {
	return GetPolarDBXBackupBinlogWithContext(context.TODO(), c, namespace, name)
}

// Deprecated: Use UpdatePolarDBXBackupBinlogWithContext for better context control.
func UpdatePolarDBXBackupBinlog(c client.Client, namespace string, binlog *polardbxv1.PolarDBXBackupBinlog) (*polardbxv1.PolarDBXBackupBinlog, error) {
	return UpdatePolarDBXBackupBinlogWithContext(context.TODO(), c, namespace, binlog)
}

// Deprecated: Use DeletePolarDBXBackupBinlogWithContext for better context control.
func DeletePolarDBXBackupBinlog(c client.Client, namespace, name string) error {
	return DeletePolarDBXBackupBinlogWithContext(context.TODO(), c, namespace, name)
}

func ListPolarDBXBackupBinlogsWithContext(ctx context.Context, c client.Client, namespace string) ([]polardbxv1.PolarDBXBackupBinlog, error) {
	var list polardbxv1.PolarDBXBackupBinlogList
	if err := c.List(ctx, &list, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	return list.Items, nil
}

func CreatePolarDBXBackupBinlogWithContext(ctx context.Context, c client.Client, namespace string, obj *polardbxv1.PolarDBXBackupBinlog) (*polardbxv1.PolarDBXBackupBinlog, error) {
	if obj.Namespace == "" {
		obj.Namespace = namespace
	}
	if err := c.Create(ctx, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func GetPolarDBXBackupBinlogWithContext(ctx context.Context, c client.Client, namespace, name string) (*polardbxv1.PolarDBXBackupBinlog, error) {
	var out polardbxv1.PolarDBXBackupBinlog
	if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func UpdatePolarDBXBackupBinlogWithContext(ctx context.Context, c client.Client, namespace string, obj *polardbxv1.PolarDBXBackupBinlog) (*polardbxv1.PolarDBXBackupBinlog, error) {
	if err := c.Update(ctx, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func DeletePolarDBXBackupBinlogWithContext(ctx context.Context, c client.Client, namespace, name string) error {
	binlog := &polardbxv1.PolarDBXBackupBinlog{}
	binlog.Name = name
	binlog.Namespace = namespace
	return c.Delete(ctx, binlog)
}
