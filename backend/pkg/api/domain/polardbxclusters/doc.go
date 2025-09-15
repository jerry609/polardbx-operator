package polardbxclusters

// Package polardbxclusters 提供“逻辑集群域”的统一入口。
//
// 领域边界（BFF）：
// - Cluster 编排与运维（scale/upgrade/log-config/alerts-summary/pods）
// - Backups / BackupSchedules / BackupBinlogs（集群备份与增量）
// - Parameters / ParameterTemplates（参数与模板）
// - Prechange / Restore / PITR（变更前检查与恢复）
// - ClusterKnobs（性能调优开关）
//
// 第一阶段（低风险）：仅提供路由薄壳，内部转发到既有 handlers；
// 不迁移原有实现，确保接口路径与入参/出参保持不变。
//
// 对应 CRD（api/v1）：
// - PolarDBXCluster
// - PolarDBXBackup / PolarDBXBackupSchedule / PolarDBXBackupBinlog
// - PolarDBXParameter / PolarDBXParameterTemplate
//
// 典型别名路由（只新增 alias，不改变既有路由）：
// - /api/v1/crd/polardbxclusters
// - /api/v1/crd/polardbxbackups
// - /api/v1/crd/polardbxbackupschedules
// - /api/v1/crd/polardbxbackupbinlogs
// - /api/v1/crd/polardbxparameters
// - /api/v1/crd/polardbxparametertemplates
//
// 后续阶段：逐步将编排下沉到 services，K8s 访问统一走 k8srepo，
// 并在 meta 中收敛常量/标签/字段名定义。
