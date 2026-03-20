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

// ==================== Branch Logic ====================

type GetBranchListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBranchListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBranchListLogic {
	return &GetBranchListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type GetBranchListReq struct {
	CompanyID string `json:"companyId,optional"`
	Page      int    `json:"page,default=1"`
	PageSize  int    `json:"pageSize,default=20"`
	Keyword   string `json:"keyword,optional"`
}

// BranchWithCompany includes branch info plus the company name
type BranchWithCompany struct {
	ID          string `json:"id"`
	CompanyID   string `json:"companyId"`
	Name        string `json:"name"`
	LeaderName  string `json:"leaderName"`
	LeaderPhone string `json:"leaderPhone"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	CompanyName string `json:"companyName"`
}

func (l *GetBranchListLogic) GetBranchList(req *GetBranchListReq) (map[string]interface{}, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	// Build query with filters
	whereClause := "WHERE b.deleted_at IS NULL"
	args := []interface{}{}
	argNum := 1

	if req.CompanyID != "" {
		whereClause += fmt.Sprintf(" AND b.company_id = $%d", argNum)
		args = append(args, req.CompanyID)
		argNum++
	}

	if req.Keyword != "" {
		whereClause += fmt.Sprintf(" AND b.name LIKE $%d", argNum)
		args = append(args, "%"+req.Keyword+"%")
		argNum++
	}

	// Query total count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM branches b %s", whereClause)
	l.svcCtx.DB.QueryRow(countQuery, args...).Scan(&total)

	// Query list
	query := fmt.Sprintf(`
		SELECT b.id, b.company_id, b.name, b.leader_name, b.leader_phone, b.status, b.created_at, b.updated_at,
		       c.name as company_name
		FROM branches b
		LEFT JOIN companies c ON b.company_id = c.id
		%s
		ORDER BY b.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)
	args = append(args, req.PageSize, offset)

	rows, err := l.svcCtx.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []BranchWithCompany
	for rows.Next() {
		var branch BranchWithCompany
		var leaderName, leaderPhone, companyName sql.NullString
		if err := rows.Scan(&branch.ID, &branch.CompanyID, &branch.Name, &leaderName, &leaderPhone, &branch.Status, &branch.CreatedAt, &branch.UpdatedAt, &companyName); err == nil {
			branch.LeaderName = leaderName.String
			branch.LeaderPhone = leaderPhone.String
			branch.CompanyName = companyName.String
			branches = append(branches, branch)
		}
	}

	return map[string]interface{}{
		"branches": branches,
		"pagination": types.PaginationResp{
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalCount: total,
			TotalPages: (total + req.PageSize - 1) / req.PageSize,
		},
	}, nil
}

type GetBranchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBranchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBranchLogic {
	return &GetBranchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBranchLogic) GetBranch(id string) (*types.BranchInfo, error) {
	var branch types.BranchInfo
	var leaderName, leaderPhone sql.NullString
	err := l.svcCtx.DB.QueryRow(`
		SELECT id, company_id, name, leader_name, leader_phone, status, created_at, updated_at
		FROM branches WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(&branch.ID, &branch.CompanyID, &branch.Name, &leaderName, &leaderPhone, &branch.Status, &branch.CreatedAt, &branch.UpdatedAt)
	if err != nil {
		return nil, errors.New("分支机构不存在")
	}
	branch.LeaderName = leaderName.String
	branch.LeaderPhone = leaderPhone.String
	return &branch, nil
}

type CreateBranchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateBranchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateBranchLogic {
	return &CreateBranchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateBranchLogic) CreateBranch(req *types.CreateBranchReq) (*types.BranchInfo, error) {
	if req.CompanyID == "" {
		return nil, errors.New("所属公司不能为空")
	}
	if req.Name == "" {
		return nil, errors.New("分支机构名称不能为空")
	}

	// Check if company exists
	var companyExists bool
	l.svcCtx.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM companies WHERE id = $1 AND deleted_at IS NULL)", req.CompanyID).Scan(&companyExists)
	if !companyExists {
		return nil, errors.New("公司不存在")
	}

	id := uuid.New().String()
	if req.Status == "" {
		req.Status = "active"
	}

	_, err := l.svcCtx.DB.Exec(`
		INSERT INTO branches (id, company_id, name, leader_name, leader_phone, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, id, req.CompanyID, req.Name, req.LeaderName, req.LeaderPhone, req.Status)
	if err != nil {
		return nil, errors.New("创建分支机构失败: " + err.Error())
	}

	return &types.BranchInfo{
		ID:          id,
		CompanyID:   req.CompanyID,
		Name:        req.Name,
		LeaderName:  req.LeaderName,
		LeaderPhone: req.LeaderPhone,
		Status:      req.Status,
		CreatedAt:   time.Now().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   time.Now().Format("2006-01-02T15:04:05Z"),
	}, nil
}

type UpdateBranchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateBranchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBranchLogic {
	return &UpdateBranchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateBranchLogic) UpdateBranch(id string, req *types.UpdateBranchReq) (*types.BranchInfo, error) {
	setClauses := []string{}
	args := []interface{}{}
	argNum := 1

	if req.Name != "" {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argNum))
		args = append(args, req.Name)
		argNum++
	}
	if req.LeaderName != "" {
		setClauses = append(setClauses, fmt.Sprintf("leader_name = $%d", argNum))
		args = append(args, req.LeaderName)
		argNum++
	}
	if req.LeaderPhone != "" {
		setClauses = append(setClauses, fmt.Sprintf("leader_phone = $%d", argNum))
		args = append(args, req.LeaderPhone)
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

	query := fmt.Sprintf("UPDATE branches SET %s WHERE id = $%d AND deleted_at IS NULL", joinStrings(setClauses, ", "), argNum)

	result, err := l.svcCtx.DB.Exec(query, args...)
	if err != nil {
		return nil, errors.New("更新分支机构失败: " + err.Error())
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, errors.New("分支机构不存在")
	}

	return NewGetBranchLogic(l.ctx, l.svcCtx).GetBranch(id)
}

type DeleteBranchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteBranchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteBranchLogic {
	return &DeleteBranchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteBranchLogic) DeleteBranch(id string) error {
	// Check if branch has users
	var count int
	l.svcCtx.DB.QueryRow("SELECT COUNT(*) FROM users WHERE branch_id = $1 AND deleted_at IS NULL", id).Scan(&count)
	if count > 0 {
		return errors.New("该分支机构下有用户，无法删除")
	}

	// Check if branch has activities
	l.svcCtx.DB.QueryRow("SELECT COUNT(*) FROM activities WHERE branch_id = $1 AND deleted_at IS NULL", id).Scan(&count)
	if count > 0 {
		return errors.New("该分支机构下有活动，无法删除")
	}

	result, err := l.svcCtx.DB.Exec("UPDATE branches SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1", id)
	if err != nil {
		return errors.New("删除分支机构失败")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("分支机构不存在")
	}

	return nil
}
