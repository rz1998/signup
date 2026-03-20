package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"signup/svc"
	"signup/types"
)

// ==================== Company Logic ====================

type GetCompanyListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCompanyListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCompanyListLogic {
	return &GetCompanyListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCompanyListLogic) GetCompanyList(req *types.PaginationReq) (map[string]interface{}, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	// Query total count
	var total int
	if req.Keyword != "" {
		l.svcCtx.DB.QueryRow(`
			SELECT COUNT(*) FROM companies 
			WHERE deleted_at IS NULL 
			AND name LIKE $1
		`, "%"+req.Keyword+"%").Scan(&total)
	} else {
		l.svcCtx.DB.QueryRow("SELECT COUNT(*) FROM companies WHERE deleted_at IS NULL").Scan(&total)
	}

	// Query list
	var rows *sql.Rows
	var err error
	if req.Keyword != "" {
		rows, err = l.svcCtx.DB.Query(`
			SELECT id, name, description, status, created_at, updated_at
			FROM companies 
			WHERE deleted_at IS NULL 
			AND name LIKE $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`, "%"+req.Keyword+"%", req.PageSize, offset)
	} else {
		rows, err = l.svcCtx.DB.Query(`
			SELECT id, name, description, status, created_at, updated_at
			FROM companies 
			WHERE deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`, req.PageSize, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []types.CompanyInfo
	for rows.Next() {
		var company types.CompanyInfo
		var description sql.NullString
		if err := rows.Scan(&company.ID, &company.Name, &description, &company.Status, &company.CreatedAt, &company.UpdatedAt); err == nil {
			company.Description = description.String
			companies = append(companies, company)
		}
	}

	return map[string]interface{}{
		"companies": companies,
		"pagination": types.PaginationResp{
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalCount: total,
			TotalPages: (total + req.PageSize - 1) / req.PageSize,
		},
	}, nil
}

type GetCompanyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCompanyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCompanyLogic {
	return &GetCompanyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCompanyLogic) GetCompany(id string) (*types.CompanyInfo, error) {
	var company types.CompanyInfo
	var description sql.NullString
	err := l.svcCtx.DB.QueryRow(`
		SELECT id, name, description, status, created_at, updated_at
		FROM companies WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(&company.ID, &company.Name, &description, &company.Status, &company.CreatedAt, &company.UpdatedAt)
	if err != nil {
		return nil, errors.New("公司不存在")
	}
	company.Description = description.String
	return &company, nil
}

type CreateCompanyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateCompanyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCompanyLogic {
	return &CreateCompanyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateCompanyLogic) CreateCompany(req *types.CreateCompanyReq) (*types.CompanyInfo, error) {
	if req.Name == "" {
		return nil, errors.New("公司名称不能为空")
	}

	id := uuid.New().String()
	if req.Status == "" {
		req.Status = "active"
	}

	_, err := l.svcCtx.DB.Exec(`
		INSERT INTO companies (id, name, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, id, req.Name, req.Description, req.Status)
	if err != nil {
		return nil, errors.New("创建公司失败: " + err.Error())
	}

	return &types.CompanyInfo{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		CreatedAt:   time.Now().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   time.Now().Format("2006-01-02T15:04:05Z"),
	}, nil
}

type UpdateCompanyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateCompanyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCompanyLogic {
	return &UpdateCompanyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateCompanyLogic) UpdateCompany(id string, req *types.UpdateCompanyReq) (*types.CompanyInfo, error) {
	setClauses := []string{}
	args := []interface{}{}
	argNum := 1

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

	query := fmt.Sprintf("UPDATE companies SET %s WHERE id = $%d AND deleted_at IS NULL", joinStrings(setClauses, ", "), argNum)

	result, err := l.svcCtx.DB.Exec(query, args...)
	if err != nil {
		return nil, errors.New("更新公司失败: " + err.Error())
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, errors.New("公司不存在")
	}

	return NewGetCompanyLogic(l.ctx, l.svcCtx).GetCompany(id)
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

type DeleteCompanyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteCompanyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCompanyLogic {
	return &DeleteCompanyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteCompanyLogic) DeleteCompany(id string) error {
	// Check if company has branches
	var count int
	l.svcCtx.DB.QueryRow("SELECT COUNT(*) FROM branches WHERE company_id = $1 AND deleted_at IS NULL", id).Scan(&count)
	if count > 0 {
		return errors.New("该公司下有分支机构，无法删除")
	}

	// Check if company has users
	l.svcCtx.DB.QueryRow("SELECT COUNT(*) FROM users WHERE company_id = $1 AND deleted_at IS NULL", id).Scan(&count)
	if count > 0 {
		return errors.New("该公司下有用户，无法删除")
	}

	result, err := l.svcCtx.DB.Exec("UPDATE companies SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1", id)
	if err != nil {
		return errors.New("删除公司失败")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("公司不存在")
	}

	return nil
}
