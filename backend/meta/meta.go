package meta

import (
	"context"
	"log/slog"
	"strings"
	"time"
)

// MetadataExtractor интерфейс стратегии
type MetadataExtractor interface {
	CanHandle(fileType string) bool
	Extract(ctx context.Context, fileItem *File, realPath string, opts *Options) error
}

// AudioExtractor стратегия для аудио файлов
type AudioExtractor struct{}

func (e *AudioExtractor) CanHandle(fileType string) bool {
	return strings.HasPrefix(fileType, "audio")
}

func (e *AudioExtractor) Extract(ctx context.Context, fileItem *File, realPath string, opts *Options) error {
	return extractAudioMetadata(ctx, fileItem, realPath, opts.AlbumArt || opts.Content, opts.Metadata)
}

// VideoExtractor стратегия для видео файлов
type VideoExtractor struct{}

func (e *VideoExtractor) CanHandle(fileType string) bool {
	return strings.HasPrefix(fileType, "video")
}

func (e *VideoExtractor) Extract(ctx context.Context, fileItem *File, realPath string, opts *Options) error {
	return extractVideoMetadata(ctx, fileItem, realPath)
}

// MetadataProcessor контекст, использующий стратегии
type MetadataProcessor struct {
	extractors []MetadataExtractor
}

func NewMetadataProcessor() *MetadataProcessor {
	return &MetadataProcessor{
		extractors: []MetadataExtractor{
			&AudioExtractor{},
			&VideoExtractor{},
		},
	}
}

func (p *MetadataProcessor) ProcessFiles(response *Response, opts *Options) int {
	if !opts.Metadata {
		return 0
	}

	startTime := time.Now()
	metadataCount := 0
	ctx := context.Background()

	for i := range response.Files {
		fileItem := &response.Files[i]

		// Получаем реальный путь
		itemRealPath, _, err := index.GetRealPath(opts.Path, fileItem.Name)
		if err != nil {
			slog.Debug("failed to get real path for file: "+fileItem.Name, err)
			continue
		}

		// Применяем подходящую стратегию
		for _, extractor := range p.extractors {
			if extractor.CanHandle(fileItem.Type) {
				err := extractor.Extract(ctx, fileItem, itemRealPath, opts)
				if err != nil {
					slog.Debug("failed to extract metadata for file: "+fileItem.Name, err)
				} else {
					metadataCount++
				}
				break // Обрабатываем только первым подходящим экстрактором
			}
		}
	}

	if metadataCount > 0 {
		elapsed := time.Since(startTime)
		slog.Debug("Extracted metadata for %d audio/video files in %v (avg: %v per file)",
			metadataCount, elapsed, elapsed/time.Duration(metadataCount))
	}

	return metadataCount
}
