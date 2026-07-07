package service

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"strings"

	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/models"
	"pickup-helper/internal/repository"

	"github.com/jmoiron/sqlx"
)

// ApplyRunnerRequest is the input for UserService.ApplyRunner.
type ApplyRunnerRequest struct {
	RealName    string `json:"real_name"`
	StudentID   string `json:"student_id"`
	IDCardImage string `json:"id_card_image"`
}

// RunnerAppDTO is the API representation of a runner application. Phone is
// masked; student_id / id_card_image are surfaced as-is for admin auditors.
type RunnerAppDTO struct {
	ID           int64  `json:"id"`
	UserID       int64  `json:"user_id"`
	RealName     string `json:"real_name"`
	Phone        string `json:"phone"` // masked
	StudentID    string `json:"student_id,omitempty"`
	IDCardImage  string `json:"id_card_image,omitempty"`
	Status       int8   `json:"status"`
	StatusText   string `json:"status_text"`
	AuditRemark  string `json:"audit_remark,omitempty"`
	CreatedAt    string `json:"created_at"`
}

// RunnerAppListFilter is the service-layer filter for ListRunnerApps.
// Status==nil means no status filter. Keyword matches real_name or phone.
type RunnerAppListFilter struct {
	Status  *int8
	Keyword string
	Offset  int
	Limit   int
}

// RunnerAppListResult bundles the page + total for ListRunnerApps.
type RunnerAppListResult struct {
	Items []*RunnerAppDTO `json:"items"`
	Total int64           `json:"total"`
}

// UserService implements user-profile, runner-application, audit, and
// blacklist operations. Cross-table mutations (ApplyRunner / AuditRunnerApp)
// are wrapped in transactions via repository.WithTx.
type UserService struct {
	userRepo   repository.UserRepo
	runnerRepo repository.RunnerAppRepo
	db         *sqlx.DB
}

// NewUserService wires up a UserService.
func NewUserService(ur repository.UserRepo, rr repository.RunnerAppRepo, db *sqlx.DB) *UserService {
	return &UserService{userRepo: ur, runnerRepo: rr, db: db}
}

// GetUserInfo returns the masked DTO for the given user id.
func (s *UserService) GetUserInfo(ctx context.Context, userID int64) (*models.UserInfoDTO, error) {
	u, err := s.userRepo.FindByID(ctx, s.db, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.ErrUserNotFound, "")
	}
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	return u.ToDTO(), nil
}

// UpdateUserInfo validates and applies profile updates. Nickname is capped
// at 50 chars; avatar must be http(s):// when non-empty.
func (s *UserService) UpdateUserInfo(ctx context.Context, userID int64, nickname, avatar string) (*models.UserInfoDTO, error) {
	if len([]rune(nickname)) > 50 {
		return nil, apperrors.New(apperrors.ErrNicknameTooLong, "")
	}
	if avatar != "" && !isValidHTTPURL(avatar) {
		return nil, apperrors.New(apperrors.ErrAvatarInvalid, "")
	}
	if err := s.userRepo.UpdateProfile(ctx, s.db, userID, nickname, avatar); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	return s.GetUserInfo(ctx, userID)
}

// ApplyRunner creates a runner application for the user and flips their
// runner_status to pending, atomically. Returns (applicationID, status).
func (s *UserService) ApplyRunner(ctx context.Context, userID int64, req ApplyRunnerRequest) (int64, int8, error) {
	realName := strings.TrimSpace(req.RealName)
	if realName == "" || len([]rune(realName)) > 50 {
		return 0, 0, apperrors.New(apperrors.ErrInvalidParam, "real_name 非法")
	}
	idCard := strings.TrimSpace(req.IDCardImage)
	if idCard == "" || !isValidHTTPURL(idCard) {
		return 0, 0, apperrors.New(apperrors.ErrIDCardInvalid, "")
	}

	user, err := s.userRepo.FindByID(ctx, s.db, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, apperrors.New(apperrors.ErrUserNotFound, "")
	}
	if err != nil {
		return 0, 0, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	// Eligibility: must be a normal user not currently pending or already a runner.
	if user.UserType != models.UserTypeNormal ||
		user.RunnerStatus == models.RunnerStatusPending ||
		user.RunnerStatus == models.RunnerStatusApproved ||
		user.UserType == models.UserTypeRunner {
		return 0, 0, apperrors.New(apperrors.ErrRunnerDuplicate, "")
	}
	if user.CreditScore < 60 {
		return 0, 0, apperrors.New(apperrors.ErrCreditLow, "")
	}

	app := &models.RunnerApplication{
		UserID:      userID,
		RealName:    realName,
		StudentID:   nullableStr(strings.TrimSpace(req.StudentID)),
		IDCardImage: nullableStr(idCard),
		Status:      models.AppStatusPending,
	}

	var appID int64
	err = repository.WithTx(ctx, s.db, func(tx *sqlx.Tx) error {
		id, e := s.runnerRepo.Create(ctx, tx, app)
		if e != nil {
			return e
		}
		appID = id
		return s.userRepo.UpdateRunnerStatus(ctx, tx, userID,
			models.UserTypeNormal, models.RunnerStatusPending)
	})
	if err != nil {
		return 0, 0, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	return appID, models.AppStatusPending, nil
}

// ListRunnerApps returns a paginated list of runner applications for admin
// auditors. Each item carries the masked phone of the applicant.
func (s *UserService) ListRunnerApps(ctx context.Context, filter RunnerAppListFilter) (*RunnerAppListResult, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	repoFilter := repository.RunnerAppFilter{
		Status:  filter.Status,
		Keyword: strings.TrimSpace(filter.Keyword),
		Offset:  filter.Offset,
		Limit:   limit,
	}
	apps, total, err := s.runnerRepo.ListByFilter(ctx, s.db, repoFilter)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	// Fetch applicant phones in one pass for masking. The repo already
	// JOINs users for keyword filtering but does not surface phone in the
	// row — keep the query surface minimal and look up per-applicant.
	// For lists of ≤20 items this is acceptable; for larger pages a batch
	// lookup should be added (v2 optimization).
	phoneCache := map[int64]string{}
	items := make([]*RunnerAppDTO, 0, len(apps))
	for _, a := range apps {
		phone, ok := phoneCache[a.UserID]
		if !ok {
			if u, e := s.userRepo.FindByID(ctx, s.db, a.UserID); e == nil {
				phone = u.Phone
			}
			phoneCache[a.UserID] = phone
		}
		items = append(items, toRunnerAppDTO(a, phone))
	}
	return &RunnerAppListResult{Items: items, Total: total}, nil
}

// AuditRunnerApp approves or rejects a pending runner application. Approval
// flips user_type=2 / runner_status=2; rejection sets runner_status=3
// (user_type unchanged). Both writes run in a single transaction.
func (s *UserService) AuditRunnerApp(ctx context.Context, appID, adminID int64, action, auditRemark string) (*RunnerAppDTO, error) {
	action = strings.ToLower(strings.TrimSpace(action))
	if action != "approve" && action != "reject" {
		return nil, apperrors.New(apperrors.ErrActionInvalid, "")
	}

	app, err := s.runnerRepo.FindByID(ctx, s.db, appID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.ErrAppNotFound, "")
	}
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	if app.Status != models.AppStatusPending {
		return nil, apperrors.New(apperrors.ErrAppNotPending, "")
	}

	var newStatus int8
	var newUserType, newRunnerStatus int8
	if action == "approve" {
		newStatus = models.AppStatusApproved
		newUserType = models.UserTypeRunner
		newRunnerStatus = models.RunnerStatusApproved
	} else {
		newStatus = models.AppStatusRejected
		newUserType = models.UserTypeNormal
		newRunnerStatus = models.RunnerStatusRejected
	}

	err = repository.WithTx(ctx, s.db, func(tx *sqlx.Tx) error {
		if e := s.runnerRepo.UpdateStatus(ctx, tx, appID, adminID, newStatus, auditRemark); e != nil {
			return e
		}
		return s.userRepo.UpdateRunnerStatus(ctx, tx, app.UserID, newUserType, newRunnerStatus)
	})
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	updated, err := s.runnerRepo.FindByID(ctx, s.db, appID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	phone := ""
	if u, e := s.userRepo.FindByID(ctx, s.db, updated.UserID); e == nil {
		phone = u.Phone
	}
	return toRunnerAppDTO(updated, phone), nil
}

// SetBlacklist toggles a user's blacklist flag. v=true → blacklist.
func (s *UserService) SetBlacklist(ctx context.Context, userID int64, isBlacklisted bool, _ string) error {
	if _, err := s.userRepo.FindByID(ctx, s.db, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.New(apperrors.ErrUserNotFound, "")
		}
		return apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	flag := int8(0)
	if isBlacklisted {
		flag = 1
	}
	if err := s.userRepo.SetBlacklist(ctx, s.db, userID, flag); err != nil {
		return apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	return nil
}

// isValidHTTPURL reports whether s is an http(s):// URL with a non-empty host.
func isValidHTTPURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// nullableStr converts "" to sql.NullString{Valid:false}.
func nullableStr(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// toRunnerAppDTO maps a RunnerApplication row to its DTO, masking phone.
func toRunnerAppDTO(a *models.RunnerApplication, phone string) *RunnerAppDTO {
	dto := &RunnerAppDTO{
		ID:          a.ID,
		UserID:      a.UserID,
		RealName:    a.RealName,
		Phone:       models.MaskPhone(phone),
		Status:      a.Status,
		StatusText:  appStatusText(a.Status),
		CreatedAt:   a.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if a.StudentID.Valid {
		dto.StudentID = a.StudentID.String
	}
	if a.IDCardImage.Valid {
		dto.IDCardImage = a.IDCardImage.String
	}
	if a.AuditRemark.Valid {
		dto.AuditRemark = a.AuditRemark.String
	}
	return dto
}

func appStatusText(s int8) string {
	switch s {
	case models.AppStatusPending:
		return "pending"
	case models.AppStatusApproved:
		return "approved"
	case models.AppStatusRejected:
		return "rejected"
	}
	return "unknown"
}
