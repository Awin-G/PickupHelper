package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/models"
	"pickup-helper/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- GetUserInfo ---

func TestGetUserInfo_Success(t *testing.T) {
	ur := &mockUserRepo{FindByIDFn: func(_ context.Context, _ repository.DBTX, id int64) (*models.User, error) {
		return &models.User{ID: id, Phone: "13800138000", Nickname: "Alice", UserType: models.UserTypeNormal, CreditScore: 100}, nil
	}}
	svc := NewUserService(ur, &mockRunnerAppRepo{}, nil)

	dto, err := svc.GetUserInfo(context.Background(), 42)
	require.NoError(t, err)
	require.NotNil(t, dto)
	assert.Equal(t, int64(42), dto.ID)
	assert.Equal(t, "138****8000", dto.Phone)
	assert.Equal(t, "Alice", dto.Nickname)
}

func TestGetUserInfo_NotFound(t *testing.T) {
	ur := &mockUserRepo{} // default FindByID returns sql.ErrNoRows
	svc := NewUserService(ur, &mockRunnerAppRepo{}, nil)

	_, err := svc.GetUserInfo(context.Background(), 999)
	requireAppErr(t, err, apperrors.ErrUserNotFound)
}

// --- UpdateUserInfo ---

func TestUpdateUserInfo_Success(t *testing.T) {
	ur := &mockUserRepo{
		FindByIDFn: func(_ context.Context, _ repository.DBTX, id int64) (*models.User, error) {
			return &models.User{ID: id, Phone: "13800138000", Nickname: "Bob", Avatar: "https://x.com/a.png", UserType: models.UserTypeNormal}, nil
		},
	}
	svc := NewUserService(ur, &mockRunnerAppRepo{}, nil)

	dto, err := svc.UpdateUserInfo(context.Background(), 42, "Bob", "https://x.com/a.png")
	require.NoError(t, err)
	require.NotNil(t, dto)
	assert.Equal(t, "Bob", dto.Nickname)
	require.Len(t, ur.UpdateProfileCalls, 1)
	assert.Equal(t, "Bob", ur.UpdateProfileCalls[0].Nickname)
	assert.Equal(t, "https://x.com/a.png", ur.UpdateProfileCalls[0].Avatar)
}

func TestUpdateUserInfo_NicknameTooLong(t *testing.T) {
	svc := NewUserService(&mockUserRepo{}, &mockRunnerAppRepo{}, nil)
	long := strings.Repeat("字", 51) // 51 runes
	_, err := svc.UpdateUserInfo(context.Background(), 1, long, "")
	requireAppErr(t, err, apperrors.ErrNicknameTooLong)
}

func TestUpdateUserInfo_AvatarInvalid(t *testing.T) {
	svc := NewUserService(&mockUserRepo{}, &mockRunnerAppRepo{}, nil)
	_, err := svc.UpdateUserInfo(context.Background(), 1, "ok", "not-a-url")
	requireAppErr(t, err, apperrors.ErrAvatarInvalid)
}

func TestUpdateUserInfo_AvatarEmptyOK(t *testing.T) {
	ur := &mockUserRepo{
		FindByIDFn: func(_ context.Context, _ repository.DBTX, id int64) (*models.User, error) {
			return &models.User{ID: id, Phone: "13800138000", Nickname: "x"}, nil
		},
	}
	svc := NewUserService(ur, &mockRunnerAppRepo{}, nil)
	_, err := svc.UpdateUserInfo(context.Background(), 1, "x", "")
	require.NoError(t, err)
}

// --- ApplyRunner (mock-based; transactional correctness is covered by
// integration tests in test/user_integration_test.go) ---

func TestApplyRunner_IDCardInvalid_Empty(t *testing.T) {
	svc := NewUserService(&mockUserRepo{}, &mockRunnerAppRepo{}, nil)
	_, _, err := svc.ApplyRunner(context.Background(), 1, ApplyRunnerRequest{RealName: "Alice", IDCardImage: ""})
	requireAppErr(t, err, apperrors.ErrIDCardInvalid)
}

func TestApplyRunner_IDCardInvalid_NotURL(t *testing.T) {
	svc := NewUserService(&mockUserRepo{}, &mockRunnerAppRepo{}, nil)
	_, _, err := svc.ApplyRunner(context.Background(), 1, ApplyRunnerRequest{RealName: "Alice", IDCardImage: "ftp://x"})
	requireAppErr(t, err, apperrors.ErrIDCardInvalid)
}

func TestApplyRunner_RealNameEmpty(t *testing.T) {
	svc := NewUserService(&mockUserRepo{}, &mockRunnerAppRepo{}, nil)
	_, _, err := svc.ApplyRunner(context.Background(), 1, ApplyRunnerRequest{RealName: "  ", IDCardImage: "https://x.com/id.jpg"})
	requireAppErr(t, err, apperrors.ErrInvalidParam)
}

func TestApplyRunner_UserNotFound(t *testing.T) {
	ur := &mockUserRepo{} // FindByID returns ErrNoRows
	svc := NewUserService(ur, &mockRunnerAppRepo{}, nil)
	_, _, err := svc.ApplyRunner(context.Background(), 1, ApplyRunnerRequest{RealName: "Alice", IDCardImage: "https://x.com/id.jpg"})
	requireAppErr(t, err, apperrors.ErrUserNotFound)
}

func TestApplyRunner_AlreadyRunner(t *testing.T) {
	ur := &mockUserRepo{FindByIDFn: func(_ context.Context, _ repository.DBTX, id int64) (*models.User, error) {
		return &models.User{ID: id, Phone: "13800138000", UserType: models.UserTypeRunner, RunnerStatus: models.RunnerStatusApproved, CreditScore: 100}, nil
	}}
	svc := NewUserService(ur, &mockRunnerAppRepo{}, nil)
	_, _, err := svc.ApplyRunner(context.Background(), 1, ApplyRunnerRequest{RealName: "Alice", IDCardImage: "https://x.com/id.jpg"})
	requireAppErr(t, err, apperrors.ErrRunnerDuplicate)
}

func TestApplyRunner_Pending(t *testing.T) {
	ur := &mockUserRepo{FindByIDFn: func(_ context.Context, _ repository.DBTX, id int64) (*models.User, error) {
		return &models.User{ID: id, Phone: "13800138000", UserType: models.UserTypeNormal, RunnerStatus: models.RunnerStatusPending, CreditScore: 100}, nil
	}}
	svc := NewUserService(ur, &mockRunnerAppRepo{}, nil)
	_, _, err := svc.ApplyRunner(context.Background(), 1, ApplyRunnerRequest{RealName: "Alice", IDCardImage: "https://x.com/id.jpg"})
	requireAppErr(t, err, apperrors.ErrRunnerDuplicate)
}

func TestApplyRunner_LowCredit(t *testing.T) {
	ur := &mockUserRepo{FindByIDFn: func(_ context.Context, _ repository.DBTX, id int64) (*models.User, error) {
		return &models.User{ID: id, Phone: "13800138000", UserType: models.UserTypeNormal, RunnerStatus: models.RunnerStatusNone, CreditScore: 50}, nil
	}}
	svc := NewUserService(ur, &mockRunnerAppRepo{}, nil)
	_, _, err := svc.ApplyRunner(context.Background(), 1, ApplyRunnerRequest{RealName: "Alice", IDCardImage: "https://x.com/id.jpg"})
	requireAppErr(t, err, apperrors.ErrCreditLow)
}

// --- ListRunnerApps ---

func TestListRunnerApps_Success_PhoneMasked(t *testing.T) {
	apps := []*models.RunnerApplication{
		{ID: 1, UserID: 11, RealName: "A", Status: models.AppStatusPending},
		{ID: 2, UserID: 22, RealName: "B", Status: models.AppStatusPending},
		{ID: 3, UserID: 33, RealName: "C", Status: models.AppStatusPending},
	}
	rr := &mockRunnerAppRepo{
		ListByFilterFn: func(_ context.Context, _ repository.DBTX, _ repository.RunnerAppFilter) ([]*models.RunnerApplication, int64, error) {
			return apps, 3, nil
		},
	}
	ur := &mockUserRepo{FindByIDFn: func(_ context.Context, _ repository.DBTX, id int64) (*models.User, error) {
		phones := map[int64]string{11: "13800138001", 22: "13800138002", 33: "13800138003"}
		return &models.User{ID: id, Phone: phones[id]}, nil
	}}
	svc := NewUserService(ur, rr, nil)

	res, err := svc.ListRunnerApps(context.Background(), RunnerAppListFilter{Offset: 0, Limit: 20})
	require.NoError(t, err)
	assert.Equal(t, int64(3), res.Total)
	require.Len(t, res.Items, 3)
	assert.Equal(t, "138****8001", res.Items[0].Phone)
	assert.Equal(t, "138****8002", res.Items[1].Phone)
	assert.Equal(t, "138****8003", res.Items[2].Phone)
	assert.Equal(t, "pending", res.Items[0].StatusText)
}

func TestListRunnerApps_DefaultLimit(t *testing.T) {
	var captured repository.RunnerAppFilter
	rr := &mockRunnerAppRepo{
		ListByFilterFn: func(_ context.Context, _ repository.DBTX, f repository.RunnerAppFilter) ([]*models.RunnerApplication, int64, error) {
			captured = f
			return nil, 0, nil
		},
	}
	svc := NewUserService(&mockUserRepo{}, rr, nil)
	_, err := svc.ListRunnerApps(context.Background(), RunnerAppListFilter{})
	require.NoError(t, err)
	assert.Equal(t, 20, captured.Limit, "default limit should be 20 when 0/unset")
}

// --- AuditRunnerApp (mock-based; transactional behavior in integration tests) ---

func TestAuditRunnerApp_InvalidAction(t *testing.T) {
	svc := NewUserService(&mockUserRepo{}, &mockRunnerAppRepo{}, nil)
	_, err := svc.AuditRunnerApp(context.Background(), 1, 1, "foo", "")
	requireAppErr(t, err, apperrors.ErrActionInvalid)
}

func TestAuditRunnerApp_NotFound(t *testing.T) {
	rr := &mockRunnerAppRepo{} // FindByID returns ErrNoRows
	svc := NewUserService(&mockUserRepo{}, rr, nil)
	_, err := svc.AuditRunnerApp(context.Background(), 1, 1, "approve", "")
	requireAppErr(t, err, apperrors.ErrAppNotFound)
}

func TestAuditRunnerApp_NotPending(t *testing.T) {
	rr := &mockRunnerAppRepo{FindByIDFn: func(_ context.Context, _ repository.DBTX, _ int64) (*models.RunnerApplication, error) {
		return &models.RunnerApplication{ID: 1, Status: models.AppStatusApproved}, nil
	}}
	svc := NewUserService(&mockUserRepo{}, rr, nil)
	_, err := svc.AuditRunnerApp(context.Background(), 1, 1, "approve", "")
	requireAppErr(t, err, apperrors.ErrAppNotPending)
}

// AuditRunnerApp approve / reject paths exercise WithTx, which requires a
// real *sqlx.DB. Those flows are validated end-to-end in
// test/user_integration_test.go (USER-07/USER-08). Here we just confirm the
// pre-flight checks above reject bad inputs before reaching the transaction.

// --- SetBlacklist ---

func TestSetBlacklist_Success(t *testing.T) {
	ur := &mockUserRepo{FindByIDFn: func(_ context.Context, _ repository.DBTX, _ int64) (*models.User, error) {
		return &models.User{ID: 1, Phone: "13800138000"}, nil
	}}
	svc := NewUserService(ur, &mockRunnerAppRepo{}, nil)

	err := svc.SetBlacklist(context.Background(), 42, true, "")
	require.NoError(t, err)
	require.Len(t, ur.SetBlacklistCalls, 1)
	assert.Equal(t, int64(42), ur.SetBlacklistCalls[0].ID)
	assert.Equal(t, int8(1), ur.SetBlacklistCalls[0].IsBlacklisted)
}

func TestSetBlacklist_Unblacklist(t *testing.T) {
	ur := &mockUserRepo{FindByIDFn: func(_ context.Context, _ repository.DBTX, _ int64) (*models.User, error) {
		return &models.User{ID: 1, Phone: "13800138000", IsBlacklisted: 1}, nil
	}}
	svc := NewUserService(ur, &mockRunnerAppRepo{}, nil)

	err := svc.SetBlacklist(context.Background(), 42, false, "")
	require.NoError(t, err)
	require.Len(t, ur.SetBlacklistCalls, 1)
	assert.Equal(t, int8(0), ur.SetBlacklistCalls[0].IsBlacklisted)
}

func TestSetBlacklist_NotFound(t *testing.T) {
	ur := &mockUserRepo{} // FindByID returns ErrNoRows
	svc := NewUserService(ur, &mockRunnerAppRepo{}, nil)
	err := svc.SetBlacklist(context.Background(), 999, true, "")
	requireAppErr(t, err, apperrors.ErrUserNotFound)
}

func TestSetBlacklist_RepoError(t *testing.T) {
	ur := &mockUserRepo{
		FindByIDFn: func(_ context.Context, _ repository.DBTX, _ int64) (*models.User, error) {
			return &models.User{ID: 1, Phone: "13800138000"}, nil
		},
		SetBlacklistFn: func(_ context.Context, _ repository.DBTX, _ int64, _ int8) error {
			return errors.New("db down")
		},
	}
	svc := NewUserService(ur, &mockRunnerAppRepo{}, nil)
	err := svc.SetBlacklist(context.Background(), 1, true, "")
	require.Error(t, err)
	// Wrapped as ErrInternal.
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperrors.ErrInternal, ae.Code)
}

// --- Admin user CRUD ---

func TestCreateUser_Success(t *testing.T) {
	ur := &mockUserRepo{
		CreateAdminFn: func(_ context.Context, _ repository.DBTX, phone, nickname string, userType int8) (int64, error) {
			return 42, nil
		},
		FindByIDFn: func(_ context.Context, _ repository.DBTX, id int64) (*models.User, error) {
			return &models.User{ID: id, Phone: "13800138000", Nickname: "alice", UserType: models.UserTypeNormal, CreditScore: 100}, nil
		},
	}
	svc := NewUserService(ur, &mockRunnerAppRepo{}, nil)
	dto, err := svc.CreateUser(context.Background(), CreateUserRequest{Phone: "13800138000", Nickname: "alice", UserType: models.UserTypeNormal})
	require.NoError(t, err)
	assert.Equal(t, int64(42), dto.ID)
	assert.Equal(t, "13800138000", dto.Phone)
}

func TestCreateUser_InvalidPhone(t *testing.T) {
	svc := NewUserService(&mockUserRepo{}, &mockRunnerAppRepo{}, nil)
	_, err := svc.CreateUser(context.Background(), CreateUserRequest{Phone: "123", Nickname: "x"})
	requireAppErr(t, err, apperrors.ErrPhoneFormat)
}

func TestCreateUser_NicknameTooLong(t *testing.T) {
	svc := NewUserService(&mockUserRepo{}, &mockRunnerAppRepo{}, nil)
	long := strings.Repeat("字", 51)
	_, err := svc.CreateUser(context.Background(), CreateUserRequest{Phone: "13800138000", Nickname: long})
	requireAppErr(t, err, apperrors.ErrNicknameTooLong)
}

func TestCreateUser_DuplicatePhone(t *testing.T) {
	ur := &mockUserRepo{
		CreateAdminFn: func(_ context.Context, _ repository.DBTX, _, _ string, _ int8) (int64, error) {
			return 0, errors.New("Duplicate entry '13800138000' for key 'phone'")
		},
	}
	svc := NewUserService(ur, &mockRunnerAppRepo{}, nil)
	_, err := svc.CreateUser(context.Background(), CreateUserRequest{Phone: "13800138000", Nickname: "x"})
	requireAppErr(t, err, apperrors.ErrPhoneExists)
}

func TestGetUserDetail_Success(t *testing.T) {
	ur := &mockUserRepo{FindByIDFn: func(_ context.Context, _ repository.DBTX, id int64) (*models.User, error) {
		return &models.User{ID: id, Phone: "13800138000", Nickname: "alice", UserType: models.UserTypeNormal, CreditScore: 100}, nil
	}}
	svc := NewUserService(ur, &mockRunnerAppRepo{}, nil)
	dto, err := svc.GetUserDetail(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "13800138000", dto.Phone) // admin sees raw phone
}

func TestGetUserDetail_NotFound(t *testing.T) {
	svc := NewUserService(&mockUserRepo{}, &mockRunnerAppRepo{}, nil)
	_, err := svc.GetUserDetail(context.Background(), 999)
	requireAppErr(t, err, apperrors.ErrUserNotFound)
}

func TestListUsers_Success(t *testing.T) {
	ur := &mockUserRepo{ListUsersFn: func(_ context.Context, _ repository.DBTX, _ repository.UserListFilter) ([]*models.User, int64, error) {
		return []*models.User{
			{ID: 1, Phone: "13800138000", Nickname: "alice"},
			{ID: 2, Phone: "13900139000", Nickname: "bob"},
		}, 2, nil
	}}
	svc := NewUserService(ur, &mockRunnerAppRepo{}, nil)
	res, err := svc.ListUsers(context.Background(), AdminUserListFilter{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Equal(t, int64(2), res.Total)
	require.Len(t, res.Items, 2)
	assert.Equal(t, "13800138000", res.Items[0].Phone) // unmasked
}

func TestUpdateUser_Success(t *testing.T) {
	current := &models.User{ID: 1, Phone: "13800138000", Nickname: "old", UserType: models.UserTypeNormal, CreditScore: 100}
	ur := &mockUserRepo{
		FindByIDFn: func(_ context.Context, _ repository.DBTX, id int64) (*models.User, error) {
			return current, nil
		},
		UpdateUserFn: func(_ context.Context, _ repository.DBTX, _ int64, cols []string, args []any) error {
			for i, col := range cols {
				switch col {
				case "nickname":
					current.Nickname = args[i].(string)
				}
			}
			return nil
		},
	}
	svc := NewUserService(ur, &mockRunnerAppRepo{}, nil)
	newNick := "new-name"
	dto, err := svc.UpdateUser(context.Background(), 1, UpdateUserRequest{Nickname: &newNick})
	require.NoError(t, err)
	assert.Equal(t, "new-name", dto.Nickname)
}

func TestUpdateUser_NotFound(t *testing.T) {
	svc := NewUserService(&mockUserRepo{}, &mockRunnerAppRepo{}, nil)
	n := "x"
	_, err := svc.UpdateUser(context.Background(), 999, UpdateUserRequest{Nickname: &n})
	requireAppErr(t, err, apperrors.ErrUserNotFound)
}

func TestDeleteUser_Success(t *testing.T) {
	ur := &mockUserRepo{DeleteUserFn: func(_ context.Context, _ repository.DBTX, _ int64) error {
		return nil
	}}
	svc := NewUserService(ur, &mockRunnerAppRepo{}, nil)
	err := svc.DeleteUser(context.Background(), 1)
	require.NoError(t, err)
}

func TestDeleteUser_NotFound(t *testing.T) {
	ur := &mockUserRepo{DeleteUserFn: func(_ context.Context, _ repository.DBTX, _ int64) error {
		return sql.ErrNoRows
	}}
	svc := NewUserService(ur, &mockRunnerAppRepo{}, nil)
	err := svc.DeleteUser(context.Background(), 999)
	requireAppErr(t, err, apperrors.ErrUserNotFound)
}
