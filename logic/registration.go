package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"signup/svc"
	"signup/types"
)

type GetRegistrationListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRegistrationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRegistrationListLogic {
	return &GetRegistrationListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRegistrationListLogic) GetRegistrationList(req *types.GetRegistrationListReq) (map[string]interface{}, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	registrations := []types.RegistrationInfo{}
	if l.svcCtx.DB != nil {
		// Check if registrations table has activity_id column (v2 schema)
		var hasActivityCol bool
		l.svcCtx.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name='registrations' AND column_name='activity_id')`).Scan(&hasActivityCol)

		if hasActivityCol {
			// Use the v2 schema query
			query := `SELECT r.id, COALESCE(r.activity_id, r.event_id, ''), COALESCE(r.sales_id, ''), COALESCE(a.name, ''), COALESCE(r.status, 'pending'), COALESCE(r.form_data::text, ''), r.created_at, r.updated_at FROM registrations r LEFT JOIN activities a ON r.activity_id = a.id WHERE r.deleted_at IS NULL`
			countQuery := `SELECT COUNT(*) FROM registrations r WHERE r.deleted_at IS NULL`
			args := []interface{}{}
			countArgs := []interface{}{}
			argNum := 1

			if req.ActivityID != "" {
				query += fmt.Sprintf(" AND r.activity_id = $%d", argNum)
				countQuery += fmt.Sprintf(" AND r.activity_id = $%d", argNum)
				args = append(args, req.ActivityID)
				countArgs = append(countArgs, req.ActivityID)
				argNum++
			}

			// sales角色只能看自己名下的报名
			if req.SalesID != "" {
				query += fmt.Sprintf(" AND r.sales_id = $%d", argNum)
				countQuery += fmt.Sprintf(" AND r.sales_id = $%d", argNum)
				args = append(args, req.SalesID)
				countArgs = append(countArgs, req.SalesID)
				argNum++
			}

			if req.Status != "" {
				query += fmt.Sprintf(" AND r.status = $%d", argNum)
				countQuery += fmt.Sprintf(" AND r.status = $%d", argNum)
				args = append(args, req.Status)
				countArgs = append(countArgs, req.Status)
				argNum++
			}

			// Get total count
			var total int
			l.svcCtx.DB.QueryRow(countQuery, countArgs...).Scan(&total)

			query += fmt.Sprintf(" ORDER BY r.created_at DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)
			args = append(args, req.PageSize, (req.Page-1)*req.PageSize)

			rows, err := l.svcCtx.DB.Query(query, args...)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var r types.RegistrationInfo
					var activityId, salesId, activityName, status, formData, createdAt, updatedAt string
					if err := rows.Scan(&r.ID, &activityId, &salesId, &activityName, &status, &formData, &createdAt, &updatedAt); err == nil {
						r.EventID = activityId
						r.UserID = salesId
						r.Status = status
						r.CreatedAt = createdAt
						r.UpdatedAt = updatedAt
						// Try to parse form_data for name and phone
						if formData != "" {
							var fd map[string]interface{}
							if json.Unmarshal([]byte(formData), &fd) == nil {
								if v, ok := fd["姓名"].(string); ok {
									r.FullName = v
								}
								if v, ok := fd["name"].(string); ok {
									r.FullName = v
								}
								if v, ok := fd["phone"].(string); ok {
									r.Phone = v
								}
								if v, ok := fd["手机号"].(string); ok {
									r.Phone = v
								}
								if v, ok := fd["email"].(string); ok {
									r.Email = v
								}
								if v, ok := fd["备注"].(string); ok {
									r.Remarks = v
								}
							}
						}
						registrations = append(registrations, r)
					}
				}
			}

			return map[string]interface{}{
				"registrations": registrations,
				"pagination": map[string]interface{}{
					"page":       req.Page,
					"pageSize":   req.PageSize,
					"totalCount": total,
					"totalPages": (total + req.PageSize - 1) / req.PageSize,
				},
			}, nil
		}

		// Fallback: use the original schema (v1)
		query := `SELECT id, event_id, user_id, COALESCE(full_name, ''), COALESCE(email, ''), COALESCE(phone, ''), COALESCE(status, 'pending'), COALESCE(remarks, ''), created_at, updated_at FROM registrations WHERE 1=1`
		args := []interface{}{}
		argNum := 1

		if req.ActivityID != "" {
			query += fmt.Sprintf(" AND event_id = $%d", argNum)
			args = append(args, req.ActivityID)
			argNum++
		}
		if req.Status != "" {
			query += fmt.Sprintf(" AND status = $%d", argNum)
			args = append(args, req.Status)
			argNum++
		}

		// sales角色只能看自己提交的报名
		if req.SalesID != "" {
			query += fmt.Sprintf(" AND user_id = $%d", argNum)
			args = append(args, req.SalesID)
			argNum++
		}

		query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)
		args = append(args, req.PageSize, (req.Page-1)*req.PageSize)

		rows, err := l.svcCtx.DB.Query(query, args...)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var r types.RegistrationInfo
				var fullName, email, phone, status, remarks, createdAt, updatedAt string
				if err := rows.Scan(&r.ID, &r.EventID, &r.UserID, &fullName, &email, &phone, &status, &remarks, &createdAt, &updatedAt); err == nil {
					r.FullName = fullName
					r.Email = email
					r.Phone = phone
					r.Status = status
					r.Remarks = remarks
					r.CreatedAt = createdAt
					r.UpdatedAt = updatedAt
					registrations = append(registrations, r)
				}
			}
		}
	}

	return map[string]interface{}{
		"registrations": registrations,
		"pagination": map[string]interface{}{
			"page":       req.Page,
			"pageSize":   req.PageSize,
			"totalCount": len(registrations),
			"totalPages": (len(registrations) + req.PageSize - 1) / req.PageSize,
		},
	}, nil
}

func (l *GetRegistrationListLogic) GetRegistration(id string) (*types.RegistrationInfo, error) {
	var r types.RegistrationInfo
	var fullName, email, phone, status, remarks, createdAt, updatedAt string
	
	err := l.svcCtx.DB.QueryRow(`
		SELECT id, event_id, user_id, COALESCE(full_name, ''), COALESCE(email, ''), 
		       COALESCE(phone, ''), COALESCE(status, 'pending'), COALESCE(remarks, ''), 
		       created_at, updated_at 
		FROM registrations WHERE id = $1
	`, id).Scan(&r.ID, &r.EventID, &r.UserID, &fullName, &email, &phone, &status, &remarks, &createdAt, &updatedAt)
	
	if err != nil {
		return nil, errors.New("报名记录不存在")
	}
	
	r.FullName = fullName
	r.Email = email
	r.Phone = phone
	r.Status = status
	r.Remarks = remarks
	r.CreatedAt = createdAt
	r.UpdatedAt = updatedAt
	
	return &r, nil
}

type CreateRegistrationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateRegistrationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRegistrationLogic {
	return &CreateRegistrationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateRegistrationLogic) CreateRegistration(req *types.CreateRegistrationReq) (*types.RegistrationInfo, error) {
	if req.EventID == "" {
		return nil, errors.New("活动ID不能为空")
	}
	if req.FullName == "" {
		return nil, errors.New("姓名不能为空")
	}
	if req.Phone == "" {
		return nil, errors.New("手机号不能为空")
	}

	// 检查活动是否存在且在报名中
	var activityStatus string
	err := l.svcCtx.DB.QueryRow("SELECT COALESCE(status, '') FROM activities WHERE id = $1", req.EventID).Scan(&activityStatus)
	if err != nil || activityStatus != "active" {
		return nil, errors.New("活动不在报名中")
	}

	id := uuid.New().String()
	_, err = l.svcCtx.DB.Exec(`
		INSERT INTO registrations (id, event_id, user_id, full_name, email, phone, status, remarks, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, id, req.EventID, req.UserID, req.FullName, req.Email, req.Phone)
	if err != nil {
		return nil, errors.New("创建报名失败: " + err.Error())
	}

	return &types.RegistrationInfo{
		ID:        id,
		EventID:   req.EventID,
		UserID:    req.UserID,
		FullName:  req.FullName,
		Email:     req.Email,
		Phone:     req.Phone,
		Status:    "pending",
		CreatedAt: time.Now().Format("2006-01-02T15:04:05Z"),
	}, nil
}

// CreateRegistrationV2 支持动态表单字段
func (l *CreateRegistrationLogic) CreateRegistrationV2(req *types.CreateRegistrationReqV2) (*types.RegistrationInfo, error) {
	if req.ActivityID == "" {
		return nil, errors.New("活动ID不能为空")
	}
	if req.Phone == "" {
		return nil, errors.New("手机号不能为空")
	}

	// 检查活动是否存在且在报名中
	var activityStatus string
	err := l.svcCtx.DB.QueryRow("SELECT COALESCE(status, '') FROM activities WHERE id = $1", req.ActivityID).Scan(&activityStatus)
	if err != nil || activityStatus != "active" {
		return nil, errors.New("活动不在报名中")
	}

	id := uuid.New().String()
	
	// 构建 form_data JSON
	formData := make(map[string]interface{})
	formData["phone"] = req.Phone
	if req.FormData != "" {
		// 解析额外的表单数据
		var extraData map[string]interface{}
		if err := json.Unmarshal([]byte(req.FormData), &extraData); err == nil {
			for k, v := range extraData {
				formData[k] = v
			}
		}
	}

	formDataJSON, _ := json.Marshal(formData)

	_, err = l.svcCtx.DB.Exec(`
		INSERT INTO registrations (id, activity_id, sales_id, visitor_openid, form_data, status, created_at, updated_at)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), $5, 'pending', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, id, req.ActivityID, req.SalesID, req.VisitorOpenID, formDataJSON)
	if err != nil {
		return nil, errors.New("创建报名失败: " + err.Error())
	}

	return &types.RegistrationInfo{
		ID:        id,
		EventID:   req.ActivityID,
		Phone:     req.Phone,
		Status:    "pending",
		CreatedAt: time.Now().Format("2006-01-02T15:04:05Z"),
	}, nil
}

type UpdateRegistrationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateRegistrationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRegistrationLogic {
	return &UpdateRegistrationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateRegistrationLogic) UpdateRegistration(id string, req *types.UpdateRegistrationReq) (*types.RegistrationInfo, error) {
	_, err := l.svcCtx.DB.Exec(`
		UPDATE registrations 
		SET status = COALESCE(NULLIF($1, ''), status),
		    remarks = COALESCE(NULLIF($2, ''), remarks),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`, req.Status, req.Remarks, id)
	if err != nil {
		return nil, errors.New("更新报名失败: " + err.Error())
	}

	var r types.RegistrationInfo
	var fullName, email, phone, status, remarks, createdAt, updatedAt string
	l.svcCtx.DB.QueryRow(`SELECT id, event_id, user_id, COALESCE(full_name, ''), COALESCE(email, ''), COALESCE(phone, ''), COALESCE(status, 'pending'), COALESCE(remarks, ''), created_at, updated_at FROM registrations WHERE id = $1`, id).Scan(&r.ID, &r.EventID, &r.UserID, &fullName, &email, &phone, &status, &remarks, &createdAt, &updatedAt)
	
	r.FullName = fullName
	r.Email = email
	r.Phone = phone
	r.Status = status
	r.Remarks = remarks
	r.CreatedAt = createdAt
	r.UpdatedAt = updatedAt

	return &r, nil
}

// UpdateRegistrationWithAuth 更新报名（带权限检查，当前保留供将来扩展）
func (l *UpdateRegistrationLogic) UpdateRegistrationWithAuth(id string, req *types.UpdateRegistrationReq, userId string, role string) (*types.RegistrationInfo, error) {
	return l.UpdateRegistration(id, req)
}

func (l *UpdateRegistrationLogic) DeleteRegistration(id string, userId string, role string) error {
	result, err := l.svcCtx.DB.Exec(`DELETE FROM registrations WHERE id = $1`, id)
	if err != nil {
		return errors.New("删除报名失败: " + err.Error())
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("报名记录不存在")
	}
	
	return nil
}

// ==================== 游客查询报名记录 ====================

type GetVisitorRegistrationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetVisitorRegistrationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVisitorRegistrationsLogic {
	return &GetVisitorRegistrationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetVisitorRegistrationsLogic) GetVisitorRegistrations(phone string) (map[string]interface{}, error) {
	// 查询该手机号的所有报名记录
	rows, err := l.svcCtx.DB.Query(`
		SELECT r.id, r.activity_id, r.status, r.form_data, r.created_at, r.updated_at,
		       a.name as activity_name, a.start_date, a.end_date
		FROM registrations r
		LEFT JOIN activities a ON r.activity_id = a.id
		WHERE r.deleted_at IS NULL
		AND r.form_data::text LIKE $1
		ORDER BY r.created_at DESC
	`, "%"+phone+"%")
	if err != nil {
		return nil, errors.New("查询报名记录失败")
	}
	defer rows.Close()

	var registrations []map[string]interface{}
	for rows.Next() {
		var id, activityID, status, formData, createdAt, updatedAt, activityName, startDate, endDate sql.NullString
		if err := rows.Scan(&id, &activityID, &status, &formData, &createdAt, &updatedAt, &activityName, &startDate, &endDate); err == nil {
			reg := map[string]interface{}{
				"id":         id.String,
				"activityId": activityID.String,
				"status":     status.String,
				"createdAt":  createdAt.String,
				"updatedAt":  updatedAt.String,
			}
			if activityName.Valid {
				reg["activityName"] = activityName.String
			}
			if startDate.Valid {
				reg["startDate"] = startDate.String
			}
			if endDate.Valid {
				reg["endDate"] = endDate.String
			}
			if formData.Valid && formData.String != "" {
				var formDataMap map[string]interface{}
				if err := json.Unmarshal([]byte(formData.String), &formDataMap); err == nil {
					reg["formData"] = formDataMap
					// 提取手机号作为标识
					if phone, ok := formDataMap["phone"]; ok {
						reg["phone"] = phone
					}
				}
			}
			registrations = append(registrations, reg)
		}
	}

	return map[string]interface{}{
		"registrations": registrations,
	}, nil
}

// ==================== 游客修改报名 ====================

type UpdateVisitorRegistrationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateVisitorRegistrationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateVisitorRegistrationLogic {
	return &UpdateVisitorRegistrationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateVisitorRegistrationLogic) UpdateVisitorRegistration(id string, req *types.UpdateVisitorRegistrationReq) (*types.RegistrationInfo, error) {
	// 检查报名是否存在
	var existingPhone string
	var status string
	err := l.svcCtx.DB.QueryRow("SELECT COALESCE(form_data::text, ''), status FROM registrations WHERE id = $1 AND deleted_at IS NULL", id).Scan(&existingPhone, &status)
	if err != nil {
		return nil, errors.New("报名记录不存在")
	}

	// 只有待审核状态可以修改
	if status != "pending" {
		return nil, errors.New("当前状态不允许修改")
	}

	// 构建更新数据
	var formData map[string]interface{}
	if req.FormData != "" {
		json.Unmarshal([]byte(req.FormData), &formData)
	} else {
		formData = make(map[string]interface{})
	}

	// 更新手机号
	if req.Phone != "" {
		formData["phone"] = req.Phone
	}

	formDataJSON, _ := json.Marshal(formData)

	// 更新报名
	remark := req.Remark
	_, err = l.svcCtx.DB.Exec(`
		UPDATE registrations 
		SET form_data = $1, remarks = COALESCE($2, remarks), updated_at = CURRENT_TIMESTAMP
		WHERE id = $3 AND deleted_at IS NULL
	`, formDataJSON, remark, id)
	if err != nil {
		return nil, errors.New("更新报名失败: " + err.Error())
	}

	return &types.RegistrationInfo{
		ID:     id,
		Status: status,
	}, nil
}
