package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"signup/svc"
	"signup/types"
)

// ==================== Activity Logic ====================

type GetActivityListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetActivityListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetActivityListLogic {
	return &GetActivityListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type GetActivityListReq struct {
	CompanyID string `json:"companyId,optional"`
	BranchID string `json:"branchId,optional"`
	Status   string `json:"status,optional"`
	Page     int    `json:"page,default=1"`
	PageSize int    `json:"pageSize,default=20"`
	Keyword  string `json:"keyword,optional"`
}

func (l *GetActivityListLogic) GetActivityList(req *GetActivityListReq) (map[string]interface{}, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	// Build query with filters
	whereClause := "WHERE deleted_at IS NULL"
	args := []interface{}{}
	argNum := 1

	if req.CompanyID != "" {
		whereClause += fmt.Sprintf(" AND company_id = $%d", argNum)
		args = append(args, req.CompanyID)
		argNum++
	}

	if req.BranchID != "" {
		whereClause += fmt.Sprintf(" AND branch_id = $%d", argNum)
		args = append(args, req.BranchID)
		argNum++
	}

	if req.Status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, req.Status)
		argNum++
	}

	if req.Keyword != "" {
		whereClause += fmt.Sprintf(" AND name LIKE $%d", argNum)
		args = append(args, "%"+req.Keyword+"%")
		argNum++
	}

	// Query total count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM activities %s", whereClause)
	l.svcCtx.DB.QueryRow(countQuery, args...).Scan(&total)

	// Query list
	query := fmt.Sprintf(`
		SELECT id, company_id, branch_id, name, description, cover_image, status, start_date, end_date, location, max_participants, current_participants, page_title, page_subtitle, page_description, registration_notice, contact_phone, contact_email, created_by, created_at
		FROM activities 
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)
	args = append(args, req.PageSize, offset)

	rows, err := l.svcCtx.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []types.ActivityInfo
	for rows.Next() {
		var activity types.ActivityInfo
		var companyID, branchID sql.NullString
		var startDate, endDate, coverImage, description, location, pageTitle, pageSubtitle, pageDesc, regNotice, contactPhone, contactEmail sql.NullString
		var maxParticipants, currentParticipants sql.NullInt64
		if err := rows.Scan(&activity.ID, &companyID, &branchID, &activity.Name, &description, &coverImage, &activity.Status,
			&startDate, &endDate, &location, &maxParticipants, &currentParticipants, &pageTitle, &pageSubtitle, &pageDesc, &regNotice, &contactPhone, &contactEmail, &activity.CreatedBy, &activity.CreatedAt); err == nil {
			activity.CompanyID = companyID.String
			activity.BranchID = branchID.String
			activity.Description = description.String
			activity.CoverImage = coverImage.String
			activity.Location = location.String
			activity.PageTitle = pageTitle.String
			activity.PageSubtitle = pageSubtitle.String
			activity.PageDescription = pageDesc.String
			activity.RegistrationNotice = regNotice.String
			activity.ContactPhone = contactPhone.String
			activity.ContactEmail = contactEmail.String
			if startDate.Valid {
				activity.StartDate = startDate.String
			}
			if endDate.Valid {
				activity.EndDate = endDate.String
			}
			if maxParticipants.Valid {
				activity.MaxParticipants = int(maxParticipants.Int64)
			}
			if currentParticipants.Valid {
				activity.CurrentParticipants = int(currentParticipants.Int64)
			}
			activities = append(activities, activity)
		}
	}

	return map[string]interface{}{
		"activities": activities,
		"pagination": types.PaginationResp{
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalCount: total,
			TotalPages: (total + req.PageSize - 1) / req.PageSize,
		},
	}, nil
}

type GetActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetActivityLogic {
	return &GetActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetActivityLogic) GetActivity(id string) (*types.ActivityInfo, error) {
	var activity types.ActivityInfo
	var companyID, branchID sql.NullString
	var startDate, endDate, coverImage, description, content sql.NullString
	var pageTitle, pageSubtitle, pageDescription, registrationNotice, contactPhone, contactEmail sql.NullString
	err := l.svcCtx.DB.QueryRow(`
		SELECT id, company_id, branch_id, name, description, cover_image, content, status, start_date, end_date, 
		       created_by, created_at, updated_at,
		       page_title, page_subtitle, page_description, registration_notice, contact_phone, contact_email
		FROM activities WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(&activity.ID, &companyID, &branchID, &activity.Name, &description, &coverImage, &content, &activity.Status,
		&startDate, &endDate, &activity.CreatedBy, &activity.CreatedAt, &activity.UpdatedAt,
		&pageTitle, &pageSubtitle, &pageDescription, &registrationNotice, &contactPhone, &contactEmail)
	if err != nil {
		return nil, errors.New("活动不存在")
	}
	activity.CompanyID = companyID.String
	activity.BranchID = branchID.String
	activity.Description = description.String
	activity.CoverImage = coverImage.String
	activity.Content = content.String
	if startDate.Valid {
		activity.StartDate = startDate.String
	}
	if endDate.Valid {
		activity.EndDate = endDate.String
	}
	activity.PageTitle = pageTitle.String
	activity.PageSubtitle = pageSubtitle.String
	activity.PageDescription = pageDescription.String
	activity.RegistrationNotice = registrationNotice.String
	activity.ContactPhone = contactPhone.String
	activity.ContactEmail = contactEmail.String
	return &activity, nil
}

type CreateActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateActivityLogic {
	return &CreateActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateActivityLogic) CreateActivity(req *types.CreateActivityReq, userId string) (*types.ActivityInfo, error) {
	if req.Name == "" {
		return nil, errors.New("活动名称不能为空")
	}
	// companyId 和 branchId 至少有一个
	if req.CompanyID == "" && req.BranchID == "" {
		return nil, errors.New("公司或分支机构至少需要指定一个")
	}

	id := uuid.New().String()
	if req.Status == "" {
		req.Status = "active"
	}

	var content interface{}
	if req.Content != "" {
		content = req.Content
	}

	var maxParticipants int
	if req.MaxParticipants > 0 {
		maxParticipants = req.MaxParticipants
	} else {
		maxParticipants = 100
	}

	var startDate, endDate interface{}
	if req.StartDate != "" {
		startDate = req.StartDate
	}
	if req.EndDate != "" {
		endDate = req.EndDate
	}

	_, err := l.svcCtx.DB.Exec(`
		INSERT INTO activities (id, company_id, branch_id, name, description, cover_image, content, status, start_date, end_date, location, max_participants, current_participants, page_title, page_subtitle, page_description, registration_notice, contact_phone, contact_email, created_by, created_at, updated_at)
		VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), $4, $5, $6, $7, $8, $9, $10, $11, $12, 0, $13, $14, $15, $16, $17, $18, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, id, req.CompanyID, req.BranchID, req.Name, req.Description, req.CoverImage, content, req.Status, startDate, endDate, req.Location, maxParticipants, req.PageTitle, req.PageSubtitle, req.PageDescription, req.RegistrationNotice, req.ContactPhone, req.ContactEmail, userId)
	if err != nil {
		return nil, errors.New("创建活动失败: " + err.Error())
	}

	// Create default form fields (name, phone)
	defaultFields := []struct {
		fieldName string
		fieldType string
		required  bool
	}{
		{"姓名", "text", true},
		{"手机号", "text", true},
	}

	for i, field := range defaultFields {
		fieldId := uuid.New().String()
		l.svcCtx.DB.Exec(`
			INSERT INTO form_fields (id, activity_id, field_name, field_type, is_required, sort_order, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
		`, fieldId, id, field.fieldName, field.fieldType, field.required, i+1)
	}

	return &types.ActivityInfo{
		ID:          id,
		CompanyID:   req.CompanyID,
		BranchID:    req.BranchID,
		Name:        req.Name,
		Description: req.Description,
		CoverImage:  req.CoverImage,
		Content:     req.Content,
		Status:      req.Status,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		CreatedBy:   userId,
		CreatedAt:   time.Now().Format("2006-01-02T15:04:05Z"),
	}, nil
}

type UpdateActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateActivityLogic {
	return &UpdateActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateActivityLogic) UpdateActivity(id string, req *types.UpdateActivityReq) (*types.ActivityInfo, error) {
	// Build dynamic update query to avoid COALESCE issues with JSONB
	setClauses := []string{}
	args := []interface{}{}
	argNum := 1

	if req.CompanyID != "" {
		setClauses = append(setClauses, fmt.Sprintf("company_id = $%d", argNum))
		args = append(args, req.CompanyID)
		argNum++
	}
	if req.BranchID != "" {
		setClauses = append(setClauses, fmt.Sprintf("branch_id = $%d", argNum))
		args = append(args, req.BranchID)
		argNum++
	}
	if req.Name != "" {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argNum))
		args = append(args, req.Name)
		argNum++
	}
	if req.Description != "" {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argNum))
		args = append(args, req.Description)
		argNum++
	}
	if req.CoverImage != "" {
		setClauses = append(setClauses, fmt.Sprintf("cover_image = $%d", argNum))
		args = append(args, req.CoverImage)
		argNum++
	}
	if req.Content != "" {
		setClauses = append(setClauses, fmt.Sprintf("content = $%d", argNum))
		args = append(args, req.Content)
		argNum++
	}
	if req.StartDate != "" {
		setClauses = append(setClauses, fmt.Sprintf("start_date = $%d", argNum))
		args = append(args, req.StartDate)
		argNum++
	}
	if req.EndDate != "" {
		setClauses = append(setClauses, fmt.Sprintf("end_date = $%d", argNum))
		args = append(args, req.EndDate)
		argNum++
	}
	if req.Location != "" {
		setClauses = append(setClauses, fmt.Sprintf("location = $%d", argNum))
		args = append(args, req.Location)
		argNum++
	}
	if req.PageTitle != "" {
		setClauses = append(setClauses, fmt.Sprintf("page_title = $%d", argNum))
		args = append(args, req.PageTitle)
		argNum++
	}
	if req.PageSubtitle != "" {
		setClauses = append(setClauses, fmt.Sprintf("page_subtitle = $%d", argNum))
		args = append(args, req.PageSubtitle)
		argNum++
	}
	if req.PageDescription != "" {
		setClauses = append(setClauses, fmt.Sprintf("page_description = $%d", argNum))
		args = append(args, req.PageDescription)
		argNum++
	}
	if req.RegistrationNotice != "" {
		setClauses = append(setClauses, fmt.Sprintf("registration_notice = $%d", argNum))
		args = append(args, req.RegistrationNotice)
		argNum++
	}
	if req.ContactPhone != "" {
		setClauses = append(setClauses, fmt.Sprintf("contact_phone = $%d", argNum))
		args = append(args, req.ContactPhone)
		argNum++
	}
	if req.ContactEmail != "" {
		setClauses = append(setClauses, fmt.Sprintf("contact_email = $%d", argNum))
		args = append(args, req.ContactEmail)
		argNum++
	}
	if req.Status != "" {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argNum))
		args = append(args, req.Status)
		argNum++
	}

	if len(setClauses) == 0 {
		return nil, errors.New("没有需要更新的字段")
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argNum))
	args = append(args, time.Now())
	argNum++

	args = append(args, id)

	query := fmt.Sprintf("UPDATE activities SET %s WHERE id = $%d AND deleted_at IS NULL", strings.Join(setClauses, ", "), argNum)

	result, err := l.svcCtx.DB.Exec(query, args...)
	if err != nil {
		return nil, errors.New("更新活动失败: " + err.Error())
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, errors.New("活动不存在")
	}

	return &types.ActivityInfo{
		ID:          id,
		CompanyID:   req.CompanyID,
		BranchID:    req.BranchID,
		Name:        req.Name,
		Description: req.Description,
		CoverImage:  req.CoverImage,
		Content:     req.Content,
		Status:      req.Status,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Location:    req.Location,
		PageTitle:           req.PageTitle,
		PageSubtitle:        req.PageSubtitle,
		PageDescription:     req.PageDescription,
		RegistrationNotice: req.RegistrationNotice,
		ContactPhone:        req.ContactPhone,
		ContactEmail:        req.ContactEmail,
	}, nil
}

type DeleteActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteActivityLogic {
	return &DeleteActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteActivityLogic) DeleteActivity(id string, userId string) error {
	result, err := l.svcCtx.DB.Exec("UPDATE activities SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1", id)
	if err != nil {
		return errors.New("删除活动失败")
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("活动不存在")
	}
	return nil
}

type UpdateActivityStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateActivityStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateActivityStatusLogic {
	return &UpdateActivityStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateActivityStatusLogic) UpdateActivityStatus(id string, status string, userId string) error {
	validStatuses := map[string]bool{"active": true, "closed": true, "draft": true}
	if !validStatuses[status] {
		return errors.New("无效的状态值")
	}

	result, err := l.svcCtx.DB.Exec(`
		UPDATE activities SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND deleted_at IS NULL
	`, status, id)
	if err != nil {
		return errors.New("更新状态失败")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("活动不存在")
	}
	return nil
}

type GetActivityStatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetActivityStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetActivityStatsLogic {
	return &GetActivityStatsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetActivityStatsLogic) GetActivityStats(activityId string) (map[string]interface{}, error) {
	var total, confirmed, pending, rejected int

	l.svcCtx.DB.QueryRow("SELECT COUNT(*) FROM registrations WHERE activity_id = $1 AND deleted_at IS NULL", activityId).Scan(&total)
	l.svcCtx.DB.QueryRow("SELECT COUNT(*) FROM registrations WHERE activity_id = $1 AND status = 'confirmed' AND deleted_at IS NULL", activityId).Scan(&confirmed)
	l.svcCtx.DB.QueryRow("SELECT COUNT(*) FROM registrations WHERE activity_id = $1 AND status = 'pending' AND deleted_at IS NULL", activityId).Scan(&pending)
	l.svcCtx.DB.QueryRow("SELECT COUNT(*) FROM registrations WHERE activity_id = $1 AND status = 'rejected' AND deleted_at IS NULL", activityId).Scan(&rejected)

	return map[string]interface{}{
		"activityId":      activityId,
		"total":           total,
		"confirmedCount":  confirmed,
		"pendingCount":    pending,
		"rejectedCount":   rejected,
	}, nil
}
