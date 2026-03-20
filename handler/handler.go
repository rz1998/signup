package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"signup/logic"
	"signup/middleware"
	"signup/svc"
	"signup/types"
)

func wrapResponse(data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"success": true,
		"data":    data,
	}
}

func getPathParam(r *http.Request, key string) string {
	path := r.URL.Path
	parts := strings.Split(path, "/")

	// Try to find by matching against route pattern (e.g., :id, :shareId)
	for i, part := range parts {
		if part == ":"+key && i+1 < len(parts) {
			return parts[i+1]
		}
	}

	// For shareId: /api/v1/visitor/activity/:shareId
	// The actual path looks like /api/v1/visitor/activity/<shareId>
	if key == "shareId" {
		for i := len(parts) - 1; i >= 0; i-- {
			part := parts[i]
			if part != "" && part[0] != ':' {
				return part
			}
		}
		return ""
	}

	// For id: handle all RESTful resources
	// go-zero replaces :id with actual value, so we find by resource name
	if key == "id" {
		// Resource names that appear before the ID in paths
		resourceNames := []string{"companies", "branches", "users", "activities", "fields", "registrations", "admin"}

		for _, resource := range resourceNames {
			for i, part := range parts {
				if part == resource && i+1 < len(parts) {
					return parts[i+1]
				}
			}
		}
	}

	return ""
}

// getCurrentUser 从请求中获取当前用户信息（包含角色）
func getCurrentUser(w http.ResponseWriter, r *http.Request, svcCtx *svc.ServiceContext) (*middleware.CurrentUser, error) {
	authHeader := r.Header.Get("Authorization")
	user, err := middleware.CheckAuth(authHeader, svcCtx.Config.JWT.Secret)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// getCurrentUserID 从请求中获取用户ID（向后兼容）
func getCurrentUserID(w http.ResponseWriter, r *http.Request, svcCtx *svc.ServiceContext) (string, error) {
	user, err := getCurrentUser(w, r, svcCtx)
	if err != nil {
		return "", err
	}
	return user.ID, nil
}

// requireRole 检查用户角色，返回错误或继续处理
func requireRole(w http.ResponseWriter, r *http.Request, svcCtx *svc.ServiceContext, allowedRoles ...string) (*middleware.CurrentUser, bool) {
	user, err := getCurrentUser(w, r, svcCtx)
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return nil, false
	}
	if !user.HasPermission(allowedRoles...) {
		httpx.ErrorCtx(r.Context(), w, errors.New("权限不足"))
		return nil, false
	}
	return user, true
}

// writeForbidden writes a forbidden response
func writeForbidden(w http.ResponseWriter, ctx context.Context) {
	httpx.ErrorCtx(ctx, w, errors.New("权限不足"))
}

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			// 健康检查
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/health",
				Handler: healthHandler(serverCtx),
			},
			// API 文档
			{
				Method:  http.MethodGet,
				Path:    "/api/docs",
				Handler: apiDocHandler(serverCtx),
			},
			// ==================== 认证 API ====================
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/auth/login",
				Handler: loginHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/auth/info",
				Handler: currentUserHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/auth/logout",
				Handler: logoutHandler(serverCtx),
			},
			// ==================== 公司管理 API ====================
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/companies",
				Handler: getCompanyListHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/companies",
				Handler: createCompanyHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/companies/:id",
				Handler: getCompanyHandler(serverCtx),
			},
			{
				Method:  http.MethodPut,
				Path:    "/api/v1/companies/:id",
				Handler: updateCompanyHandler(serverCtx),
			},
			{
				Method:  http.MethodDelete,
				Path:    "/api/v1/companies/:id",
				Handler: deleteCompanyHandler(serverCtx),
			},
			// ==================== 分支机构管理 API ====================
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/branches",
				Handler: getBranchListHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/branches",
				Handler: createBranchHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/branches/:id",
				Handler: getBranchHandler(serverCtx),
			},
			{
				Method:  http.MethodPut,
				Path:    "/api/v1/branches/:id",
				Handler: updateBranchHandler(serverCtx),
			},
			{
				Method:  http.MethodDelete,
				Path:    "/api/v1/branches/:id",
				Handler: deleteBranchHandler(serverCtx),
			},
			// ==================== 用户管理 API ====================
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/users",
				Handler: getUserListHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/users/me",
				Handler: currentUserHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/users/:id",
				Handler: getUserHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/users",
				Handler: createUserHandler(serverCtx),
			},
			{
				Method:  http.MethodPut,
				Path:    "/api/v1/users/:id",
				Handler: updateUserHandler(serverCtx),
			},
			{
				Method:  http.MethodDelete,
				Path:    "/api/v1/users/:id",
				Handler: deleteUserHandler(serverCtx),
			},
			{
				Method:  http.MethodPut,
				Path:    "/api/v1/users/:id/reset-password",
				Handler: resetPasswordHandler(serverCtx),
			},
			// ==================== 活动管理 API ====================
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/activities",
				Handler: getActivityListHandler(serverCtx),
			},
			// ==================== 表单字段 API ====================
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/activities/:id/fields",
				Handler: getFormFieldsHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/activities/:id/fields",
				Handler: createFormFieldHandler(serverCtx),
			},
			{
				Method:  http.MethodPut,
				Path:    "/api/v1/activities/:id/status",
				Handler: updateActivityStatusHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/activities/:id/stats",
				Handler: getActivityStatsHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/activities/:id",
				Handler: getActivityHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/activities",
				Handler: createActivityHandler(serverCtx),
			},
			{
				Method:  http.MethodPut,
				Path:    "/api/v1/activities/:id",
				Handler: updateActivityHandler(serverCtx),
			},
			{
				Method:  http.MethodDelete,
				Path:    "/api/v1/activities/:id",
				Handler: deleteActivityHandler(serverCtx),
			},
			{
				Method:  http.MethodPut,
				Path:    "/api/v1/fields/:id",
				Handler: updateFormFieldHandler(serverCtx),
			},
			{
				Method:  http.MethodDelete,
				Path:    "/api/v1/fields/:id",
				Handler: deleteFormFieldHandler(serverCtx),
			},
			// ==================== 报名 API ====================
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/activities/:id/registrations",
				Handler: getRegistrationsByActivityHandler(serverCtx),
			},
			// 游客查询自己的报名（通过手机号）
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/visitor/registrations",
				Handler: getVisitorRegistrationsHandler(serverCtx),
			},
			// 游客修改报名
			{
				Method:  http.MethodPut,
				Path:    "/api/v1/visitor/registrations/:id",
				Handler: updateVisitorRegistrationHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/registrations",
				Handler: getRegistrationListHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/registrations/:id",
				Handler: getRegistrationHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/registrations",
				Handler: createRegistrationHandler(serverCtx),
			},
			{
				Method:  http.MethodPut,
				Path:    "/api/v1/registrations/:id",
				Handler: updateRegistrationHandler(serverCtx),
			},
			// ==================== 管理员报名 API ====================
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/admin/registrations",
				Handler: getAdminRegistrationListHandler(serverCtx),
			},
			{
				Method:  http.MethodPut,
				Path:    "/api/v1/admin/registrations/:id/review",
				Handler: adminReviewRegistrationHandler(serverCtx),
			},
			{
				Method:  http.MethodDelete,
				Path:    "/api/v1/admin/registrations/:id",
				Handler: adminDeleteRegistrationHandler(serverCtx),
			},
			// ==================== 管理员设置 API ====================
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/admin/settings",
				Handler: getSettingsHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/admin/settings",
				Handler: saveSettingsHandler(serverCtx),
			},
			// ==================== 文件上传 API ====================
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/upload/image",
				Handler: uploadImageHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/upload/file",
				Handler: uploadFileHandler(serverCtx),
			},
			// ==================== 分享 API ====================
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/share/generate",
				Handler: generateShareHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/visitor/activity/:shareId",
				Handler: getVisitorActivityHandler(serverCtx),
			},
		},
	)
}

// ==================== 健康检查 ====================

func healthHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewHealthLogic(r.Context(), svcCtx)
		resp, err := l.Health()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

// ==================== API 文档 ====================

func apiDocHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>报名系统 - API文档</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f7fa; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; }
        h1 { color: #1e3a8a; }
        .endpoint { background: #f8fafc; padding: 15px; margin: 10px 0; border-radius: 4px; border-left: 4px solid #1e3a8a; }
        .method { display: inline-block; padding: 4px 8px; border-radius: 4px; font-weight: bold; margin-right: 10px; }
        .get { background: #10b981; color: white; }
        .post { background: #3b82f6; color: white; }
        .put { background: #f59e0b; color: white; }
        .delete { background: #ef4444; color: white; }
        .path { font-family: monospace; font-size: 14px; color: #1e3a8a; }
    </style>
</head>
<body>
    <div class="container">
        <h1>📚 报名系统 API 文档</h1>
        <p>Base URL: <code>/api/v1</code> | 认证方式: Bearer Token</p>
        
        <h2>认证</h2>
        <div class="endpoint">
            <span class="method post">POST</span>
            <span class="path">/api/v1/auth/login</span>
            <div>用户登录</div>
        </div>
        <div class="endpoint">
            <span class="method get">GET</span>
            <span class="path">/api/v1/auth/info</span>
            <div>获取当前用户</div>
        </div>
        
        <h2>用户管理</h2>
        <div class="endpoint">
            <span class="method get">GET</span>
            <span class="path">/api/v1/users</span>
            <div>用户列表</div>
        </div>
        <div class="endpoint">
            <span class="method post">POST</span>
            <span class="path">/api/v1/users</span>
            <div>创建用户</div>
        </div>
        <div class="endpoint">
            <span class="method put">PUT</span>
            <span class="path">/api/v1/users/:id</span>
            <div>更新用户</div>
        </div>
        <div class="endpoint">
            <span class="method delete">DELETE</span>
            <span class="path">/api/v1/users/:id</span>
            <div>删除用户</div>
        </div>
        
        <h2>活动管理</h2>
        <div class="endpoint">
            <span class="method get">GET</span>
            <span class="path">/api/v1/activities</span>
            <div>活动列表</div>
        </div>
        <div class="endpoint">
            <span class="method post">POST</span>
            <span class="path">/api/v1/activities</span>
            <div>创建活动</div>
        </div>
        <div class="endpoint">
            <span class="method put">PUT</span>
            <span class="path">/api/v1/activities/:id</span>
            <div>更新活动</div>
        </div>
        <div class="endpoint">
            <span class="method delete">DELETE</span>
            <span class="path">/api/v1/activities/:id</span>
            <div>删除活动</div>
        </div>
        <div class="endpoint">
            <span class="method put">PUT</span>
            <span class="path">/api/v1/activities/:id/status</span>
            <div>更新活动状态</div>
        </div>
        
        <h2>表单字段</h2>
        <div class="endpoint">
            <span class="method get">GET</span>
            <span class="path">/api/v1/activities/:id/fields</span>
            <div>获取表单字段</div>
        </div>
        <div class="endpoint">
            <span class="method post">POST</span>
            <span class="path">/api/v1/activities/:id/fields</span>
            <div>创建字段</div>
        </div>
        <div class="endpoint">
            <span class="method put">PUT</span>
            <span class="path">/api/v1/fields/:id</span>
            <div>更新字段</div>
        </div>
        <div class="endpoint">
            <span class="method delete">DELETE</span>
            <span class="path">/api/v1/fields/:id</span>
            <div>删除字段</div>
        </div>
        
        <h2>报名管理</h2>
        <div class="endpoint">
            <span class="method get">GET</span>
            <span class="path">/api/v1/activities/:id/registrations</span>
            <div>报名列表</div>
        </div>
        <div class="endpoint">
            <span class="method post">POST</span>
            <span class="path">/api/v1/registrations</span>
            <div>创建报名</div>
        </div>
        <div class="endpoint">
            <span class="method put">PUT</span>
            <span class="path">/api/v1/registrations/:id</span>
            <div>更新报名</div>
        </div>
        
        <h2>分享</h2>
        <div class="endpoint">
            <span class="method post">POST</span>
            <span class="path">/api/v1/share/generate</span>
            <div>生成分享链接</div>
        </div>
        <div class="endpoint">
            <span class="method get">GET</span>
            <span class="path">/api/v1/visitor/activity/:shareId</span>
            <div>游客访问活动</div>
        </div>
    </div>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	}
}

// ==================== 认证接口 ====================

func loginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewLoginLogic(r.Context(), svcCtx)
		resp, err := l.Login(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func currentUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, err := getCurrentUserID(w, r, svcCtx)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetUserLogic(r.Context(), svcCtx)
		user, err := l.GetUser(userId)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, fmt.Errorf("获取用户信息失败: %v", err))
			return
		}

		httpx.OkJsonCtx(r.Context(), w, wrapResponse(user))
	}
}

func logoutHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{"success": true, "message": "logout successful"})
	}
}

// ==================== 用户管理 ====================

func getUserListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req logic.GetUserListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewGetUserListLogic(r.Context(), svcCtx)
		resp, err := l.GetUserList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func getUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetUserLogic(r.Context(), svcCtx)
		resp, err := l.GetUser(getPathParam(r, "id"))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func createUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin, middleware.RoleCompanyMgr, middleware.RoleBranchMgr, middleware.RoleMgr)

		if !ok {
			return
		}
		_ = user
		var req types.CreateUserReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 只有admin可以创建admin/sys_mgr角色用户
		if req.Role == middleware.RoleAdmin || req.Role == middleware.RoleSysMgr {
			if !user.IsSystemAdmin() {
				httpx.ErrorCtx(r.Context(), w, errors.New("只有系统管理员可以创建系统角色"))
				return
			}
		}

		// 公司管理员可以创建公司管理员、分支机构管理员、营销管理人员、普通营销人员
		if req.Role == middleware.RoleCompanyMgr {
			if !user.IsSystemAdmin() {
				httpx.ErrorCtx(r.Context(), w, errors.New("只有系统管理员可以创建公司管理员"))
				return
			}
		}

		l := logic.NewCreateUserLogic(r.Context(), svcCtx)
		resp, err := l.CreateUser(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func updateUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin, middleware.RoleCompanyMgr, middleware.RoleBranchMgr, middleware.RoleMgr)

		if !ok {
			return
		}
		_ = user
		var req types.UpdateUserReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewUpdateUserLogic(r.Context(), svcCtx)
		resp, err := l.UpdateUser(getPathParam(r, "id"), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func deleteUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin, middleware.RoleCompanyMgr)

		if !ok {
			return
		}
		_ = user
		l := logic.NewDeleteUserLogic(r.Context(), svcCtx)
		err := l.DeleteUser(getPathParam(r, "id"), user.ID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{"success": true, "message": "User deleted"})
		}
	}
}

func resetPasswordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin)
		if !ok {
			return
		}
		_ = user
		var req types.ResetPasswordReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewResetPasswordLogic(r.Context(), svcCtx)
		err := l.ResetPassword(getPathParam(r, "id"), req.Password)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{"success": true, "message": "Password reset successfully"})
		}
	}
}

// ==================== 活动管理 ====================

func getActivityListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := getCurrentUser(w, r, svcCtx)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		var req logic.GetActivityListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 公司管理员 / 系统管理员：自动按公司ID过滤（不传则用用户所属公司）
		if user.IsCompanyAdmin() {
			if req.CompanyID == "" {
				req.CompanyID = user.CompanyID
			}
			// 不允许跨公司查询
			if req.CompanyID != user.CompanyID && !user.IsSystemAdmin() {
				req.CompanyID = user.CompanyID
			}
		}
		// 分支机构管理员 / 营销管理人员：自动按机构ID过滤
		if user.IsBranchAdmin() || user.IsMgr() {
			if req.BranchID == "" {
				req.BranchID = user.BranchID
			}
		}

		l := logic.NewGetActivityListLogic(r.Context(), svcCtx)
		resp, err := l.GetActivityList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func getActivityHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetActivityLogic(r.Context(), svcCtx)
		resp, err := l.GetActivity(getPathParam(r, "id"))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func createActivityHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin, middleware.RoleCompanyMgr, middleware.RoleBranchMgr, middleware.RoleMgr)
		
				if !ok {
			return
		}
		_ = user
		var req types.CreateActivityReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 公司管理员：只能创建本公司下的活动
		if user.IsCompanyAdmin() && !user.IsSystemAdmin() {
			if req.CompanyID != "" && req.CompanyID != user.CompanyID {
				httpx.ErrorCtx(r.Context(), w, errors.New("不能为其他公司创建活动"))
				return
			}
			if req.CompanyID == "" && req.BranchID != "" {
				// 需要校验 branch 属于本公司
				req.CompanyID = user.CompanyID
			}
			// 如果都没传，自动填本公司ID
			if req.CompanyID == "" && req.BranchID == "" {
				req.CompanyID = user.CompanyID
			}
		}
		// 分支机构管理员 / 营销管理人员：只能创建本机构的活动
		if user.IsBranchAdmin() || user.IsMgr() {
			if user.BranchID == "" {
				httpx.ErrorCtx(r.Context(), w, errors.New("您没有所属分支机构，无法创建活动"))
				return
			}
			req.BranchID = user.BranchID
			req.CompanyID = user.CompanyID
		}

		l := logic.NewCreateActivityLogic(r.Context(), svcCtx)
		resp, err := l.CreateActivity(&req, user.ID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func updateActivityHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin, middleware.RoleCompanyMgr, middleware.RoleBranchMgr, middleware.RoleMgr)
		
				if !ok {
			return
		}
		_ = user
		var req types.UpdateActivityReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 公司层面活动（company_id 有值）只能由公司管理员或创建人编辑
		// 分支机构活动只能由本机构的管理人员或创建人编辑
		activityID := getPathParam(r, "id")
		al := logic.NewGetActivityLogic(r.Context(), svcCtx)
		activity, err := al.GetActivity(activityID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 非系统/公司管理员，不能编辑公司层面活动
		if activity.CompanyID != "" && !user.IsSystemAdmin() && !user.IsCompanyAdmin() {
			httpx.ErrorCtx(r.Context(), w, errors.New("权限不足：不能编辑公司层面活动"))
			return
		}
		// 非系统/公司管理员 且 非本机构人员，不能编辑其他机构的活动
		if activity.BranchID != "" && activity.BranchID != user.BranchID && !user.IsSystemAdmin() && !user.IsCompanyAdmin() {
			httpx.ErrorCtx(r.Context(), w, errors.New("权限不足：不能编辑其他机构的活动"))
			return
		}
		// 如果是非管理员，需要检查创建人
		if !user.IsSystemAdmin() && !user.IsCompanyAdmin() && activity.CreatedBy != user.ID {
			httpx.ErrorCtx(r.Context(), w, errors.New("权限不足：只能编辑自己创建的活动"))
			return
		}

		l := logic.NewUpdateActivityLogic(r.Context(), svcCtx)
		resp, err := l.UpdateActivity(activityID, &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func deleteActivityHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin, middleware.RoleBranchMgr)
		
		if !ok {
			return
		}

		// 营销管理人员不能删除活动
		if user.Role == middleware.RoleMgr {
			httpx.ErrorCtx(r.Context(), w, errors.New("权限不足：营销管理人员不能删除活动"))
			return
		}

		// 系统管理员可以删除任何活动
		// 公司管理员可以删除本公司活动
		// 分支机构管理员只能删除本机构创建的活动
		activityID := getPathParam(r, "id")
		if !user.IsSystemAdmin() {
			al := logic.NewGetActivityLogic(r.Context(), svcCtx)
			activity, err := al.GetActivity(activityID)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
				return
			}

			// 公司管理员：可以删除本公司活动
			if user.IsCompanyAdmin() {
				if activity.CompanyID != user.CompanyID {
					httpx.ErrorCtx(r.Context(), w, errors.New("权限不足：只能删除本公司创建的活动"))
					return
				}
			} else if user.IsBranchAdmin() {
				// 分支机构管理员：只能删除本机构创建的活动
				if activity.BranchID != user.BranchID {
					httpx.ErrorCtx(r.Context(), w, errors.New("权限不足：只能删除本机构创建的活动"))
					return
				}
			} else {
				// 其他角色不能删除
				httpx.ErrorCtx(r.Context(), w, errors.New("权限不足"))
				return
			}
		}

		l := logic.NewDeleteActivityLogic(r.Context(), svcCtx)
		err := l.DeleteActivity(activityID, user.ID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{"success": true, "message": "Activity deleted"})
		}
	}
}

func updateActivityStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin, middleware.RoleMgr)
		
				if !ok {
			return
		}
		_ = user
		id := getPathParam(r, "id")
		var req struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{"success": false, "message": "参数错误"})
			return
		}

		validStatuses := map[string]bool{"active": true, "closed": true, "draft": true}
		if req.Status == "" || !validStatuses[req.Status] {
			httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{"success": false, "message": "无效的状态值"})
			return
		}

		l := logic.NewUpdateActivityStatusLogic(r.Context(), svcCtx)
		err := l.UpdateActivityStatus(id, req.Status, user.ID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{"success": true, "message": "状态更新成功"})
		}
	}
}

func getActivityStatsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetActivityStatsLogic(r.Context(), svcCtx)
		resp, err := l.GetActivityStats(getPathParam(r, "id"))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

// ==================== 表单字段 ====================

func getFormFieldsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		activityId := getPathParam(r, "id")
		l := logic.NewGetFormFieldsLogic(r.Context(), svcCtx)
		resp, err := l.GetFormFields(activityId)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func createFormFieldHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin, middleware.RoleMgr)
		
				if !ok {
			return
		}
		_ = user
		activityId := getPathParam(r, "id")
		var req types.CreateFormFieldReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewCreateFormFieldLogic(r.Context(), svcCtx)
		resp, err := l.CreateFormField(activityId, &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func updateFormFieldHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin, middleware.RoleMgr)
		
				if !ok {
			return
		}
		_ = user
		var req types.UpdateFormFieldReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewUpdateFormFieldLogic(r.Context(), svcCtx)
		resp, err := l.UpdateFormField(getPathParam(r, "id"), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func deleteFormFieldHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin, middleware.RoleMgr)
		
				if !ok {
			return
		}
		_ = user
		l := logic.NewDeleteFormFieldLogic(r.Context(), svcCtx)
		err := l.DeleteFormField(getPathParam(r, "id"))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{"success": true, "message": "Field deleted"})
		}
	}
}

// ==================== 报名管理 ====================

func getRegistrationsByActivityHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := getCurrentUser(w, r, svcCtx)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		activityId := getPathParam(r, "id")
		var req types.GetRegistrationListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		req.ActivityID = activityId

		// sales角色只能看自己名下的报名
		if user.IsSales() {
			req.SalesID = user.ID
		}

		l := logic.NewGetRegistrationListLogic(r.Context(), svcCtx)
		resp, err := l.GetRegistrationList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func getRegistrationListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := getCurrentUser(w, r, svcCtx)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		var req types.GetRegistrationListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if req.Page < 1 {
			req.Page = 1
		}
		if req.PageSize < 1 || req.PageSize > 100 {
			req.PageSize = 20
		}

		// sales角色只能看自己名下的报名
		if user.IsSales() {
			req.SalesID = user.ID
		}

		l := logic.NewGetRegistrationListLogic(r.Context(), svcCtx)
		resp, err := l.GetRegistrationList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func getRegistrationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetRegistrationListLogic(r.Context(), svcCtx)
		resp, err := l.GetRegistration(getPathParam(r, "id"))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func createRegistrationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Try new format first
		var reqV2 types.CreateRegistrationReqV2
		if err := httpx.Parse(r, &reqV2); err == nil && reqV2.ActivityID != "" {
			l := logic.NewCreateRegistrationLogic(r.Context(), svcCtx)
			resp, err := l.CreateRegistrationV2(&reqV2)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
			} else {
				httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
			}
			return
		}
		
		// Fall back to old format
		var req types.CreateRegistrationReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewCreateRegistrationLogic(r.Context(), svcCtx)
		resp, err := l.CreateRegistration(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func updateRegistrationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateRegistrationReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewUpdateRegistrationLogic(r.Context(), svcCtx)
		resp, err := l.UpdateRegistration(getPathParam(r, "id"), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

// ==================== 管理员报名管理 ====================

func getAdminRegistrationListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := getCurrentUser(w, r, svcCtx)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 非 admin/mgr 不可访问管理员报名接口
		if !user.IsMgr() {
			httpx.ErrorCtx(r.Context(), w, errors.New("权限不足"))
			return
		}

		page := 1
		pageSize := 20

		if p := r.URL.Query().Get("page"); p != "" {
			fmt.Sscanf(p, "%d", &page)
		}
		if ps := r.URL.Query().Get("pageSize"); ps != "" {
			fmt.Sscanf(ps, "%d", &pageSize)
		}
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		req := &types.GetRegistrationListReq{
			Page:     page,
			PageSize: pageSize,
		}
		if activityId := r.URL.Query().Get("activityId"); activityId != "" && activityId != "undefined" {
			req.ActivityID = activityId
		}

		l := logic.NewGetRegistrationListLogic(r.Context(), svcCtx)
		resp, err := l.GetRegistrationList(req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func adminReviewRegistrationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin, middleware.RoleMgr)
		
				if !ok {
			return
		}
		_ = user
		status := r.URL.Query().Get("status")
		if status == "" {
			// Try to get from body
			var req types.AdminReviewRegistrationReq
			if err := httpx.Parse(r, &req); err == nil && req.Status != "" {
				status = req.Status
			}
		}

		l := logic.NewUpdateRegistrationLogic(r.Context(), svcCtx)
		resp, err := l.UpdateRegistration(getPathParam(r, "id"), &types.UpdateRegistrationReq{
			Status: status,
		})
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func adminDeleteRegistrationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin, middleware.RoleMgr)
		
				if !ok {
			return
		}
		_ = user
		l := logic.NewUpdateRegistrationLogic(r.Context(), svcCtx)
		err := l.DeleteRegistration(getPathParam(r, "id"), user.ID, user.Role)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(map[string]interface{}{"success": true}))
		}
	}
}

// ==================== 分享 ====================

func generateShareHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GenerateShareReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		
		// Get current user ID (optional, for sales tracking)
		salesID := ""
		userId, err := getCurrentUserID(w, r, svcCtx)
		if err == nil {
			salesID = userId
		}
		
		l := logic.NewGenerateShareLogic(r.Context(), svcCtx)
		resp, err := l.GenerateShare(&req, salesID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func getVisitorActivityHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shareId := getPathParam(r, "shareId")
		l := logic.NewGetVisitorActivityLogic(r.Context(), svcCtx)
		resp, err := l.GetVisitorActivity(shareId)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

// ==================== 公司管理 ====================

func getCompanyListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin)
		if !ok {
			return
		}
		_ = user
		var req types.PaginationReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewGetCompanyListLogic(r.Context(), svcCtx)
		resp, err := l.GetCompanyList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func getCompanyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin)
		if !ok {
			return
		}
		_ = user
		l := logic.NewGetCompanyLogic(r.Context(), svcCtx)
		resp, err := l.GetCompany(getPathParam(r, "id"))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func createCompanyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin)
		if !ok {
			return
		}
		_ = user
		var req types.CreateCompanyReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewCreateCompanyLogic(r.Context(), svcCtx)
		resp, err := l.CreateCompany(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func updateCompanyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin)
		if !ok {
			return
		}
		_ = user
		var req types.UpdateCompanyReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewUpdateCompanyLogic(r.Context(), svcCtx)
		resp, err := l.UpdateCompany(getPathParam(r, "id"), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func deleteCompanyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin)
		if !ok {
			return
		}
		_ = user
		l := logic.NewDeleteCompanyLogic(r.Context(), svcCtx)
		err := l.DeleteCompany(getPathParam(r, "id"))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{"success": true, "message": "Company deleted"})
		}
	}
}

// ==================== 分支机构管理 ====================

func getBranchListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin, middleware.RoleCompanyMgr)
		if !ok {
			return
		}
		_ = user
		var req logic.GetBranchListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewGetBranchListLogic(r.Context(), svcCtx)
		resp, err := l.GetBranchList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func getBranchHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin, middleware.RoleCompanyMgr)
		if !ok {
			return
		}
		_ = user
		l := logic.NewGetBranchLogic(r.Context(), svcCtx)
		resp, err := l.GetBranch(getPathParam(r, "id"))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func createBranchHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin, middleware.RoleCompanyMgr)
		if !ok {
			return
		}
		_ = user
		var req types.CreateBranchReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewCreateBranchLogic(r.Context(), svcCtx)
		resp, err := l.CreateBranch(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func updateBranchHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin, middleware.RoleCompanyMgr)
		if !ok {
			return
		}
		_ = user
		var req types.UpdateBranchReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewUpdateBranchLogic(r.Context(), svcCtx)
		resp, err := l.UpdateBranch(getPathParam(r, "id"), &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func deleteBranchHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireRole(w, r, svcCtx, middleware.RoleAdmin, middleware.RoleCompanyMgr)
		if !ok {
			return
		}
		_ = user
		l := logic.NewDeleteBranchLogic(r.Context(), svcCtx)
		err := l.DeleteBranch(getPathParam(r, "id"))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{"success": true, "message": "Branch deleted"})
		}
	}
}

// ==================== 管理员设置 ====================

func getSettingsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetSettingsLogic(r.Context(), svcCtx)
		resp, err := l.GetSettings()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func saveSettingsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SettingsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewSaveSettingsLogic(r.Context(), svcCtx)
		err := l.SaveSettings(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(map[string]interface{}{"success": true}))
		}
	}
}

// ==================== 游客报名查询 ====================

func getVisitorRegistrationsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phone := r.URL.Query().Get("phone")
		if phone == "" {
			httpx.ErrorCtx(r.Context(), w, errors.New("请提供手机号"))
			return
		}
		l := logic.NewGetVisitorRegistrationsLogic(r.Context(), svcCtx)
		resp, err := l.GetVisitorRegistrations(phone)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func updateVisitorRegistrationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := getPathParam(r, "id")
		var req types.UpdateVisitorRegistrationReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewUpdateVisitorRegistrationLogic(r.Context(), svcCtx)
		resp, err := l.UpdateVisitorRegistration(id, &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

// ==================== 文件上传 ====================

func uploadImageHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewUploadLogic(r.Context(), svcCtx)
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			httpx.ErrorCtx(r.Context(), w, errors.New("解析上传数据失败"))
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, errors.New("未找到上传文件"))
			return
		}
		defer file.Close()

		resp, err := l.UploadImage(header)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}

func uploadFileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewUploadLogic(r.Context(), svcCtx)
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			httpx.ErrorCtx(r.Context(), w, errors.New("解析上传数据失败"))
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, errors.New("未找到上传文件"))
			return
		}
		defer file.Close()

		resp, err := l.UploadFile(header)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, wrapResponse(resp))
		}
	}
}
