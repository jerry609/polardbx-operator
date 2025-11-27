package logstrategy

import (
	"polardbx-ui-backend/pkg/api/domain/platform/logstrategy/handler"
)

// Re-export types for backward compatibility
type Strategy = handler.Strategy
type ApplyRecord = handler.ApplyRecord

// Delegate all handlers to domain handler
var (
	List             = handler.List
	Get              = handler.Get
	Create           = handler.Create
	Update           = handler.Update
	Delete           = handler.Delete
	Precheck         = handler.Precheck
	Apply            = handler.Apply
	TestConnection   = handler.TestConnection
	ListApplyRecords = handler.ListApplyRecords
)
