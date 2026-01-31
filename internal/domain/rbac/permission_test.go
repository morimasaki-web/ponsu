package rbac

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestHasPermission(t *testing.T) {
	tests := []struct {
		name           string
		role           string
		permission     Permission
		setupMock      func(sqlmock.Sqlmock)
		wantPermission bool
		wantErr        bool
	}{
		{
			name:       "admin has export permission",
			role:       "admin",
			permission: PermissionRequestExport,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"has_permission"}).AddRow(true)
				mock.ExpectQuery("SELECT EXISTS").
					WithArgs("admin", "request.export").
					WillReturnRows(rows)
			},
			wantPermission: true,
			wantErr:        false,
		},
		{
			name:       "member does not have export permission",
			role:       "member",
			permission: PermissionRequestExport,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"has_permission"}).AddRow(false)
				mock.ExpectQuery("SELECT EXISTS").
					WithArgs("member", "request.export").
					WillReturnRows(rows)
			},
			wantPermission: false,
			wantErr:        false,
		},
		{
			name:       "member has view permission",
			role:       "member",
			permission: PermissionRequestView,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"has_permission"}).AddRow(true)
				mock.ExpectQuery("SELECT EXISTS").
					WithArgs("member", "request.view").
					WillReturnRows(rows)
			},
			wantPermission: true,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.setupMock(mock)

			hasPermission, err := HasPermission(context.Background(), db, tt.role, tt.permission)

			if (err != nil) != tt.wantErr {
				t.Errorf("HasPermission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if hasPermission != tt.wantPermission {
				t.Errorf("HasPermission() = %v, want %v", hasPermission, tt.wantPermission)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestRequirePermission(t *testing.T) {
	tests := []struct {
		name       string
		role       string
		permission Permission
		setupMock  func(sqlmock.Sqlmock)
		wantErr    bool
	}{
		{
			name:       "admin has permission - no error",
			role:       "admin",
			permission: PermissionRequestExport,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"has_permission"}).AddRow(true)
				mock.ExpectQuery("SELECT EXISTS").
					WithArgs("admin", "request.export").
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:       "member lacks permission - error",
			role:       "member",
			permission: PermissionRequestExport,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"has_permission"}).AddRow(false)
				mock.ExpectQuery("SELECT EXISTS").
					WithArgs("member", "request.export").
					WillReturnRows(rows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.setupMock(mock)

			err = RequirePermission(context.Background(), db, tt.role, tt.permission)

			if (err != nil) != tt.wantErr {
				t.Errorf("RequirePermission() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}
