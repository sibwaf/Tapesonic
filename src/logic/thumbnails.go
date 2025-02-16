package logic

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"tapesonic/storage"
	"tapesonic/util"

	"github.com/google/uuid"
)

type ThumbnailService struct {
	storage     *storage.ThumbnailStorage
	contentPath string
}

func NewThumbnailService(
	storage *storage.ThumbnailStorage,
	contentPath string,
) *ThumbnailService {
	return &ThumbnailService{
		storage:     storage,
		contentPath: contentPath,
	}
}

func (s *ThumbnailService) CreateFromUrl(url string) (storage.Thumbnail, error) {
	thumbnail := storage.Thumbnail{}

	err := os.MkdirAll(s.contentPath, 0777)
	if err != nil {
		return thumbnail, err
	}

	response, err := http.Get(url)
	if err != nil {
		return thumbnail, err
	}
	defer response.Body.Close()

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return thumbnail, err
	}

	hashBytes := sha256.Sum256(content)
	hashString := hex.EncodeToString(hashBytes[:])

	thumbnail.DeduplicationId = hashString
	thumbnail.FilePath = hashString

	contentType := response.Header.Get("Content-Type")
	if contentType != "" {
		format := util.MediaTypeToFormat(contentType)
		thumbnail.Format = format
		thumbnail.FilePath = fmt.Sprintf("%s.%s", thumbnail.FilePath, format)
	}

	filePath := path.Join(s.contentPath, thumbnail.FilePath)
	err = os.WriteFile(filePath, content, 0777)
	if err != nil {
		return thumbnail, err
	}

	return s.storage.Upsert(thumbnail)
}

func (s *ThumbnailService) GetListForApi(sourceIds []uuid.UUID) ([]storage.Thumbnail, error) {
	return s.storage.Search(sourceIds)
}

func (s *ThumbnailService) GetThumbnailContent(id uuid.UUID) (string, io.ReadCloser, error) {
	thumbnail, err := s.storage.GetById(id)
	if err != nil {
		return "", nil, err
	}

	filePath := path.Join(s.contentPath, thumbnail.FilePath)
	reader, err := os.Open(filePath)
	if err != nil {
		return "", nil, err
	}

	return util.FormatToMediaType(thumbnail.Format), reader, nil
}
