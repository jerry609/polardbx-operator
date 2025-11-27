package logs

import (
	"polardbx-ui-backend/pkg/api/domain/platform/logs/handler"
)

// Re-export LogPreset type for backward compatibility
type LogPreset = handler.LogPreset

// Presets delegates to domain handler
var Presets = handler.Presets

// PresetByPattern delegates to domain handler
var PresetByPattern = handler.PresetByPattern
