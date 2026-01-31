package rbac

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/morimasaki-web/ponsu/internal/infrastructure/dbgen"
)

// Permission はパーミッションIDを表す型（タイプセーフティのため）
type Permission string

const (
	PermissionRequestView    Permission = "request.view"
	PermissionRequestCreate  Permission = "request.create"
	PermissionRequestApprove Permission = "request.approve"
	PermissionRequestExport  Permission = "request.export"
)

// HasPermission はロールが指定されたパーミッションを持っているかチェックする
func HasPermission(ctx context.Context, db *sql.DB, role string, perm Permission) (bool, error) {
	if db == nil {
		return false, errors.New("database not available")
	}

	q := dbgen.New(db)
	result, err := q.HasPermission(ctx, dbgen.HasPermissionParams{
		Role:         role,
		PermissionID: string(perm),
	})
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	return result, nil
}

// RequirePermission はパーミッションを持っていない場合にエラーを返す
func RequirePermission(ctx context.Context, db *sql.DB, role string, perm Permission) error {
	hasPermission, err := HasPermission(ctx, db, role, perm)
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("permission denied: %s required", perm)
	}
	return nil
}
