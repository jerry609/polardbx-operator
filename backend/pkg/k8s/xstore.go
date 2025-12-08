package k8s

import (
	"context"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ---- XStore ----

// Deprecated: Use ListXStoresWithContext for better context control.
func ListXStores(c client.Client, namespace string) ([]polardbxv1.XStore, error) {
	return ListXStoresWithContext(context.TODO(), c, namespace)
}

// Deprecated: Use CreateXStoreWithContext for better context control.
func CreateXStore(c client.Client, namespace string, xstore *polardbxv1.XStore) (*polardbxv1.XStore, error) {
	return CreateXStoreWithContext(context.TODO(), c, namespace, xstore)
}

// Deprecated: Use GetXStoreWithContext for better context control.
func GetXStore(c client.Client, namespace, name string) (*polardbxv1.XStore, error) {
	return GetXStoreWithContext(context.TODO(), c, namespace, name)
}

// Deprecated: Use UpdateXStoreWithContext for better context control.
func UpdateXStore(c client.Client, namespace string, xstore *polardbxv1.XStore) (*polardbxv1.XStore, error) {
	return UpdateXStoreWithContext(context.TODO(), c, namespace, xstore)
}

// Deprecated: Use DeleteXStoreWithContext for better context control.
func DeleteXStore(c client.Client, namespace, name string) error {
	return DeleteXStoreWithContext(context.TODO(), c, namespace, name)
}

func ListXStoresWithContext(ctx context.Context, c client.Client, namespace string) ([]polardbxv1.XStore, error) {
	var list polardbxv1.XStoreList
	if err := c.List(ctx, &list, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	return list.Items, nil
}

func CreateXStoreWithContext(ctx context.Context, c client.Client, namespace string, x *polardbxv1.XStore) (*polardbxv1.XStore, error) {
	if x.Namespace == "" {
		x.Namespace = namespace
	}
	if err := c.Create(ctx, x); err != nil {
		return nil, err
	}
	return x, nil
}

func GetXStoreWithContext(ctx context.Context, c client.Client, namespace, name string) (*polardbxv1.XStore, error) {
	var out polardbxv1.XStore
	if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func UpdateXStoreWithContext(ctx context.Context, c client.Client, namespace string, x *polardbxv1.XStore) (*polardbxv1.XStore, error) {
	if err := c.Update(ctx, x); err != nil {
		return nil, err
	}
	return x, nil
}

func DeleteXStoreWithContext(ctx context.Context, c client.Client, namespace, name string) error {
	obj := &polardbxv1.XStore{}
	obj.Name = name
	obj.Namespace = namespace
	return c.Delete(ctx, obj)
}

// ---- XStoreBackup ----

// Deprecated: Use ListXStoreBackupsWithContext for better context control.
func ListXStoreBackups(c client.Client, namespace string) ([]polardbxv1.XStoreBackup, error) {
	return ListXStoreBackupsWithContext(context.TODO(), c, namespace)
}

// Deprecated: Use CreateXStoreBackupWithContext for better context control.
func CreateXStoreBackup(c client.Client, namespace string, backup *polardbxv1.XStoreBackup) (*polardbxv1.XStoreBackup, error) {
	return CreateXStoreBackupWithContext(context.TODO(), c, namespace, backup)
}

// Deprecated: Use GetXStoreBackupWithContext for better context control.
func GetXStoreBackup(c client.Client, namespace, name string) (*polardbxv1.XStoreBackup, error) {
	return GetXStoreBackupWithContext(context.TODO(), c, namespace, name)
}

// Deprecated: Use UpdateXStoreBackupWithContext for better context control.
func UpdateXStoreBackup(c client.Client, namespace string, backup *polardbxv1.XStoreBackup) (*polardbxv1.XStoreBackup, error) {
	return UpdateXStoreBackupWithContext(context.TODO(), c, namespace, backup)
}

// Deprecated: Use DeleteXStoreBackupWithContext for better context control.
func DeleteXStoreBackup(c client.Client, namespace, name string) error {
	return DeleteXStoreBackupWithContext(context.TODO(), c, namespace, name)
}

func ListXStoreBackupsWithContext(ctx context.Context, c client.Client, namespace string) ([]polardbxv1.XStoreBackup, error) {
	var list polardbxv1.XStoreBackupList
	if err := c.List(ctx, &list, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	return list.Items, nil
}

func CreateXStoreBackupWithContext(ctx context.Context, c client.Client, namespace string, obj *polardbxv1.XStoreBackup) (*polardbxv1.XStoreBackup, error) {
	if obj.Namespace == "" {
		obj.Namespace = namespace
	}
	if err := c.Create(ctx, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func GetXStoreBackupWithContext(ctx context.Context, c client.Client, namespace, name string) (*polardbxv1.XStoreBackup, error) {
	var out polardbxv1.XStoreBackup
	if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func UpdateXStoreBackupWithContext(ctx context.Context, c client.Client, namespace string, obj *polardbxv1.XStoreBackup) (*polardbxv1.XStoreBackup, error) {
	if err := c.Update(ctx, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func DeleteXStoreBackupWithContext(ctx context.Context, c client.Client, namespace, name string) error {
	obj := &polardbxv1.XStoreBackup{}
	obj.Name = name
	obj.Namespace = namespace
	return c.Delete(ctx, obj)
}

// ---- XStoreFollower ----

// Deprecated: Use ListXStoreFollowersWithContext for better context control.
func ListXStoreFollowers(c client.Client, namespace string) ([]polardbxv1.XStoreFollower, error) {
	return ListXStoreFollowersWithContext(context.TODO(), c, namespace)
}

func ListXStoreFollowersWithContext(ctx context.Context, c client.Client, namespace string) ([]polardbxv1.XStoreFollower, error) {
	var followerList polardbxv1.XStoreFollowerList
	if err := c.List(ctx, &followerList, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	return followerList.Items, nil
}

// Deprecated: Use CreateXStoreFollowerWithContext for better context control.
func CreateXStoreFollower(c client.Client, namespace string, follower *polardbxv1.XStoreFollower) (*polardbxv1.XStoreFollower, error) {
	return CreateXStoreFollowerWithContext(context.TODO(), c, namespace, follower)
}

func CreateXStoreFollowerWithContext(ctx context.Context, c client.Client, namespace string, follower *polardbxv1.XStoreFollower) (*polardbxv1.XStoreFollower, error) {
	if follower.Namespace == "" {
		follower.Namespace = namespace
	}
	err := c.Create(ctx, follower)
	return follower, err
}

// Deprecated: Use GetXStoreFollowerWithContext for better context control.
func GetXStoreFollower(c client.Client, namespace, name string) (*polardbxv1.XStoreFollower, error) {
	return GetXStoreFollowerWithContext(context.TODO(), c, namespace, name)
}

func GetXStoreFollowerWithContext(ctx context.Context, c client.Client, namespace, name string) (*polardbxv1.XStoreFollower, error) {
	var follower polardbxv1.XStoreFollower
	err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &follower)
	if err != nil {
		return nil, err
	}
	return &follower, nil
}

// Deprecated: Use UpdateXStoreFollowerWithContext for better context control.
func UpdateXStoreFollower(c client.Client, namespace string, follower *polardbxv1.XStoreFollower) (*polardbxv1.XStoreFollower, error) {
	return UpdateXStoreFollowerWithContext(context.TODO(), c, namespace, follower)
}

func UpdateXStoreFollowerWithContext(ctx context.Context, c client.Client, namespace string, follower *polardbxv1.XStoreFollower) (*polardbxv1.XStoreFollower, error) {
	err := c.Update(ctx, follower)
	return follower, err
}

// Deprecated: Use DeleteXStoreFollowerWithContext for better context control.
func DeleteXStoreFollower(c client.Client, namespace, name string) error {
	return DeleteXStoreFollowerWithContext(context.TODO(), c, namespace, name)
}

func DeleteXStoreFollowerWithContext(ctx context.Context, c client.Client, namespace, name string) error {
	follower := &polardbxv1.XStoreFollower{}
	follower.Name = name
	follower.Namespace = namespace
	return c.Delete(ctx, follower)
}

// ---- XStoreBackupBinlog ----

// Deprecated: Use ListXStoreBackupBinlogsWithContext for better context control.
func ListXStoreBackupBinlogs(c client.Client, namespace string) ([]polardbxv1.XStoreBackupBinlog, error) {
	return ListXStoreBackupBinlogsWithContext(context.TODO(), c, namespace)
}

func ListXStoreBackupBinlogsWithContext(ctx context.Context, c client.Client, namespace string) ([]polardbxv1.XStoreBackupBinlog, error) {
	var list polardbxv1.XStoreBackupBinlogList
	if err := c.List(ctx, &list, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	return list.Items, nil
}

// Deprecated: Use CreateXStoreBackupBinlogWithContext for better context control.
func CreateXStoreBackupBinlog(c client.Client, namespace string, obj *polardbxv1.XStoreBackupBinlog) (*polardbxv1.XStoreBackupBinlog, error) {
	return CreateXStoreBackupBinlogWithContext(context.TODO(), c, namespace, obj)
}

func CreateXStoreBackupBinlogWithContext(ctx context.Context, c client.Client, namespace string, obj *polardbxv1.XStoreBackupBinlog) (*polardbxv1.XStoreBackupBinlog, error) {
	if obj.Namespace == "" {
		obj.Namespace = namespace
	}
	if err := c.Create(ctx, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

// Deprecated: Use GetXStoreBackupBinlogWithContext for better context control.
func GetXStoreBackupBinlog(c client.Client, namespace, name string) (*polardbxv1.XStoreBackupBinlog, error) {
	return GetXStoreBackupBinlogWithContext(context.TODO(), c, namespace, name)
}

func GetXStoreBackupBinlogWithContext(ctx context.Context, c client.Client, namespace, name string) (*polardbxv1.XStoreBackupBinlog, error) {
	var out polardbxv1.XStoreBackupBinlog
	if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Deprecated: Use UpdateXStoreBackupBinlogWithContext for better context control.
func UpdateXStoreBackupBinlog(c client.Client, namespace string, obj *polardbxv1.XStoreBackupBinlog) (*polardbxv1.XStoreBackupBinlog, error) {
	return UpdateXStoreBackupBinlogWithContext(context.TODO(), c, namespace, obj)
}

func UpdateXStoreBackupBinlogWithContext(ctx context.Context, c client.Client, namespace string, obj *polardbxv1.XStoreBackupBinlog) (*polardbxv1.XStoreBackupBinlog, error) {
	if obj.Namespace == "" {
		obj.Namespace = namespace
	}
	if err := c.Update(ctx, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

// Deprecated: Use DeleteXStoreBackupBinlogWithContext for better context control.
func DeleteXStoreBackupBinlog(c client.Client, namespace, name string) error {
	return DeleteXStoreBackupBinlogWithContext(context.TODO(), c, namespace, name)
}

func DeleteXStoreBackupBinlogWithContext(ctx context.Context, c client.Client, namespace, name string) error {
	obj := &polardbxv1.XStoreBackupBinlog{}
	obj.Name = name
	obj.Namespace = namespace
	return c.Delete(ctx, obj)
}
