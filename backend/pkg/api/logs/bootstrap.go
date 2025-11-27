package logs

import (
	"polardbx-ui-backend/pkg/api/domain/platform/logs/handler"
)

// Bootstrap delegates to domain handler
var Bootstrap = handler.Bootstrap

// BootstrapStatus delegates to domain handler
var BootstrapStatus = handler.BootstrapStatus

// BootstrapLogs delegates to domain handler
var BootstrapLogs = handler.BootstrapLogs
