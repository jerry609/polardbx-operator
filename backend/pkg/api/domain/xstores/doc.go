package xstores

// Package xstores 提供“存储/引擎域”的统一入口。
//
// 领域边界（BFF）：
// - XStore CRUD、Pod 管理、故障恢复能力（Follower/重搭/日志/同步）
// - XStoreBackup（单机备份）
// - XStoreFollower（备库重搭）
//
// 第一阶段：路由薄壳转发到既有 handlers，不迁移实现，确保接口不变。
// 对应 CRD（api/v1）：XStore / XStoreBackup / XStoreFollower。
