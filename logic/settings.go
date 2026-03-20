package logic

import (
	"context"
	"encoding/json"
	"errors"

	"signup/svc"
	"signup/types"
)

type GetSettingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSettingsLogic {
	return &GetSettingsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSettingsLogic) GetSettings() (*types.SettingsReq, error) {
	settings := &types.SettingsReq{
		HomeTitle:             "活动报名",
		HomeSubtitle:          "欢迎参加我们的活动",
		HomeDescription:       "这里汇集了各类精彩活动，等你来参与",
		RegistrationTitle:     "活动报名",
		RegistrationSuccessMsg: "恭喜您报名成功！",
		RegistrationNotice:   "请确保填写真实信息",
		BgColor:              "#FFFFFF",
		BgImage:              "",
		ContactPhone:          "",
		ContactEmail:          "",
	}

	// Try to get from database
	var settingsJSON string
	err := l.svcCtx.DB.QueryRow(`SELECT value FROM settings WHERE key = 'page_settings'`).Scan(&settingsJSON)
	if err == nil && settingsJSON != "" {
		var dbSettings types.SettingsReq
		if json.Unmarshal([]byte(settingsJSON), &dbSettings) == nil {
			settings = &dbSettings
		}
	}

	return settings, nil
}

type SaveSettingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSaveSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveSettingsLogic {
	return &SaveSettingsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SaveSettingsLogic) SaveSettings(req *types.SettingsReq) error {
	if req == nil {
		return errors.New("设置不能为空")
	}

	settingsJSON, err := json.Marshal(req)
	if err != nil {
		return errors.New("序列化设置失败")
	}

	_, err = l.svcCtx.DB.Exec(`
		INSERT INTO settings (key, value, updated_at)
		VALUES ('page_settings', $1, CURRENT_TIMESTAMP)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP
	`, string(settingsJSON))
	
	if err != nil {
		return errors.New("保存设置失败: " + err.Error())
	}

	return nil
}
