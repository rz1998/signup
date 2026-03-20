package logic

import (
	"context"

	"signup/svc"
)

type HealthLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHealthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HealthLogic {
	return &HealthLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HealthLogic) Health() (map[string]interface{}, error) {
	return map[string]interface{}{
		"status":  "ok",
		"message": "Signup API is running",
	}, nil
}
