package systemtasks

// Package systemtasks 提供“平台任务域”的统一入口。
//
// 领域边界（BFF）：
// - SystemTask CRUD 与状态管理
//
// 第一阶段：仅路由薄壳，转发到既有 handlers；接口不变。
// 对应 CRD（api/v1）：SystemTask。
