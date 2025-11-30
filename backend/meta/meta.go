package meta

import (
	"context"
	"log/slog"
	"strings"
	"time"
)

func handleAudioMetadata(ctx context.Context, fileItem *File, realPath string, opts *Options) error {
	return extractAudioMetadata(ctx, fileItem, realPath, opts.AlbumArt || opts.Content, opts.Metadata)
}

func handleVideoMetadata(ctx context.Context, fileItem *File, realPath string, opts *Options) error {
	return extractVideoMetadata(ctx, fileItem, realPath)
}

func processMetadata(response *Response, opts *Options) int {
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

		isAudio := strings.HasPrefix(fileItem.Type, "audio")
		isVideo := strings.HasPrefix(fileItem.Type, "video")

		var extractErr error
		switch {
		case isAudio:
			extractErr = handleAudioMetadata(ctx, fileItem, itemRealPath, opts)
		case isVideo:
			extractErr = handleVideoMetadata(ctx, fileItem, itemRealPath, opts)
		default:
			continue
		}

		if extractErr != nil {
			slog.Debug("failed to extract metadata for file: "+fileItem.Name, extractErr)
		} else if isAudio || isVideo {
			metadataCount++
		}
	}

	if metadataCount > 0 {
		elapsed := time.Since(startTime)
		slog.Debug("Extracted metadata for %d audio/video files in %v (avg: %v per file)",
			metadataCount, elapsed, elapsed/time.Duration(metadataCount))
	}

	return metadataCount
}
