package artworks

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"tapesonic/util"
	"time"

	"github.com/google/uuid"
)

type ArtworkService struct {
	artworks   *ArtworkStorage
	artworkDir string
}

func newArtworkService(artworks *ArtworkStorage, artworkDir string) *ArtworkService {
	return &ArtworkService{
		artworks:   artworks,
		artworkDir: artworkDir,
	}
}

func (svc *ArtworkService) CreateFromUrl(url string) (Artwork, error) {
	err := os.MkdirAll(svc.artworkDir, 0777)
	if err != nil {
		return Artwork{}, err
	}

	response, err := http.Get(url)
	if err != nil {
		return Artwork{}, err
	}
	defer response.Body.Close()

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return Artwork{}, err
	}

	hashBytes := sha256.Sum256(content)
	hashString := hex.EncodeToString(hashBytes[:])

	artwork := Artwork{
		Id:              uuid.New(),
		DeduplicationId: hashString,
		FilePath:        hashString,
		CreatedAt:       util.NewTimestampWrapper(time.Now()),
		UpdatedAt:       util.NewTimestampWrapper(time.Now()),
	}

	contentType := response.Header.Get("Content-Type")
	if contentType != "" {
		format := util.MediaTypeToFormat(contentType)
		artwork.Format = format
		artwork.FilePath = fmt.Sprintf("%s.%s", artwork.FilePath, format)
	}

	filePath := path.Join(svc.artworkDir, artwork.FilePath)
	err = os.WriteFile(filePath, content, 0777)
	if err != nil {
		return Artwork{}, err
	}

	return svc.artworks.Upsert(artwork)
}

func (svc *ArtworkService) GetFileDescriptor(id uuid.UUID) (ArtworkFileDescriptor, error) {
	artwork, err := svc.artworks.GetById(id)
	if err != nil {
		return ArtworkFileDescriptor{}, err
	}

	descriptor := ArtworkFileDescriptor{
		LocalPath: path.Join(svc.artworkDir, artwork.FilePath),
	}

	return descriptor, nil
}
