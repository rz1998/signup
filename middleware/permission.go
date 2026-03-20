package middleware

import (
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"signup/config"
)



// CurrentUser holds the authenticated user info extracted from JWT
type CurrentUser struct {
	ID        string
	Username  string
	Role      string
	CompanyID string
	BranchID  string
}

// Role constants
const (
	RoleAdmin      = "admin"
	RoleSysMgr     = "sys_mgr"
	RoleCompanyMgr = "company_mgr"
	RoleBranchMgr  = "branch_mgr"
	RoleMgr        = "mgr"
	RoleSales      = "sales"
)

// ErrUnauthorized is returned when the request is not authenticated
var ErrUnauthorized = errors.New("未登录或token无效")

// ErrForbidden is returned when the user doesn't have permission
var ErrForbidden = errors.New("权限不足")

// ExtractUserFromToken extracts user info from a JWT token string
func ExtractUserFromToken(tokenString string, secret string) (*CurrentUser, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrUnauthorized
	}

	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return nil, ErrUnauthorized
	}

	role, _ := claims["role"].(string)
	if role == "" {
		role = RoleSales
	}

	username, _ := claims["username"].(string)
	companyID, _ := claims["company_id"].(string)
	branchID, _ := claims["branch_id"].(string)

	return &CurrentUser{
		ID:        userID,
		Username:  username,
		Role:      role,
		CompanyID: companyID,
		BranchID:  branchID,
	}, nil
}

// ExtractBearerToken extracts the token from Authorization header
func ExtractBearerToken(header string) string {
	header = strings.TrimSpace(header)
	if len(header) > 7 && strings.ToLower(header[:7]) == "bearer " {
		return header[7:]
	}
	return ""
}

// HasPermission checks if the current user has the required role
func (u *CurrentUser) HasPermission(allowedRoles ...string) bool {
	for _, role := range allowedRoles {
		if u.Role == role {
			return true
		}
	}
	return false
}

// IsAdmin returns true if user is admin (system admin)
func (u *CurrentUser) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsCompanyAdmin returns true if user is company admin
func (u *CurrentUser) IsCompanyAdmin() bool {
	return u.Role == RoleAdmin || u.Role == RoleCompanyMgr
}

// IsBranchAdmin returns true if user is branch admin or above
func (u *CurrentUser) IsBranchAdmin() bool {
	return u.Role == RoleAdmin || u.Role == RoleCompanyMgr || u.Role == RoleBranchMgr
}

// IsMgr returns true if user is manager or above (company_mgr, branch_mgr, mgr, admin)
func (u *CurrentUser) IsMgr() bool {
	return u.Role == RoleAdmin || u.Role == RoleCompanyMgr || u.Role == RoleBranchMgr || u.Role == RoleMgr
}

// IsSales returns true if user is sales
func (u *CurrentUser) IsSales() bool {
	return u.Role == RoleSales
}

// IsSystemAdmin returns true if user is system admin only
func (u *CurrentUser) IsSystemAdmin() bool {
	return u.Role == RoleAdmin
}

// CanManageUsers returns true if user can manage users
func (u *CurrentUser) CanManageUsers() bool {
	return u.IsAdmin()
}

// CanCreateUsers returns true if user can create users
func (u *CurrentUser) CanCreateUsers() bool {
	return u.IsMgr()
}

// CanManageActivities returns true if user can create/edit/delete activities
func (u *CurrentUser) CanManageActivities() bool {
	return u.IsMgr()
}

// CanViewAllRegistrations returns true if user can see all registrations
func (u *CurrentUser) CanViewAllRegistrations() bool {
	return u.IsMgr()
}

// CanManageCompanies returns true if user can manage companies (system admin only)
func (u *CurrentUser) CanManageCompanies() bool {
	return u.IsSystemAdmin()
}

// CanManageBranches returns true if user can manage branches
func (u *CurrentUser) CanManageBranches() bool {
	return u.IsCompanyAdmin()
}

// GetJWTSecret returns the JWT secret from config
func GetJWTSecret(cfg *config.Config) string {
	return cfg.JWT.Secret
}

// CheckAuth is a helper to check auth from request headers
func CheckAuth(authHeader string, secret string) (*CurrentUser, error) {
	token := ExtractBearerToken(authHeader)
	if token == "" {
		return nil, ErrUnauthorized
	}
	return ExtractUserFromToken(token, secret)
}
