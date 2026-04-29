package temp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	coreerrors "github.com/kriuchkov/jiraforge/internal/core/errors"
	"github.com/kriuchkov/jiraforge/internal/core/models"
	"github.com/kriuchkov/jiraforge/internal/core/ports"
)

type Store struct {
	baseDir string
}

func NewStore() ports.AttachmentStore {
	return &Store{baseDir: filepath.Join(os.TempDir(), "jiraforge-attachments")}
}

func (s *Store) Save(_ context.Context, attachment models.AttachmentContent) (*models.StoredAttachment, error) {
	if err := os.MkdirAll(s.baseDir, 0o755); err != nil {
		return nil, coreerrors.Wrap(coreerrors.CodeInternal, "storage.temp.Save", "create attachment directory", err)
	}

	filename := strings.TrimSpace(attachment.Filename)
	if filename == "" {
		filename = fmt.Sprintf("attachment-%s", attachment.ID)
	}
	filename = strings.NewReplacer("/", "_", "\\", "_").Replace(filename)

	path := filepath.Join(s.baseDir, fmt.Sprintf("%s_%s", attachment.ID, filename))
	if err := os.WriteFile(path, attachment.Data, 0o644); err != nil {
		return nil, coreerrors.Wrap(coreerrors.CodeInternal, "storage.temp.Save", "write attachment data", err)
	}

	return &models.StoredAttachment{
		Path:     path,
		Filename: filename,
		MimeType: attachment.MimeType,
		Size:     attachment.Size,
	}, nil
}
