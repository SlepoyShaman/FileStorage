package files

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"time"
	"unicode"
	"unicode/utf8"

	"os"
	"path/filepath"
	"strings"

	"github.com/SlepoyShaman/FileStorage/adapters/fs/fileutils"
	"github.com/SlepoyShaman/FileStorage/common/settings"
	"github.com/SlepoyShaman/FileStorage/common/utils"
	"github.com/SlepoyShaman/FileStorage/database/access"
	"github.com/SlepoyShaman/FileStorage/ffmpeg"
	"github.com/SlepoyShaman/FileStorage/indexing"
	"github.com/SlepoyShaman/FileStorage/indexing/iteminfo"
	"github.com/dhowden/tag"
)

func FileInfoFaster(opts utils.FileOptions, access *access.Storage) (*iteminfo.ExtendedFileInfo, error) {
	response := &iteminfo.ExtendedFileInfo{}
	index := indexing.GetIndex(opts.Source)
	if index == nil {
		return response, fmt.Errorf("could not get index: %v ", opts.Source)
	}
	if !strings.HasPrefix(opts.Path, "/") {
		opts.Path = "/" + opts.Path
	}
	realPath, isDir, err := index.GetRealPath(opts.Path)
	if err != nil {
		return response, fmt.Errorf("could not get real path for requested path: %v, error: %v", opts.Path, err)
	}
	if !strings.HasSuffix(opts.Path, "/") && isDir {
		opts.Path = opts.Path + "/"
	}
	opts.IsDir = isDir
	var info *iteminfo.FileInfo

	// Check if path is viewable (allows filesystem access without indexing)
	isViewable := index.IsViewable(isDir, opts.Path)

	// For non-viewable paths, verify they are indexed
	// Skip this check if indexing is disabled for the entire source
	if !isViewable && !index.Config.DisableIndexing {
		err = index.RefreshFileInfo(opts)
		if err != nil {
			return response, fmt.Errorf("path not accessible: %v", err)
		}
	}

	if isDir {
		info, err = index.GetFsDirInfo(opts.Path)
		if err != nil {
			return response, err
		}
	} else {
		// For files, get info from parent directory to ensure HasPreview is set correctly
		info, err = index.GetFsDirInfo(opts.Path)
		if err != nil {
			return response, err
		}
	}

	response.FileInfo = *info
	response.RealPath = realPath
	response.Source = opts.Source

	if access != nil && !access.Permitted(index.Path, opts.Path, opts.Username) {
		// User doesn't have access to the current folder, but check if they have access to any subitems
		// This allows specific allow rules on subfolders/files to work even when parent is denied
		err := access.CheckChildItemAccess(response, index, opts.Username)
		if err != nil {
			return response, err
		}
	}

	if isDir && opts.Metadata {
		startTime := time.Now()
		metadataCount := 0
		ctx := context.Background()

		for i := range response.Files {
			fileItem := &response.Files[i]

			isItemAudio := strings.HasPrefix(fileItem.Type, "audio")
			isItemVideo := strings.HasPrefix(fileItem.Type, "video")
			isMediaFile := isItemAudio || isItemVideo

			if isMediaFile {
				itemRealPath, _, err := index.GetRealPath(opts.Path, fileItem.Name)
				if err != nil {
					slog.Debug("failed to get real path for file: "+fileItem.Name, err)
					continue
				}

				if isItemAudio {
					shouldExtractArt := opts.AlbumArt || opts.Content
					err := extractAudioMetadata(ctx, fileItem, itemRealPath, shouldExtractArt, opts.Metadata)
					if err != nil {
						slog.Debug("failed to extract metadata for file: "+fileItem.Name, err)
					} else {
						metadataCount++
					}
				}

				if isItemVideo {
					err := extractVideoMetadata(ctx, fileItem, itemRealPath)
					if err != nil {
						slog.Debug("failed to extract video metadata for file: "+fileItem.Name, err)
					} else {
						metadataCount++
					}
				}
			}
		}

		if metadataCount > 0 {
			elapsed := time.Since(startTime)
			slog.Debug("Extracted metadata for %d audio/video files in %v (avg: %v per file)",
				metadataCount, elapsed, elapsed/time.Duration(metadataCount))
		}
	}

	// Extract content/metadata when explicitly requested OR for single file audio/video requests
	isAudioVideo := strings.HasPrefix(info.Type, "audio") || strings.HasPrefix(info.Type, "video")
	if opts.Content || opts.Metadata || (!isDir && isAudioVideo) {
		processContent(response, index, opts)
	}

	if settings.Config.Integrations.OnlyOffice.Secret != "" && info.Type != "directory" && iteminfo.IsOnlyOffice(info.Name) {
		response.OnlyOfficeId = generateOfficeId(realPath)
	}

	return response, nil
}

func WriteDirectory(opts utils.FileOptions) error {
	idx := indexing.GetIndex(opts.Source)
	if idx == nil {
		return fmt.Errorf("could not get index: %v ", opts.Source)
	}
	realPath, _, _ := idx.GetRealPath(opts.Path)

	var stat os.FileInfo
	var err error
	// Check if the destination exists and is a file
	if stat, err = os.Stat(realPath); err == nil && !stat.IsDir() {
		// If it's a file and we're trying to create a directory, remove the file first
		err = os.Remove(realPath)
		if err != nil {
			return fmt.Errorf("could not remove existing file to create directory: %v", err)
		}
	}

	// Ensure the parent directories exist
	// Permissions are set by MkdirAll (subject to umask, which is usually acceptable)
	err = os.MkdirAll(realPath, fileutils.PermDir)
	if err != nil {
		return err
	}

	return RefreshIndex(idx.Name, opts.Path, true, true)
}

func RefreshIndex(source string, path string, isDir bool, recursive bool) error {
	idx := indexing.GetIndex(source)
	if idx == nil {
		return fmt.Errorf("could not get index: %v ", source)
	}
	if idx.Config.DisableIndexing {
		return nil
	}
	// Always normalize path using MakeIndexPath
	path = idx.MakeIndexPath(path)

	// MakeIndexPath always adds trailing slash, but for files we need to remove it
	if !isDir {
		path = strings.TrimSuffix(path, "/")
	}

	// Skip indexing for viewable paths (viewable: true means don't index, just allow FS access)
	if idx.IsViewable(isDir, path) {
		return nil
	}

	// For directories, check if the path exists on disk
	// If it doesn't exist, remove it from the index
	if isDir {
		realPath, _, err := idx.GetRealPath(path)
		if err == nil {
			// Check if the directory exists on disk
			if !Exists(realPath) {
				// Directory no longer exists, remove it from the index
				// This clears both Directories and DirectoriesLedger maps
				idx.DeleteMetadata(path, true, false)
				return nil
			}
		}
	}

	err := idx.RefreshFileInfo(utils.FileOptions{Path: path, IsDir: isDir, Recursive: recursive})
	return err
}

func WriteFile(source, path string, in io.Reader) error {
	// Strip trailing slash from realPath if it's meant to be a file
	realPath := strings.TrimRight(path, "/")
	// Ensure the parent directories exist
	parentDir := filepath.Dir(realPath)
	err := os.MkdirAll(parentDir, fileutils.PermDir)
	if err != nil {
		return err
	}
	var stat os.FileInfo
	// Check if the destination exists and is a directory
	if stat, err = os.Stat(realPath); err == nil && stat.IsDir() {
		// If it's a directory and we're trying to create a file, remove the directory first
		err = os.RemoveAll(realPath)
		if err != nil {
			return fmt.Errorf("could not remove existing directory to create file: %v", err)
		}
	}

	// Open the file for writing (create if it doesn't exist, truncate if it does)
	file, err := os.OpenFile(realPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileutils.PermFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy the contents from the reader to the file
	_, err = io.Copy(file, in)
	if err != nil {
		return err
	}

	// Explicitly set file permissions to bypass umask
	err = os.Chmod(realPath, fileutils.PermFile)
	return err
}

// getContent reads and returns the file content if it's considered an editable text file.
func getContent(realPath string) (string, error) {
	const headerSize = 4096
	// Thresholds for detecting binary-like content (these can be tuned)
	const maxNullBytesInHeaderAbs = 10    // Max absolute null bytes in header
	const maxNullByteRatioInHeader = 0.1  // Max 10% null bytes in header
	const maxNullByteRatioInFile = 0.05   // Max 5% null bytes in the entire file
	const maxNonPrintableRuneRatio = 0.05 // Max 5% non-printable runes in the entire file

	// Open file
	f, err := os.Open(realPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Read header
	headerBytes := make([]byte, headerSize)
	n, err := f.Read(headerBytes)
	if err != nil && err != io.EOF {
		return "", err
	}
	actualHeader := headerBytes[:n]

	// --- Start of new heuristic checks ---

	if n > 0 {
		// 1. Basic Check: Is the header valid UTF-8?
		// If not, it's unlikely an editable UTF-8 text file.
		if !utf8.Valid(actualHeader) {
			return "", nil // Not an error, just not the text file we want
		}

		// 2. Check for excessive null bytes in the header
		nullCountInHeader := 0
		for _, b := range actualHeader {
			if b == 0x00 {
				nullCountInHeader++
			}
		}
		// Reject if too many nulls absolutely or relatively in the header
		if nullCountInHeader > 0 { // Only perform check if there are any nulls
			if nullCountInHeader > maxNullBytesInHeaderAbs ||
				(float64(nullCountInHeader)/float64(n) > maxNullByteRatioInHeader) {
				return "", nil // Too many nulls in header
			}
		}

		// 3. Check for other non-text ASCII control characters in the header
		// (C0 controls excluding \t, \n, \r)
		for _, b := range actualHeader {
			if b < 0x20 && b != '\t' && b != '\n' && b != '\r' {
				return "", nil // Found problematic control character
			}
			// C1 control characters (0x80-0x9F) would be caught by utf8.Valid if part of invalid sequences,
			// or by the non-printable rune check later if they form valid (but undesirable) codepoints.
		}

		// Optional: Use http.DetectContentType for an additional check on the header
		// contentType := http.DetectContentType(actualHeader)
		// if !strings.HasPrefix(contentType, "text/") && contentType != "application/octet-stream" {
		//     // If it's clearly a non-text MIME type (e.g., "image/jpeg"), reject it.
		//     // "application/octet-stream" is ambiguous, so we rely on other heuristics.
		//     return "", nil
		// }
	}
	// --- End of new heuristic checks for header ---

	// Now read the full file (original logic)
	content, err := os.ReadFile(realPath)
	if err != nil {
		return "", err
	}
	// Handle empty file (original logic - returns specific string)
	if len(content) == 0 {
		return "empty-file-x6OlSil", nil
	}

	stringContent := string(content)

	// 4. Final UTF-8 validation for the entire file
	// (This is crucial as the header might be fine, but the rest of the file isn't)
	if !utf8.ValidString(stringContent) {
		return "", nil
	}

	// 5. Check for excessive null bytes in the entire file content
	if len(content) > 0 { // Check only for non-empty files
		totalNullCount := 0
		for _, b := range content {
			if b == 0x00 {
				totalNullCount++
			}
		}
		if float64(totalNullCount)/float64(len(content)) > maxNullByteRatioInFile {
			return "", nil // Too many nulls in the entire file
		}
	}

	// 6. Check for excessive non-printable runes in the entire file content
	// (Excluding tab, newline, carriage return, which are common in text files)
	if len(stringContent) > 0 { // Check only for non-empty strings
		nonPrintableRuneCount := 0
		totalRuneCount := 0
		for _, r := range stringContent {
			totalRuneCount++
			// unicode.IsPrint includes letters, numbers, punctuation, symbols, and spaces.
			// It excludes control characters. We explicitly allow \t, \n, \r.
			if !unicode.IsPrint(r) && r != '\t' && r != '\n' && r != '\r' {
				nonPrintableRuneCount++
			}
		}

		if totalRuneCount > 0 { // Avoid division by zero
			if float64(nonPrintableRuneCount)/float64(totalRuneCount) > maxNonPrintableRuneRatio {
				return "", nil // Too many non-printable runes
			}
		}
	}

	// The file has passed all checks and is considered editable text.
	return stringContent, nil
}

func IsNamedPipe(mode os.FileMode) bool {
	return mode&os.ModeNamedPipe != 0
}

func IsSymlink(mode os.FileMode) bool {
	return mode&os.ModeSymlink != 0
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// extractAudioMetadata extracts metadata from an audio file using dhowden/tag
// and optionally extracts duration using the ffmpeg service with concurrency control
func extractAudioMetadata(ctx context.Context, item *iteminfo.ExtendedItemInfo, realPath string, getArt bool, getDuration bool) error {
	file, err := os.Open(realPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Check file size first to prevent reading extremely large files
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	// Skip files larger than 300MB to prevent memory issues
	maxSize := int64(300)
	if fileInfo.Size() > maxSize*1024*1024 {
		return fmt.Errorf("file with size %d MB exceeds metadata check limit: %d MB", fileInfo.Size()/1024/1024, maxSize)
	}

	m, err := tag.ReadFrom(file)
	if err != nil {
		return err
	}

	item.Metadata = &iteminfo.MediaMetadata{
		Title:  m.Title(),
		Artist: m.Artist(),
		Album:  m.Album(),
		Year:   m.Year(),
		Genre:  m.Genre(),
	}

	// Extract track number
	track, _ := m.Track()
	item.Metadata.Track = track

	// Extract duration ONLY if explicitly requested using the ffmpeg VideoService
	// This respects concurrency limits and gracefully handles missing ffmpeg
	if getDuration {
		ffmpegService := ffmpeg.NewFFmpegService(5, false, "")
		if ffmpegService != nil {
			startTime := time.Now()
			if duration, err := ffmpegService.GetMediaDuration(ctx, realPath); err == nil {
				item.Metadata.Duration = int(duration)
				elapsed := time.Since(startTime)
				if elapsed > 50*time.Millisecond {
					slog.Debug("Duration extraction took %v for file: %s", elapsed, item.Name)
				}
			}
		}
	}

	if !getArt {
		return nil
	}

	// Extract album art and encode as base64 with strict size limits
	if picture := m.Picture(); picture != nil && picture.Data != nil {
		// More aggressive size limit to prevent memory issues (max 5MB)
		if len(picture.Data) <= 5*1024*1024 {
			item.Metadata.AlbumArt = base64.StdEncoding.EncodeToString(picture.Data)
		} else {
			slog.Debug("Skipping album art for %s: too large (%d bytes)", realPath, len(picture.Data))
		}
	}

	return nil
}

func processContent(info *iteminfo.ExtendedFileInfo, idx *indexing.Index, opts utils.FileOptions) {
	isVideo := strings.HasPrefix(info.Type, "video")
	isAudio := strings.HasPrefix(info.Type, "audio")
	isFolder := info.Type == "directory"
	if isFolder {
		return
	}

	if isVideo {
		// Extract duration for video
		extItem := &iteminfo.ExtendedItemInfo{
			ItemInfo: info.ItemInfo,
		}
		err := extractVideoMetadata(context.Background(), extItem, info.RealPath)
		if err != nil {
			slog.Debug("failed to extract video metadata for file: "+info.RealPath, info.Name, err)
		} else {
			info.Metadata = extItem.Metadata
		}

		// Handle subtitles if requested
		if opts.ExtractEmbeddedSubtitles {
			parentPath := filepath.Dir(info.Path)
			parentInfo, exists := idx.GetReducedMetadata(parentPath, true)
			if exists {
				info.DetectSubtitles(parentInfo)
				err := info.LoadSubtitleContent()
				if err != nil {
					slog.Debug("failed to load subtitle content: " + err.Error())
				}
			}
		}
		return
	}

	if isAudio {
		// Create an ExtendedItemInfo to hold the metadata
		extItem := &iteminfo.ExtendedItemInfo{
			ItemInfo: info.ItemInfo,
		}
		err := extractAudioMetadata(context.Background(), extItem, info.RealPath, opts.AlbumArt || opts.Content, opts.Metadata || opts.Content)
		if err != nil {
			slog.Debug("failed to extract audio metadata for file: "+info.RealPath, info.Name, err)
		} else {
			// Copy metadata to ExtendedFileInfo
			info.Metadata = extItem.Metadata
			info.HasPreview = extItem.Metadata != nil && extItem.Metadata.AlbumArt != ""
		}
		return
	}

	// Process text content for non-video, non-audio files
	if info.Size < 20*1024*1024 { // 20 megabytes in bytes
		content, err := getContent(info.RealPath)
		if err != nil {
			slog.Debug("could not get content for file: "+info.RealPath, info.Name, err)
			return
		}
		info.Content = content
	} else {
		slog.Debug("skipping large text file contents (20MB limit): "+info.Path, info.Name, info.Type)
	}
}

// extractVideoMetadata extracts duration from video files using ffprobe
func extractVideoMetadata(ctx context.Context, item *iteminfo.ExtendedItemInfo, realPath string) error {
	// Extract duration using the ffmpeg VideoService with concurrency control
	videoService := ffmpeg.NewFFmpegService(10, false, "")
	if videoService != nil {
		duration, err := videoService.GetMediaDuration(ctx, realPath)
		if err != nil {
			return err
		}
		if duration > 0 {
			item.Metadata = &iteminfo.MediaMetadata{
				Duration: int(duration),
			}
		}
		return nil
	}
	return nil
}

func generateOfficeId(realPath string) string {
	key, ok := utils.OnlyOfficeCache.Get(realPath)
	if !ok {
		timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		documentKey := utils.HashSHA256(realPath + timestamp)
		utils.OnlyOfficeCache.Set(realPath, documentKey)
		return documentKey
	}
	return key
}
