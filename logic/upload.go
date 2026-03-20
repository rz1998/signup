package logic

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"signup/svc"
)

const (
	MaxFileSize       = 10 * 1024 * 1024 // 10MB
	uploadDir         = "uploads"
	imageDir          = "uploads/images"
	fileDir           = "uploads/files"
)

var allowedImageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
}

var allowedFileExts = map[string]bool{
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
}

type UploadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadLogic {
	return &UploadLogic{ctx: ctx, svcCtx: svcCtx}
}

type UploadResp struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

func (l *UploadLogic) ensureDirs() error {
	dirs := []string{uploadDir, imageDir, fileDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建上传目录失败: %w", err)
		}
	}
	return nil
}

func (l *UploadLogic) getFileURL(filename string, subdir string) string {
	baseURL := l.svcCtx.Config.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8082"
	}
	return fmt.Sprintf("%s/uploads/%s/%s", strings.TrimSuffix(baseURL, "/"), subdir, filename)
}

func (l *UploadLogic) UploadImage(file *multipart.FileHeader) (*UploadResp, error) {
	return l.uploadFile(file, "images", allowedImageExts)
}

func (l *UploadLogic) UploadFile(file *multipart.FileHeader) (*UploadResp, error) {
	return l.uploadFile(file, "files", allowedFileExts)
}

func (l *UploadLogic) uploadFile(file *multipart.FileHeader, subdir string, allowedExts map[string]bool) (*UploadResp, error) {
	if err := l.ensureDirs(); err != nil {
		return nil, err
	}

	// 检查文件大小
	if file.Size > MaxFileSize {
		return nil, fmt.Errorf("文件大小超过限制(最大10MB)")
	}

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		return nil, fmt.Errorf("文件没有扩展名")
	}

	if !allowedExts[ext] {
		return nil, fmt.Errorf("不支持的文件类型: %s", ext)
	}

	// 生成唯一文件名
	uuidStr := uuid.New().String()
	now := time.Now()
	filename := fmt.Sprintf("%s_%d%s", uuidStr[:8], now.UnixNano(), ext)

	// 构建保存路径
	savePath := filepath.Join(uploadDir, subdir, filename)

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("无法读取上传文件: %w", err)
	}
	defer src.Close()

	// 创建目标文件
	dst, err := os.Create(savePath)
	if err != nil {
		return nil, fmt.Errorf("无法创建目标文件: %w", err)
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	return &UploadResp{
		URL:      l.getFileURL(filename, subdir),
		Filename: filename,
		Size:     file.Size,
	}, nil
}

// ServeUploadedFile serves an uploaded file via HTTP
func (l *UploadLogic) ServeFile(w http.ResponseWriter, r *http.Request, subdir string, filename string) {
	filePath := filepath.Join(uploadDir, subdir, filename)
	http.ServeFile(w, r, filePath)
}
