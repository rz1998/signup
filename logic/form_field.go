package logic

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"signup/svc"
	"signup/types"
)

// ==================== Form Field Logic ====================

type GetFormFieldsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFormFieldsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFormFieldsLogic {
	return &GetFormFieldsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFormFieldsLogic) GetFormFields(activityId string) ([]types.FormFieldInfo, error) {
	rows, err := l.svcCtx.DB.Query(`
		SELECT id, activity_id, field_name, field_type, is_required, options, sort_order, created_at
		FROM form_fields 
		WHERE activity_id = $1 AND deleted_at IS NULL
		ORDER BY sort_order ASC
	`, activityId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fields []types.FormFieldInfo
	for rows.Next() {
		var field types.FormFieldInfo
		var options sql.NullString
		if err := rows.Scan(&field.ID, &field.ActivityID, &field.FieldName, &field.FieldType, 
			&field.IsRequired, &options, &field.SortOrder, &field.CreatedAt); err == nil {
			if options.Valid {
				field.Options = options.String
			}
			fields = append(fields, field)
		}
	}
	return fields, nil
}

type CreateFormFieldLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateFormFieldLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateFormFieldLogic {
	return &CreateFormFieldLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateFormFieldLogic) CreateFormField(activityId string, req *types.CreateFormFieldReq) (*types.FormFieldInfo, error) {
	if req.FieldName == "" {
		return nil, errors.New("字段名称不能为空")
	}
	if req.FieldType == "" {
		return nil, errors.New("字段类型不能为空")
	}

	// Get max sort order
	var maxSort int
	l.svcCtx.DB.QueryRow("SELECT COALESCE(MAX(sort_order), 0) FROM form_fields WHERE activity_id = $1", activityId).Scan(&maxSort)

	id := uuid.New().String()
	_, err := l.svcCtx.DB.Exec(`
		INSERT INTO form_fields (id, activity_id, field_name, field_type, is_required, options, sort_order, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
	`, id, activityId, req.FieldName, req.FieldType, req.IsRequired, req.Options, maxSort+1)
	if err != nil {
		return nil, errors.New("创建字段失败: " + err.Error())
	}

	return &types.FormFieldInfo{
		ID:         id,
		ActivityID: activityId,
		FieldName:  req.FieldName,
		FieldType:  req.FieldType,
		IsRequired: req.IsRequired,
		Options:    req.Options,
		SortOrder:  maxSort + 1,
	}, nil
}

type UpdateFormFieldLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateFormFieldLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFormFieldLogic {
	return &UpdateFormFieldLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateFormFieldLogic) UpdateFormField(id string, req *types.UpdateFormFieldReq) (*types.FormFieldInfo, error) {
	// Handle nil pointers
	var sortOrder int
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	result, err := l.svcCtx.DB.Exec(`
		UPDATE form_fields 
		SET field_name = COALESCE(NULLIF($1, ''), field_name),
		    field_type = COALESCE(NULLIF($2, ''), field_type),
		    is_required = COALESCE($3, is_required),
		    options = COALESCE(NULLIF($4, ''), options),
		    sort_order = COALESCE($5, sort_order)
		WHERE id = $6 AND deleted_at IS NULL
	`, req.FieldName, req.FieldType, req.IsRequired, req.Options, sortOrder, id)
	if err != nil {
		return nil, errors.New("更新字段失败")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, errors.New("字段不存在")
	}

	// Get updated field info
	var field types.FormFieldInfo
	var options sql.NullString
	err = l.svcCtx.DB.QueryRow(`
		SELECT id, activity_id, field_name, field_type, is_required, options, sort_order, created_at
		FROM form_fields WHERE id = $1
	`, id).Scan(&field.ID, &field.ActivityID, &field.FieldName, &field.FieldType, &field.IsRequired, &options, &field.SortOrder, &field.CreatedAt)
	if err != nil {
		return nil, errors.New("获取字段信息失败")
	}
	if options.Valid {
		field.Options = options.String
	}

	return &field, nil
}

type DeleteFormFieldLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteFormFieldLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFormFieldLogic {
	return &DeleteFormFieldLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteFormFieldLogic) DeleteFormField(id string) error {
	result, err := l.svcCtx.DB.Exec("UPDATE form_fields SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1", id)
	if err != nil {
		return errors.New("删除字段失败")
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("字段不存在")
	}
	return nil
}
