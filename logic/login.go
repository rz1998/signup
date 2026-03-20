package logic

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"signup/svc"
	"signup/types"
	"golang.org/x/crypto/bcrypt"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (*types.LoginResp, error) {
	if req.Username == "" || req.Password == "" {
		return nil, errors.New("用户名和密码不能为空")
	}

	var user types.UserInfo
	var storedPassword string
	var fullName, email, phone, status, createdAt, companyID, branchID *string

	err := l.svcCtx.DB.QueryRow(`
		SELECT id, username, full_name, email, phone, role, status, created_at, password, company_id, branch_id
		FROM users WHERE username = $1 AND deleted_at IS NULL
	`, req.Username).Scan(
		&user.ID, &user.Username, &fullName, &email, &phone,
		&user.Role, &status, &createdAt, &storedPassword, &companyID, &branchID,
	)

	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	if fullName != nil {
		user.FullName = *fullName
	}
	if email != nil {
		user.Email = *email
	}
	if phone != nil {
		user.Phone = *phone
	}
	if status != nil {
		user.Status = *status
	}
	if createdAt != nil {
		user.CreatedAt = *createdAt
	}
	if companyID != nil {
		user.CompanyID = *companyID
	}
	if branchID != nil {
		user.BranchID = *branchID
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(req.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	if user.Status != "active" {
		return nil, errors.New("用户已被禁用")
	}

	// 生成 JWT Token
	claims := jwt.MapClaims{
		"user_id":    user.ID,
		"username":   user.Username,
		"role":       user.Role,
		"company_id": user.CompanyID,
		"branch_id":  user.BranchID,
		"exp":        time.Now().Add(time.Hour * time.Duration(l.svcCtx.Config.JWT.Expire)).Unix(),
		"iat":        time.Now().Unix(),
		"issuer":     l.svcCtx.Config.JWT.Issuer,
		"audience":   l.svcCtx.Config.JWT.Audience,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(l.svcCtx.Config.JWT.Secret))
	if err != nil {
		return nil, errors.New("Token生成失败")
	}

	return &types.LoginResp{
		Token: tokenString,
		User:  &user,
	}, nil
}
