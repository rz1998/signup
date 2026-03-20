package types

// 通用请求/响应类型

type PaginationReq struct {
	Page     int    `json:"page,default=1"`
	PageSize int    `json:"pageSize,default=20"`
	Keyword  string `json:"keyword,optional"`
}

type PaginationResp struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalCount int `json:"totalCount"`
	TotalPages int `json:"totalPages"`
}

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResp struct {
	Token string      `json:"token"`
	User  *UserInfo   `json:"user"`
}

type UserInfo struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	FullName  string `json:"fullName"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
	CompanyID string `json:"companyId"`
	BranchID  string `json:"branchId"`
	ManagerID string `json:"managerId"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

// Company types
type CompanyInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type CreateCompanyReq struct {
	Name        string `json:"name"`
	Description string `json:"description,optional"`
	Status     string `json:"status,default=active"`
}

type UpdateCompanyReq struct {
	Name        string `json:"name,optional"`
	Description string `json:"description,optional"`
	Status     string `json:"status,optional"`
}

// Branch types
type BranchInfo struct {
	ID          string `json:"id"`
	CompanyID   string `json:"companyId"`
	Name        string `json:"name"`
	LeaderName  string `json:"leaderName"`
	LeaderPhone string `json:"leaderPhone"`
	LeaderID    string `json:"leaderId"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type CreateBranchReq struct {
	CompanyID   string `json:"companyId"`
	Name        string `json:"name"`
	LeaderName  string `json:"leaderName,optional"`
	LeaderPhone string `json:"leaderPhone,optional"`
	LeaderID    string `json:"leaderId,optional"`
	Status     string `json:"status,default=active"`
}

type UpdateBranchReq struct {
	Name        string `json:"name,optional"`
	LeaderName  string `json:"leaderName,optional"`
	LeaderPhone string `json:"leaderPhone,optional"`
	LeaderID    string `json:"leaderId,optional"`
	Status     string `json:"status,optional"`
}

type CreateUserReq struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	FullName  string `json:"fullName"`
	Email     string `json:"email,optional"`
	Phone     string `json:"phone,optional"`
	Role      string `json:"role,default=sales"`
	Status    string `json:"status,default=active"`
	CompanyID string `json:"companyId,optional"`
	BranchID  string `json:"branchId,optional"`
	ManagerID string `json:"managerId,optional"`
}

type UpdateUserReq struct {
	FullName  string `json:"fullName,optional"`
	Email     string `json:"email,optional"`
	Phone     string `json:"phone,optional"`
	Role      string `json:"role,optional"`
	CompanyID string `json:"companyId,optional"`
	BranchID  string `json:"branchId,optional"`
	ManagerID string `json:"managerId,optional"`
	Status    string `json:"status,optional"`
}

type ResetPasswordReq struct {
	Password string `json:"password"`
}

// 活动相关类型
type EventInfo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	Location    string `json:"location"`
	MaxParticipants int `json:"maxParticipants"`
	CurrentParticipants int `json:"currentParticipants"`
	Status      string `json:"status"`
	CreatedBy   string `json:"createdBy"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type CreateEventReq struct {
	Title          string `json:"title"`
	Description    string `json:"description"`
	StartTime      string `json:"startTime"`
	EndTime        string `json:"endTime"`
	Location       string `json:"location"`
	MaxParticipants int   `json:"maxParticipants"`
}

type UpdateEventReq struct {
	Title          string `json:"title,optional"`
	Description    string `json:"description,optional"`
	StartTime      string `json:"startTime,optional"`
	EndTime        string `json:"endTime,optional"`
	Location       string `json:"location,optional"`
	MaxParticipants int   `json:"maxParticipants,optional"`
	Status         string `json:"status,optional"`
}

// 报名相关类型
type RegistrationInfo struct {
	ID          string `json:"id"`
	EventID     string `json:"eventId"`
	UserID      string `json:"userId"`
	Username    string `json:"username"`
	FullName    string `json:"fullName"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Status      string `json:"status"`
	Remarks     string `json:"remarks"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type CreateRegistrationReq struct {
	EventID  string `json:"eventId"`
	UserID   string `json:"userId"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Remarks  string `json:"remarks,optional"`
}

type UpdateRegistrationReq struct {
	Status  string `json:"status,optional"`
	Remarks string `json:"remarks,optional"`
}

type AdminReviewRegistrationReq struct {
	Status string `json:"status"`
}

// 统计相关类型
type StatsResp struct {
	TotalUsers         int `json:"totalUsers"`
	ActiveUsers        int `json:"activeUsers"`
	TotalEvents        int `json:"totalEvents"`
	ActiveEvents       int `json:"activeEvents"`
	TotalRegistrations int `json:"totalRegistrations"`
	PendingRegistrations int `json:"pendingRegistrations"`
}

type EventStatsResp struct {
	EventID            string `json:"eventId"`
	TotalRegistrations int    `json:"totalRegistrations"`
	ConfirmedCount    int    `json:"confirmedCount"`
	PendingCount      int    `json:"pendingCount"`
	CancelledCount    int    `json:"cancelledCount"`
}

// Activity types (matching requirements)
type ActivityInfo struct {
	ID          string `json:"id"`
	CompanyID   string `json:"companyId"`
	BranchID    string `json:"branchId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CoverImage  string `json:"coverImage"`
	Content     string `json:"content"`
	Status      string `json:"status"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Location             string `json:"location"`
	MaxParticipants      int    `json:"maxParticipants"`
	CurrentParticipants int    `json:"currentParticipants"`
	PageTitle           string `json:"pageTitle"`
	PageSubtitle        string `json:"pageSubtitle"`
	PageDescription     string `json:"pageDescription"`
	RegistrationNotice string `json:"registrationNotice"`
	ContactPhone        string `json:"contactPhone"`
	ContactEmail        string `json:"contactEmail"`
	CreatedBy           string `json:"createdBy"`
	CreatedAt           string `json:"createdAt"`
	UpdatedAt           string `json:"updatedAt"`
}

type CreateActivityReq struct {
	CompanyID   string `json:"companyId,optional"`
	BranchID    string `json:"branchId,optional"`
	Name            string `json:"name"`
	Description    string `json:"description,optional"`
	CoverImage     string `json:"coverImage,optional"`
	Content        string `json:"content,optional"`
	Status         string `json:"status,optional"`
	StartDate      string `json:"startDate,optional"`
	EndDate        string `json:"endDate,optional"`
	Location       string `json:"location,optional"`
	MaxParticipants int    `json:"maxParticipants,optional"`
	PageTitle           string `json:"pageTitle,optional"`
	PageSubtitle        string `json:"pageSubtitle,optional"`
	PageDescription     string `json:"pageDescription,optional"`
	RegistrationNotice string `json:"registrationNotice,optional"`
	ContactPhone        string `json:"contactPhone,optional"`
	ContactEmail        string `json:"contactEmail,optional"`
}

type UpdateActivityReq struct {
	CompanyID   string `json:"companyId,optional"`
	BranchID   string `json:"branchId,optional"`
	Name        string `json:"name,optional"`
	Description string `json:"description,optional"`
	CoverImage  string `json:"coverImage,optional"`
	Content    string `json:"content,optional"`
	Status     string `json:"status,optional"`
	StartDate  string `json:"startDate,optional"`
	EndDate    string `json:"endDate,optional"`
	Location   string `json:"location,optional"`
	PageTitle           string `json:"pageTitle,optional"`
	PageSubtitle        string `json:"pageSubtitle,optional"`
	PageDescription     string `json:"pageDescription,optional"`
	RegistrationNotice string `json:"registrationNotice,optional"`
	ContactPhone        string `json:"contactPhone,optional"`
	ContactEmail        string `json:"contactEmail,optional"`
}

// Form Field types
type FormFieldInfo struct {
	ID         string `json:"id"`
	ActivityID string `json:"activityId"`
	FieldName  string `json:"fieldName"`
	FieldType  string `json:"fieldType"`
	IsRequired bool   `json:"isRequired"`
	Options    string `json:"options,optional"`
	SortOrder  int    `json:"sortOrder"`
	CreatedAt  string `json:"createdAt"`
}

type CreateFormFieldReq struct {
	FieldName  string `json:"fieldName"`
	FieldType  string `json:"fieldType"`
	IsRequired bool   `json:"isRequired,default=false"`
	Options    string `json:"options,optional"`
}

type UpdateFormFieldReq struct {
	FieldName  string `json:"fieldName,optional"`
	FieldType  string `json:"fieldType,optional"`
	IsRequired *bool  `json:"isRequired,optional"`
	Options    string `json:"options,optional"`
	SortOrder  *int   `json:"sortOrder,optional"`
}

// Registration types
type RegistrationInfoV2 struct {
	ID            string `json:"id"`
	ActivityID    string `json:"activityId"`
	BranchID      string `json:"branchId"`
	ActivityName  string `json:"activityName"`
	SalesID       string `json:"salesId"`
	VisitorOpenID string `json:"visitorOpenId"`
	FormData      string `json:"formData"`
	Status        string `json:"status"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type GetRegistrationListReq struct {
	Page       int
	PageSize   int
	ActivityID string `json:"activityId,optional"`
	SalesID    string `json:"salesId,optional"`
	Status     string `json:"status,optional"`
}

type CreateRegistrationReqV2 struct {
	ActivityID    string `json:"activityId"`
	Phone         string `json:"phone"`
	SalesID       string `json:"salesId,optional"`
	VisitorOpenID string `json:"visitorOpenId,optional"`
	FormData      string `json:"formData"`
	Status        string `json:"status,optional"`
}

type UpdateRegistrationReqV2 struct {
	FormData string `json:"formData,optional"`
	Status   string `json:"status,optional"`
}

// Share types
type ShareInfo struct {
	ID          string `json:"id"`
	ActivityID  string `json:"activityId"`
	SalesID     string `json:"salesId"`
	ShareType   string `json:"shareType"`
	ShareURL    string `json:"shareUrl"`
	QRCodeImage string `json:"qrCodeImage"`
	VisitCount  int    `json:"visitCount"`
	CreatedAt   string `json:"createdAt"`
}

type GenerateShareReq struct {
	ActivityID string `json:"activityId"`
	ShareType  string `json:"shareType,optional"`
	BaseURL    string `json:"baseUrl,optional"`
}

type VisitorActivityResp struct {
	ShareID    string         `json:"shareId"`
	SalesID    string         `json:"salesId"`
	Activity   ActivityInfo   `json:"activity"`
	FormFields []FormFieldInfo `json:"formFields"`
}

// Settings types
type SettingsReq struct {
	HomeTitle             string `json:"homeTitle,optional"`
	HomeSubtitle          string `json:"homeSubtitle,optional"`
	HomeDescription       string `json:"homeDescription,optional"`
	RegistrationTitle     string `json:"registrationTitle,optional"`
	RegistrationSuccessMsg string `json:"registrationSuccessMsg,optional"`
	RegistrationNotice   string `json:"registrationNotice,optional"`
	BgColor              string `json:"bgColor,optional"`
	BgImage              string `json:"bgImage,optional"`
	ContactPhone          string `json:"contactPhone,optional"`
	ContactEmail          string `json:"contactEmail,optional"`
}

// 游客修改报名
type UpdateVisitorRegistrationReq struct {
	Phone    string `json:"phone"`
	FormData string `json:"formData"`
	Remark   string `json:"remark"`
}
