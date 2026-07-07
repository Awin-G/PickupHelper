package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"pickup-helper/internal/models"

	"github.com/jmoiron/sqlx"
)

// UserRepo abstracts persistence for the users table. Methods accept a
// DBTX so the service layer can run them inside a transaction (see
// repository.WithTx). When called outside a tx, pass the *sqlx.DB which
// also implements DBTX.
type UserRepo interface {
	FindByPhone(ctx context.Context, db DBTX, phone string) (*models.User, error)
	FindByID(ctx context.Context, db DBTX, id int64) (*models.User, error)
	FindByOpenID(ctx context.Context, db DBTX, openid string) (*models.User, error)
	Create(ctx context.Context, db DBTX, phone, openid string) (int64, error)
	UpdateProfile(ctx context.Context, db DBTX, id int64, nickname, avatar string) error
	UpdateRunnerStatus(ctx context.Context, db DBTX, id int64, userType, runnerStatus int8) error
	SetBlacklist(ctx context.Context, db DBTX, id int64, isBlacklisted int8) error
	UpdateOpenID(ctx context.Context, db DBTX, id int64, openid string) error
}

// AdminRepo abstracts persistence for the admins table.
type AdminRepo interface {
	FindByUsername(ctx context.Context, db DBTX, username string) (*models.Admin, error)
	UpdateLastLogin(ctx context.Context, db DBTX, id int64, lastLogin any) error
}

// RunnerAppRepo abstracts persistence for the runner_applications table.
type RunnerAppRepo interface {
	Create(ctx context.Context, db DBTX, app *models.RunnerApplication) (int64, error)
	FindByID(ctx context.Context, db DBTX, id int64) (*models.RunnerApplication, error)
	ListByFilter(ctx context.Context, db DBTX, filter RunnerAppFilter) ([]*models.RunnerApplication, int64, error)
	UpdateStatus(ctx context.Context, db DBTX, id, auditAdminID int64, status int8, auditRemark string) error
}

// RunnerAppFilter holds optional filters for listing runner applications.
// Status is a pointer so callers can distinguish "no filter" (nil) from
// "filter status=0" (which is unused today but reserved).
type RunnerAppFilter struct {
	Status  *int8
	Keyword string // matches real_name (runner_applications) or phone (users)
	Offset  int
	Limit   int
}

// mysqlUserRepo implements UserRepo against *sqlx.DB.
type mysqlUserRepo struct{}

// NewUserRepo returns a stateless UserRepo. The connection is supplied
// per-call via DBTX so a single instance is safe for concurrent use.
func NewUserRepo() UserRepo { return &mysqlUserRepo{} }

func (r *mysqlUserRepo) FindByPhone(ctx context.Context, db DBTX, phone string) (*models.User, error) {
	var u models.User
	err := db.GetContext(ctx, &u,
		`SELECT id, phone, nickname, avatar, openid, user_type, runner_status,
		        credit_score, is_blacklisted, created_at, updated_at
		 FROM users WHERE phone = ?`, phone)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("user_repo.FindByPhone: %w", err)
	}
	return &u, nil
}

func (r *mysqlUserRepo) FindByID(ctx context.Context, db DBTX, id int64) (*models.User, error) {
	var u models.User
	err := db.GetContext(ctx, &u,
		`SELECT id, phone, nickname, avatar, openid, user_type, runner_status,
		        credit_score, is_blacklisted, created_at, updated_at
		 FROM users WHERE id = ?`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("user_repo.FindByID: %w", err)
	}
	return &u, nil
}

func (r *mysqlUserRepo) FindByOpenID(ctx context.Context, db DBTX, openid string) (*models.User, error) {
	var u models.User
	err := db.GetContext(ctx, &u,
		`SELECT id, phone, nickname, avatar, openid, user_type, runner_status,
		        credit_score, is_blacklisted, created_at, updated_at
		 FROM users WHERE openid = ?`, openid)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("user_repo.FindByOpenID: %w", err)
	}
	return &u, nil
}

// Create inserts a new user with INSERT IGNORE so duplicate phone does
// not error. Returns the new id, or the existing id if the phone was
// already present (LastInsertId()==0 in that case).
func (r *mysqlUserRepo) Create(ctx context.Context, db DBTX, phone, openid string) (int64, error) {
	res, err := db.ExecContext(ctx,
		`INSERT IGNORE INTO users (phone, openid, user_type, runner_status, credit_score, is_blacklisted)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		phone, nullable(openid), models.UserTypeNormal, models.RunnerStatusNone, 100, 0)
	if err != nil {
		return 0, fmt.Errorf("user_repo.Create: %w", err)
	}
	if id, err := res.LastInsertId(); err == nil && id > 0 {
		return id, nil
	}
	// INSERT IGNORE on duplicate key → LastInsertId()==0; look up the row.
	u, err := r.FindByPhone(ctx, db, phone)
	if err != nil {
		return 0, fmt.Errorf("user_repo.Create lookup: %w", err)
	}
	return u.ID, nil
}

func (r *mysqlUserRepo) UpdateProfile(ctx context.Context, db DBTX, id int64, nickname, avatar string) error {
	_, err := db.ExecContext(ctx,
		`UPDATE users SET nickname = ?, avatar = ? WHERE id = ?`,
		nickname, avatar, id)
	if err != nil {
		return fmt.Errorf("user_repo.UpdateProfile: %w", err)
	}
	return nil
}

func (r *mysqlUserRepo) UpdateRunnerStatus(ctx context.Context, db DBTX, id int64, userType, runnerStatus int8) error {
	_, err := db.ExecContext(ctx,
		`UPDATE users SET user_type = ?, runner_status = ? WHERE id = ?`,
		userType, runnerStatus, id)
	if err != nil {
		return fmt.Errorf("user_repo.UpdateRunnerStatus: %w", err)
	}
	return nil
}

func (r *mysqlUserRepo) SetBlacklist(ctx context.Context, db DBTX, id int64, isBlacklisted int8) error {
	_, err := db.ExecContext(ctx,
		`UPDATE users SET is_blacklisted = ? WHERE id = ?`,
		isBlacklisted, id)
	if err != nil {
		return fmt.Errorf("user_repo.SetBlacklist: %w", err)
	}
	return nil
}

func (r *mysqlUserRepo) UpdateOpenID(ctx context.Context, db DBTX, id int64, openid string) error {
	_, err := db.ExecContext(ctx,
		`UPDATE users SET openid = ? WHERE id = ?`,
		nullable(openid), id)
	if err != nil {
		return fmt.Errorf("user_repo.UpdateOpenID: %w", err)
	}
	return nil
}

// mysqlAdminRepo implements AdminRepo.
type mysqlAdminRepo struct{}

func NewAdminRepo() AdminRepo { return &mysqlAdminRepo{} }

func (r *mysqlAdminRepo) FindByUsername(ctx context.Context, db DBTX, username string) (*models.Admin, error) {
	var a models.Admin
	err := db.GetContext(ctx, &a,
		`SELECT id, username, password_hash, role_id, station_id, real_name,
		        phone, status, last_login, created_at, updated_at
		 FROM admins WHERE username = ?`, username)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("admin_repo.FindByUsername: %w", err)
	}
	return &a, nil
}

func (r *mysqlAdminRepo) UpdateLastLogin(ctx context.Context, db DBTX, id int64, lastLogin any) error {
	// Accept either time.Time or sql.NullTime to keep the interface simple.
	var arg any
	switch v := lastLogin.(type) {
	case nil:
		arg = sql.NullTime{}
	case sql.NullTime:
		arg = v
	default:
		// assume time.Time-like; pass through.
		arg = v
	}
	_, err := db.ExecContext(ctx,
		`UPDATE admins SET last_login = ? WHERE id = ?`, arg, id)
	if err != nil {
		return fmt.Errorf("admin_repo.UpdateLastLogin: %w", err)
	}
	return nil
}

// mysqlRunnerAppRepo implements RunnerAppRepo.
type mysqlRunnerAppRepo struct{}

func NewRunnerAppRepo() RunnerAppRepo { return &mysqlRunnerAppRepo{} }

func (r *mysqlRunnerAppRepo) Create(ctx context.Context, db DBTX, app *models.RunnerApplication) (int64, error) {
	res, err := db.ExecContext(ctx,
		`INSERT INTO runner_applications
		   (user_id, real_name, student_id, id_card_image, status, audit_admin_id, audit_remark)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		app.UserID, app.RealName, app.StudentID, app.IDCardImage, app.Status,
		app.AuditAdminID, app.AuditRemark)
	if err != nil {
		return 0, fmt.Errorf("runner_app_repo.Create: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("runner_app_repo.Create LastInsertId: %w", err)
	}
	return id, nil
}

func (r *mysqlRunnerAppRepo) FindByID(ctx context.Context, db DBTX, id int64) (*models.RunnerApplication, error) {
	var a models.RunnerApplication
	err := db.GetContext(ctx, &a,
		`SELECT id, user_id, real_name, student_id, id_card_image, status,
		        audit_admin_id, audit_remark, created_at, updated_at
		 FROM runner_applications WHERE id = ?`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("runner_app_repo.FindByID: %w", err)
	}
	return &a, nil
}

func (r *mysqlRunnerAppRepo) ListByFilter(ctx context.Context, db DBTX, filter RunnerAppFilter) ([]*models.RunnerApplication, int64, error) {
	where, args := buildAppWhere(filter)
	// COUNT query — LEFT JOIN users so keyword filter on phone works.
	countQ := "SELECT COUNT(*) FROM runner_applications a LEFT JOIN users u ON a.user_id = u.id"
	if where != "" {
		countQ += " WHERE " + where
	}
	var total int64
	if err := db.GetContext(ctx, &total, countQ, args...); err != nil {
		return nil, 0, fmt.Errorf("runner_app_repo.ListByFilter count: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	listQ := `SELECT a.id, a.user_id, a.real_name, a.student_id, a.id_card_image,
	                 a.status, a.audit_admin_id, a.audit_remark, a.created_at, a.updated_at
	          FROM runner_applications a
	          LEFT JOIN users u ON a.user_id = u.id`
	if where != "" {
		listQ += " WHERE " + where
	}
	listQ += " ORDER BY a.created_at DESC LIMIT ? OFFSET ?"
	listArgs := append(args, limit, filter.Offset)

	var apps []*models.RunnerApplication
	if err := db.SelectContext(ctx, &apps, listQ, listArgs...); err != nil {
		return nil, 0, fmt.Errorf("runner_app_repo.ListByFilter list: %w", err)
	}
	return apps, total, nil
}

func (r *mysqlRunnerAppRepo) UpdateStatus(ctx context.Context, db DBTX, id, auditAdminID int64, status int8, auditRemark string) error {
	_, err := db.ExecContext(ctx,
		`UPDATE runner_applications
		    SET status = ?, audit_admin_id = ?, audit_remark = ?
		  WHERE id = ?`,
		status, sql.NullInt64{Int64: auditAdminID, Valid: auditAdminID > 0},
		sql.NullString{String: auditRemark, Valid: auditRemark != ""}, id)
	if err != nil {
		return fmt.Errorf("runner_app_repo.UpdateStatus: %w", err)
	}
	return nil
}

// buildAppWhere constructs the WHERE clause (without the "WHERE" keyword)
// for ListByFilter. Returns ("", nil) when no filters are set.
func buildAppWhere(filter RunnerAppFilter) (string, []any) {
	var conds []string
	var args []any
	if filter.Status != nil {
		conds = append(conds, "a.status = ?")
		args = append(args, *filter.Status)
	}
	if kw := strings.TrimSpace(filter.Keyword); kw != "" {
		like := "%" + kw + "%"
		conds = append(conds, "(a.real_name LIKE ? OR u.phone LIKE ?)")
		args = append(args, like, like)
	}
	if len(conds) == 0 {
		return "", nil
	}
	return strings.Join(conds, " AND "), args
}

// nullable converts an empty string to sql.NullString so INSERT/UPDATE
// statements write NULL rather than "" for nullable columns.
func nullable(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// Ensure *sqlx.DB and *sqlx.Tx satisfy DBTX at compile time.
var (
	_ DBTX = (*sqlx.DB)(nil)
	_ DBTX = (*sqlx.Tx)(nil)
)
