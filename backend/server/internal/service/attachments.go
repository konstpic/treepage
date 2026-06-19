package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/konstpic/treepage/backend/pkg/models"
	"gorm.io/gorm"
)

const maxAttachmentBytes = 10 << 20 // 10 MiB

type AttachmentService struct {
	db  *gorm.DB
	dir string
}

func NewAttachmentService(db *gorm.DB, dir string) *AttachmentService {
	if dir == "" {
		dir = "/data/attachments"
	}
	return &AttachmentService{db: db, dir: dir}
}

func (s *AttachmentService) EnsureDir() error {
	return os.MkdirAll(s.dir, 0o750)
}

func (s *AttachmentService) ListByDocument(ctx context.Context, documentID string) ([]models.DocumentAttachment, error) {
	var items []models.DocumentAttachment
	err := s.db.WithContext(ctx).Where("document_id = ?", documentID).Order("created_at ASC").Find(&items).Error
	return items, err
}

func (s *AttachmentService) Upload(ctx context.Context, documentID, userID, filename, mimeType string, r io.Reader, size int64) (*models.DocumentAttachment, error) {
	if size <= 0 || size > maxAttachmentBytes {
		return nil, fmt.Errorf("attachment size must be 1 byte to %d bytes", maxAttachmentBytes)
	}
	filename = filepath.Base(strings.TrimSpace(filename))
	if filename == "" || filename == "." || filename == ".." {
		return nil, fmt.Errorf("invalid filename")
	}
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	if err := s.EnsureDir(); err != nil {
		return nil, err
	}
	key, err := randomStorageKey()
	if err != nil {
		return nil, err
	}
	key = key + "_" + filename
	dest := filepath.Join(s.dir, key)
	f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o640)
	if err != nil {
		return nil, err
	}
	written, err := io.Copy(f, io.LimitReader(r, maxAttachmentBytes+1))
	f.Close()
	if err != nil {
		os.Remove(dest)
		return nil, err
	}
	if written > maxAttachmentBytes {
		os.Remove(dest)
		return nil, fmt.Errorf("attachment too large")
	}
	att := models.DocumentAttachment{
		DocumentID: documentID,
		Filename:   filename,
		StorageKey: key,
		MimeType:   mimeType,
		SizeBytes:  written,
		UploadedBy: &userID,
	}
	if err := s.db.WithContext(ctx).Create(&att).Error; err != nil {
		os.Remove(dest)
		return nil, err
	}
	return &att, nil
}

func (s *AttachmentService) GetByID(ctx context.Context, attachmentID string) (*models.DocumentAttachment, error) {
	var att models.DocumentAttachment
	if err := s.db.WithContext(ctx).First(&att, "id = ?", attachmentID).Error; err != nil {
		return nil, err
	}
	return &att, nil
}

func (s *AttachmentService) Open(ctx context.Context, attachmentID string) (*models.DocumentAttachment, *os.File, error) {
	att, err := s.GetByID(ctx, attachmentID)
	if err != nil {
		return nil, nil, err
	}
	path := filepath.Join(s.dir, att.StorageKey)
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	return att, f, nil
}

func (s *AttachmentService) Delete(ctx context.Context, attachmentID string) error {
	att, err := s.GetByID(ctx, attachmentID)
	if err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Delete(att).Error; err != nil {
		return err
	}
	_ = os.Remove(filepath.Join(s.dir, att.StorageKey))
	return nil
}

func randomStorageKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *AttachmentService) PublicURL(attachmentID string) string {
	return "/api/attachments/" + attachmentID + "/download"
}
