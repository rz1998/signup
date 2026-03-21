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
	"golang.org/x/crypto/bcrypt"
)

type GetUserListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserListLogic {
	return &GetUserListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type GetUserListReq struct {
	CompanyID string `json:"companyId,optional"`
	BranchID string `json:"branchId,optional"`
	Role     string `json:"role,optional"`
	Status   string `json:"status,optional"`
	Page     int    `json:"page,default=1"`
	PageSize int    `json:"pageSize,default=20"`
	Keyword  string `json:"keyword,optional"`
}

func (l *GetUserListLogic) GetUserList(req *GetUserListReq) (map[string]interface{}, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	// Build query with filters
	whereClause := "WHERE u.deleted_at IS NULL"
	args := []interface{}{}
	argNum := 1

	if req.CompanyID != "" {
		whereClause += fmt.Sprintf(" AND u.company_id = $%d", argNum)
		args = append(args, req.CompanyID)
		argNum++
	}

	if req.BranchID != "" {
		whereClause += fmt.Sprintf(" AND u.branch_id = $%d", argNum)
		args = append(args, req.BranchID)
		argNum++
	}

	if req.Role != "" {
		// 支持多角色查询，用逗号分隔，如 "company_admin,branch_admin"
		if strings.Contains(req.Role, ",") {
			roles := strings.Split(req.Role, ",")
			rolePlaceholders := []string{}
			for _, role := range roles {
				rolePlaceholders = append(rolePlaceholders, fmt.Sprintf("$%d", argNum))
				args = append(args, strings.TrimSpace(role))
				argNum++
			}
			whereClause += fmt.Sprintf(" AND u.role IN (%s)", strings.Join(rolePlaceholders, ", "))
		} else {
			whereClause += fmt.Sprintf(" AND u.role = $%d", argNum)
			args = append(args, req.Role)
			argNum++
		}
	}

	if req.Status != "" {
		whereClause += fmt.Sprintf(" AND u.status = $%d", argNum)
		args = append(args, req.Status)
		argNum++
	}

	if req.Keyword != "" {
		whereClause += fmt.Sprintf(" AND (u.username LIKE $%d OR u.full_name LIKE $%d OR u.email LIKE $%d OR u.phone LIKE $%d)", argNum, argNum, argNum, argNum)
		args = append(args, "%"+req.Keyword+"%")
		argNum++
	}

	// Query total count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users u %s", whereClause)
	l.svcCtx.DB.QueryRow(countQuery, args...).Scan(&total)

	// Query list
	query := fmt.Sprintf(`
		SELECT u.id, u.username, u.full_name, u.email, u.phone, u.role, u.company_id, u.branch_id, u.manager_id, u.status, u.created_at
		FROM users u
		%s
		ORDER BY u.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)
	args = append(args, req.PageSize, offset)

	rows, err := l.svcCtx.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []types.UserInfo
	for rows.Next() {
		var user types.UserInfo
		var fullName, email, phone, status, createdAt sql.NullString
		var companyID, branchID, managerID sql.NullString
		if err := rows.Scan(&user.ID, &user.Username, &fullName, &email, &phone, &user.Role, &companyID, &branchID, &managerID, &status, &createdAt); err == nil {
			user.FullName = fullName.String
			user.Email = email.String
			user.Phone = phone.String
			user.Status = status.String
			user.CreatedAt = createdAt.String
			user.CompanyID = companyID.String
			user.BranchID = branchID.String
			user.ManagerID = managerID.String
			users = append(users, user)
		}
	}

	return map[string]interface{}{
		"users": users,
		"pagination": types.PaginationResp{
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalCount: total,
			TotalPages: (total + req.PageSize - 1) / req.PageSize,
		},
	}, nil
}

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserLogic) GetUser(id string) (*types.UserInfo, error) {
	var user types.UserInfo
	var fullName, email, phone, status, createdAt sql.NullString
	var companyID, branchID, managerID sql.NullString
	err := l.svcCtx.DB.QueryRow(`
		SELECT id, username, full_name, email, phone, role, company_id, branch_id, manager_id, status, created_at
		FROM users WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(&user.ID, &user.Username, &fullName, &email, &phone, &user.Role, &companyID, &branchID, &managerID, &status, &createdAt)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	user.FullName = fullName.String
	user.Email = email.String
	user.Phone = phone.String
	user.Status = status.String
	user.CreatedAt = createdAt.String
	user.CompanyID = companyID.String
	user.BranchID = branchID.String
	user.ManagerID = managerID.String
	return &user, nil
}

type CreateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateUserLogic) CreateUser(req *types.CreateUserReq) (*types.UserInfo, error) {
	if req.Username == "" || req.Password == "" {
		return nil, errors.New("用户名和密码不能为空")
	}

	// 检查用户名是否已存在
	var count int
	l.svcCtx.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1 AND deleted_at IS NULL", req.Username).Scan(&count)
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	// 验证公司存在（如果提供了）
	if req.CompanyID != "" {
		var exists bool
		l.svcCtx.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM companies WHERE id = $1 AND deleted_at IS NULL)", req.CompanyID).Scan(&exists)
		if !exists {
			return nil, errors.New("公司不存在")
		}
	}

	// 验证分支机构存在（如果提供了）
	if req.BranchID != "" {
		var exists bool
		l.svcCtx.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM branches WHERE id = $1 AND deleted_at IS NULL)", req.BranchID).Scan(&exists)
		if !exists {
			return nil, errors.New("分支机构不存在")
		}
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	id := uuid.New().String()
	if req.Role == "" {
		req.Role = "sales"
	}
	if req.Status == "" {
		req.Status = "active"
	}

	_, err = l.svcCtx.DB.Exec(`
		INSERT INTO users (id, username, password, full_name, email, phone, role, company_id, branch_id, manager_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''), $11, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, id, req.Username, string(hashedPassword), req.FullName, req.Email, req.Phone, req.Role, req.CompanyID, req.BranchID, req.ManagerID, req.Status)
	if err != nil {
		return nil, errors.New("创建用户失败: " + err.Error())
	}

	return &types.UserInfo{
		ID:        id,
		Username:  req.Username,
		FullName:  req.FullName,
		Email:     req.Email,
		Phone:     req.Phone,
		Role:      req.Role,
		CompanyID: req.CompanyID,
		BranchID:  req.BranchID,
		ManagerID: req.ManagerID,
		Status:    req.Status,
		CreatedAt: time.Now().Format("2006-01-02T15:04:05Z"),
	}, nil
}

type UpdateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserLogic) UpdateUser(id string, req *types.UpdateUserReq) (*types.UserInfo, error) {
	setClauses := []string{}
	args := []interface{}{}
	argNum := 1

	if req.FullName != "" {
		setClauses = append(setClauses, fmt.Sprintf("full_name = $%d", argNum))
		args = append(args, req.FullName)
		argNum++
	}
	if req.Email != "" {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argNum))
		args = append(args, req.Email)
		argNum++
	}
	if req.Phone != "" {
		setClauses = append(setClauses, fmt.Sprintf("phone = $%d", argNum))
		args = append(args, req.Phone)
		argNum++
	}
	if req.Role != "" {
		setClauses = append(setClauses, fmt.Sprintf("role = $%d", argNum))
		args = append(args, req.Role)
		argNum++
	}
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
	if req.ManagerID != "" {
		setClauses = append(setClauses, fmt.Sprintf("manager_id = $%d", argNum))
		args = append(args, req.ManagerID)
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

	setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d AND deleted_at IS NULL", joinStrings(setClauses, ", "), argNum)

	result, err := l.svcCtx.DB.Exec(query, args...)
	if err != nil {
		return nil, errors.New("更新用户失败: " + err.Error())
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, errors.New("用户不存在")
	}

	return NewGetUserLogic(l.ctx, l.svcCtx).GetUser(id)
}

type DeleteUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserLogic {
	return &DeleteUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteUserLogic) DeleteUser(id string, currentUserId string) error {
	if id == currentUserId {
		return errors.New("不能删除自己")
	}
	result, err := l.svcCtx.DB.Exec("UPDATE users SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1", id)
	if err != nil {
		return errors.New("删除用户失败")
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("用户不存在")
	}
	return nil
}

type ResetPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetPasswordLogic {
	return &ResetPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetPasswordLogic) ResetPassword(id string, newPassword string) error {
	if newPassword == "" {
		return errors.New("新密码不能为空")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	result, err := l.svcCtx.DB.Exec("UPDATE users SET password = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND deleted_at IS NULL", string(hashedPassword), id)
	if err != nil {
		return errors.New("重置密码失败: " + err.Error())
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("用户不存在")
	}

	return nil
}

type GetUserStatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserStatsLogic {
	return &GetUserStatsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserStatsLogic) GetUserStats() (map[string]interface{}, error) {
	var totalUsers, activeUsers int

	l.svcCtx.DB.QueryRow("SELECT COUNT(*) FROM users WHERE deleted_at IS NULL").Scan(&totalUsers)
	l.svcCtx.DB.QueryRow("SELECT COUNT(*) FROM users WHERE status = 'active' AND deleted_at IS NULL").Scan(&activeUsers)

	return map[string]interface{}{
		"totalUsers":  totalUsers,
		"activeUsers": activeUsers,
	}, nil
}
