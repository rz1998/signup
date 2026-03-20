package logic

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
	"signup/svc"
	"signup/types"
)

// ==================== Share Logic ====================

type GenerateShareLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGenerateShareLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateShareLogic {
	return &GenerateShareLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateShareLogic) GenerateShare(req *types.GenerateShareReq, salesID string) (*types.ShareInfo, error) {
	if req.ActivityID == "" {
		return nil, errors.New("活动ID不能为空")
	}

	// Check if activity exists
	var activityName string
	err := l.svcCtx.DB.QueryRow("SELECT name FROM activities WHERE id = $1 AND deleted_at IS NULL", req.ActivityID).Scan(&activityName)
	if err != nil {
		return nil, errors.New("活动不存在")
	}

	// Generate unique share ID
	shareID := uuid.New().String()

	shareType := "link"
	if req.ShareType != "" {
		shareType = req.ShareType
	}

	// Determine base URL
	baseURL := req.BaseURL
	if baseURL == "" {
		baseURL = l.svcCtx.Config.BaseURL
		if baseURL == "" {
			baseURL = ""
		}
	}

	// Generate share URL
	shareURL := shareID
	if baseURL != "" {
		shareURL = baseURL + "/visitor/activity/" + shareID
	} else {
		shareURL = "/visitor/activity/" + shareID
	}

	// Create share record with share_url
	var result error
	if salesID != "" {
		result = nil
		_, result = l.svcCtx.DB.Exec(`
			INSERT INTO shares (id, activity_id, sales_id, share_type, share_url, visit_count, created_at)
			VALUES ($1, $2, $3, $4, $5, 0, CURRENT_TIMESTAMP)
		`, shareID, req.ActivityID, salesID, shareType, shareURL)
	} else {
		result = nil
		_, result = l.svcCtx.DB.Exec(`
			INSERT INTO shares (id, activity_id, share_type, share_url, visit_count, created_at)
			VALUES ($1, $2, $3, $4, 0, CURRENT_TIMESTAMP)
		`, shareID, req.ActivityID, shareType, shareURL)
	}
	if result != nil {
		return nil, errors.New("创建分享记录失败: " + result.Error())
	}

	// Generate QR code for qrcode type or if requested
	qrCodeImage := ""
	if shareType == "qrcode" || req.ShareType == "qrcode" {
		qrCodeImage, err = generateQRCode(shareURL)
		if err != nil {
			qrCodeImage = "" // Don't fail if QR code generation fails
		}
	}

	return &types.ShareInfo{
		ID:          shareID,
		ActivityID:  req.ActivityID,
		SalesID:     salesID,
		ShareType:   shareType,
		ShareURL:    shareURL,
		QRCodeImage: qrCodeImage,
		VisitCount:  0,
		CreatedAt:   time.Now().Format("2006-01-02T15:04:05Z"),
	}, nil
}

// generateQRCode generates a QR code as base64 PNG for the given URL
func generateQRCode(content string) (string, error) {
	pngData, err := qrcode.Encode(content, qrcode.Medium, 256)
	if err != nil {
		return "", fmt.Errorf("生成二维码失败: %v", err)
	}

	// Convert to data URL
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngData), nil
}

type GetVisitorActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetVisitorActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVisitorActivityLogic {
	return &GetVisitorActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetVisitorActivityLogic) GetVisitorActivity(shareID string) (*types.VisitorActivityResp, error) {
	var activityID, salesID, shareType sql.NullString
	
	// Get share record
	err := l.svcCtx.DB.QueryRow(`
		SELECT activity_id, sales_id, share_type FROM shares WHERE id = $1
	`, shareID).Scan(&activityID, &salesID, &shareType)
	if err != nil {
		return nil, errors.New("分享链接无效")
	}

	if !activityID.Valid {
		return nil, errors.New("分享链接无效")
	}

	// Increment visit count
	l.svcCtx.DB.Exec("UPDATE shares SET visit_count = visit_count + 1 WHERE id = $1", shareID)

	// Get activity details
	var activity types.ActivityInfo
	var startDate, endDate, coverImage, description, content sql.NullString
	err = l.svcCtx.DB.QueryRow(`
		SELECT id, name, description, cover_image, content, status, start_date, end_date, created_by, created_at, updated_at
		FROM activities WHERE id = $1 AND deleted_at IS NULL
	`, activityID.String).Scan(&activity.ID, &activity.Name, &description, &coverImage, &content, &activity.Status,
		&startDate, &endDate, &activity.CreatedBy, &activity.CreatedAt, &activity.UpdatedAt)
	if err != nil {
		return nil, errors.New("活动不存在")
	}

	activity.Description = description.String
	activity.CoverImage = coverImage.String
	activity.Content = content.String
	if startDate.Valid {
		activity.StartDate = startDate.String
	}
	if endDate.Valid {
		activity.EndDate = endDate.String
	}

	// Get form fields
	fields, err := NewGetFormFieldsLogic(l.ctx, l.svcCtx).GetFormFields(activityID.String)
	if err != nil {
		fields = []types.FormFieldInfo{}
	}

	return &types.VisitorActivityResp{
		ShareID:    shareID,
		SalesID:    salesID.String,
		Activity:   activity,
		FormFields: fields,
	}, nil
}

// Get share stats
type GetShareStatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetShareStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetShareStatsLogic {
	return &GetShareStatsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetShareStatsLogic) GetShareStats(shareID string) (map[string]interface{}, error) {
	var visitCount int
	var activityID sql.NullString

	err := l.svcCtx.DB.QueryRow("SELECT visit_count, activity_id FROM shares WHERE id = $1", shareID).Scan(&visitCount, &activityID)
	if err != nil {
		return nil, errors.New("分享记录不存在")
	}

	// Get registrations from this share
	var registrationCount int
	if activityID.Valid {
		l.svcCtx.DB.QueryRow("SELECT COUNT(*) FROM registrations WHERE activity_id = $1 AND sales_id = (SELECT sales_id FROM shares WHERE id = $2)", activityID.String, shareID).Scan(&registrationCount)
	}

	return map[string]interface{}{
		"shareId":           shareID,
		"visitCount":        visitCount,
		"registrationCount": registrationCount,
	}, nil
}
