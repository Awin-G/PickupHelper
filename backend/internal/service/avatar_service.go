package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/repository"

	"github.com/jmoiron/sqlx"
	"golang.org/x/image/draw"
)

const (
	maxAvatarBytes    = 150 * 1024 // 150 KB max per avatar
	maxAvatarDim      = 512        // resize to max 512x512
	jpegQualityStart  = 65         // starting JPEG quality
	jpegQualityMin    = 25         // minimum before giving up
	jpegQualityStep   = 5          // quality reduction step
)

// AvatarService handles avatar upload, compression, and serving.
type AvatarService struct {
	userRepo repository.UserRepo
	db       *sqlx.DB
}

func NewAvatarService(ur repository.UserRepo, db *sqlx.DB) *AvatarService {
	return &AvatarService{userRepo: ur, db: db}
}

// UploadAvatar reads raw multipart image bytes, validates the image is square,
// resizes to maxAvatarDim, compresses to JPEG under maxAvatarBytes, and stores
// the binary in the users table.
func (s *AvatarService) UploadAvatar(ctx context.Context, userID int64, r io.Reader) error {
	raw, err := io.ReadAll(r)
	if err != nil {
		return apperrors.New(apperrors.ErrInvalidParam, "无法读取图片数据")
	}
	if len(raw) > 5*1024*1024 {
		return apperrors.New(apperrors.ErrPayloadTooLarge, "图片大小不能超过 5MB")
	}

	img, format, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return apperrors.New(apperrors.ErrInvalidParam, "图片格式不支持: "+format)
	}
	_ = format

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w != h {
		return apperrors.New(apperrors.ErrInvalidParam, "请上传正方形图片")
	}
	if w < 64 || h < 64 {
		return apperrors.New(apperrors.ErrInvalidParam, "图片尺寸至少 64x64")
	}

	// Resize to maxAvatarDim if needed.
	if w > maxAvatarDim {
		img = resizeSquare(img, maxAvatarDim)
	}

	// Compress to JPEG with binary quality search under maxAvatarBytes.
	compressed, err := compressJPEG(img, maxAvatarBytes, jpegQualityStart)
	if err != nil {
		return err
	}

	if err := s.userRepo.SaveAvatar(ctx, s.db, userID, compressed, "image/jpeg"); err != nil {
		return apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	return nil
}

// GetAvatar returns the avatar binary and content type for a user.
func (s *AvatarService) GetAvatar(ctx context.Context, userID int64) ([]byte, string, error) {
	u, err := s.userRepo.FindByID(ctx, s.db, userID)
	if err != nil {
		return nil, "", apperrors.New(apperrors.ErrNotFound, "用户不存在")
	}
	if len(u.AvatarData) == 0 {
		return nil, "", apperrors.New(apperrors.ErrNotFound, "未设置头像")
	}
	return u.AvatarData, u.AvatarContentType, nil
}

// compressJPEG compresses img to JPEG under maxBytes.
// Quality starts at startQuality and is reduced until size fits or limit reached.
func compressJPEG(img image.Image, maxBytes int, startQuality int) ([]byte, error) {
	q := startQuality
	for q >= jpegQualityMin {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: q}); err != nil {
			return nil, apperrors.New(apperrors.ErrInternal, "图片压缩失败")
		}
		if buf.Len() <= maxBytes {
			return buf.Bytes(), nil
		}
		q -= jpegQualityStep
	}
	return nil, apperrors.New(apperrors.ErrPayloadTooLarge,
		fmt.Sprintf("图片太大，无法压缩到 %dKB 以内", maxBytes/1024))
}

// resizeSquare downsamples img to a square of targetSize×targetSize.
func resizeSquare(img image.Image, targetSize int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, targetSize, targetSize))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	return dst
}

// guard for unused imports from image/* packages (used by image.Decode).
var (
	_ = png.Decode
	_ = gif.Decode
	_ = jpeg.Encode
)
