package systemtasks

import (
	"polardbx-ui-backend/pkg/api/domain/systemtasks/handler"
	"polardbx-ui-backend/pkg/api/domain/systemtasks/repository"
	"polardbx-ui-backend/pkg/api/domain/systemtasks/service"

	"github.com/gin-gonic/gin"
)

// 默认的 Handler 实例（使用依赖注入）
var defaultHandler = handler.NewSystemTaskHandler(
	service.NewSystemTaskService(
		repository.NewK8sSystemTaskRepository(),
	),
)

// 以下函数保持向后兼容，委托给 Handler

func List(c *gin.Context)   { defaultHandler.List(c) }
func Create(c *gin.Context) { defaultHandler.Create(c) }
func Get(c *gin.Context)    { defaultHandler.Get(c) }
func Update(c *gin.Context) { defaultHandler.Update(c) }
func Delete(c *gin.Context) { defaultHandler.Delete(c) }
