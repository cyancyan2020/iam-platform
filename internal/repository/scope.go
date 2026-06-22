package repository

import (
	"context"

	"gorm.io/gorm"
)

// DataScope 数据权限范围
type DataScope struct {
	All     bool
	DeptIDs []uint64
	Self    bool
	UserID  uint64
}

type scopeContextKey string

const dataScopeKey scopeContextKey = "dataScope"

// SetDataScope 将 DataScope 存入 context
func SetDataScope(ctx context.Context, scope DataScope) context.Context {
	return context.WithValue(ctx, dataScopeKey, scope)
}

// GetDataScope 从 context 获取 DataScope
func GetDataScope(ctx context.Context) (DataScope, bool) {
	scope, ok := ctx.Value(dataScopeKey).(DataScope)
	return scope, ok
}

// ApplyScope 将 DataScope 转换为 GORM Scope 函数
func ApplyScope(scope DataScope) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if scope.All {
			return db
		}
		if scope.Self {
			return db.Where("user_id = ?", scope.UserID)
		}
		return db.Where("1 = 0")
	}
}
