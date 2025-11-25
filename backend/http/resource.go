package http

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/SlepoyShaman/FileStorage/adapters/fs/files"
	"github.com/SlepoyShaman/FileStorage/adapters/fs/fileutils"
	"github.com/SlepoyShaman/FileStorage/common/errors"
	"github.com/SlepoyShaman/FileStorage/common/settings"
	"github.com/SlepoyShaman/FileStorage/common/utils"
	"github.com/SlepoyShaman/FileStorage/indexing"
	"github.com/SlepoyShaman/FileStorage/indexing/iteminfo"
	"github.com/SlepoyShaman/FileStorage/preview"
)

// resourceGetHandler retrieves information about a resource.
// @Summary Get resource information
// @Description Returns metadata and optionally file contents for a specified resource path.
// @Tags Resources
// @Accept json
// @Produce json
// @Param path query string true "Path to the resource"
// @Param source query string true "Source name for the desired source, default is used if not provided"
// @Param content query string false "Include file content if true"
// @Param checksum query string false "Optional checksum validation"
// @Success 200 {object} iteminfo.FileInfo "Resource metadata"
// @Failure 404 {object} map[string]string "Resource not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/resources [get]
func resourceGetHandler(w http.ResponseWriter, r *http.Request, d *requestContext) (int, error) {
	encodedPath := r.URL.Query().Get("path")
	source := r.URL.Query().Get("source")
	// Decode the URL-encoded path
	path, err := url.QueryUnescape(encodedPath)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("invalid path encoding: %v", err)
	}
	source, err = url.QueryUnescape(source)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("invalid source encoding: %v", err)
	}
	userscope, err := settings.GetScopeFromSourceName(d.user.Scopes, source)
	if err != nil {
		return http.StatusForbidden, err
	}
	userscope = strings.TrimRight(userscope, "/")
	scopePath := utils.JoinPathAsUnix(userscope, path)
	getContent := r.URL.Query().Get("content") == "true"
	fileInfo, err := files.FileInfoFaster(utils.FileOptions{
		Username:                 d.user.Username,
		Path:                     scopePath,
		Source:                   source,
		Expand:                   true,
		Content:                  getContent,
		Metadata:                 true,
		ExtractEmbeddedSubtitles: settings.Config.Integrations.Media.ExtractEmbeddedSubtitles,
	}, store.Access)
	if err != nil {
		return errToStatus(err), err
	}
	if !d.user.Permissions.Download && fileInfo.Content != "" {
		return http.StatusForbidden, fmt.Errorf("user is not allowed to get content, requires download permission")
	}
	if userscope != "/" {
		fileInfo.Path = strings.TrimPrefix(fileInfo.Path, userscope)
	}
	if fileInfo.Path == "" {
		fileInfo.Path = "/"
	}
	if fileInfo.Type == "directory" {
		return renderJSON(w, r, fileInfo)
	}
	if algo := r.URL.Query().Get("checksum"); algo != "" {
		idx := indexing.GetIndex(source)
		if idx == nil {
			return http.StatusNotFound, fmt.Errorf("source %s not found", source)
		}
		realPath, _, _ := idx.GetRealPath(userscope, path)
		checksum, err := utils.GetChecksum(realPath, algo)
		if err == errors.ErrInvalidOption {
			return http.StatusBadRequest, nil
		} else if err != nil {
			return http.StatusInternalServerError, err
		}
		fileInfo.Checksums = make(map[string]string)
		fileInfo.Checksums[algo] = checksum
	}
	return renderJSON(w, r, fileInfo)

}

// resourcePostHandler creates or uploads a new resource.
// @Summary Create or upload a resource
// @Description Creates a new resource or uploads a file at the specified path. Supports file uploads and directory creation.
// @Tags Resources
// @Accept json
// @Produce json
// @Param path query string true "url encoded destination path where to place the files inside the destination source, a directory must end in / to create a directory"
// @Param source query string true "Name for the desired filebrowser destination source name, default is used if not provided"
// @Param override query bool false "Override existing file if true"
// @Param isDir query bool false "Explicitly specify if the resource is a directory"
// @Success 200 "Resource created successfully"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Resource not found"
// @Failure 409 {object} map[string]string "Conflict - Resource already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/resources [post]
func resourcePostHandler(w http.ResponseWriter, r *http.Request, d *requestContext) (int, error) {
	path := r.URL.Query().Get("path")
	unescapedPath := path
	source := r.URL.Query().Get("source")
	var err error
	accessStore := store.Access
	// if share is not nil, then set accessStore to nil
	if d.share != nil {
		accessStore = nil
	} else {
		// decode url encoded source name
		source, err = url.QueryUnescape(source)
		if err != nil {
			slog.Debug("invalid source encoding: %v", err)
			return http.StatusBadRequest, fmt.Errorf("invalid source encoding: %v", err)
		}
		unescapedPath, err = url.QueryUnescape(path)
		if err != nil {
			slog.Debug("invalid path encoding: %v", err)
			return http.StatusBadRequest, fmt.Errorf("invalid path encoding: %v", err)
		}
		if !d.user.Permissions.Create {
			return http.StatusForbidden, fmt.Errorf("user is not allowed to create or modify")
		}
		userscope := ""
		// Determine if this is a directory or file based on trailing slash
		// Strip trailing slash from userscope to prevent double slashes
		userscope, err = settings.GetScopeFromSourceName(d.user.Scopes, source)
		if err != nil {
			slog.Debug("error getting scope from source name: %v", err)
			return http.StatusForbidden, err
		}
		userscope = strings.TrimRight(userscope, "/")
		path = utils.JoinPathAsUnix(userscope, unescapedPath)
	}

	// Determine if this is a directory based on isDir query param or trailing slash (for backwards compatibility)
	isDirParam := r.URL.Query().Get("isDir")
	isDir := isDirParam == "true" || strings.HasSuffix(unescapedPath, "/")
	fileOpts := utils.FileOptions{
		Username: d.user.Username,
		Path:     path,
		Source:   source,
		Expand:   false,
	}
	idx := indexing.GetIndex(source)
	if idx == nil {
		slog.Debug("source %s not found", source)
		return http.StatusNotFound, fmt.Errorf("source %s not found", source)
	}
	realPath, _, _ := idx.GetRealPath(path)

	// Check access control for the target path
	if accessStore != nil && !accessStore.Permitted(idx.Path, path, d.user.Username) {
		return http.StatusForbidden, fmt.Errorf("access denied to path %s", path)
	}

	// Check for file/folder conflicts before creation
	if stat, statErr := os.Stat(realPath); statErr == nil {
		// Path exists, check for type conflicts
		existingIsDir := stat.IsDir()
		requestingDir := isDir

		// If type mismatch (file vs folder or folder vs file) and not overriding
		if existingIsDir != requestingDir && r.URL.Query().Get("override") != "true" {
			slog.Debug("Type conflict detected in chunked: existing is dir=%v, requesting dir=%v at path=%v", existingIsDir, requestingDir, realPath)
			return http.StatusConflict, nil
		}
	}

	// Directories creation on POST.
	if isDir {
		err = files.WriteDirectory(fileOpts)
		if err != nil {
			slog.Debug("error writing directory: %v", err)
			return errToStatus(err), err
		}
		return http.StatusOK, nil
	}

	// Handle Chunked Uploads
	chunkOffsetStr := r.Header.Get("X-File-Chunk-Offset")
	if chunkOffsetStr != "" {
		var offset int64
		offset, err = strconv.ParseInt(chunkOffsetStr, 10, 64)
		if err != nil {
			slog.Debug("invalid chunk offset: %v", err)
			return http.StatusBadRequest, fmt.Errorf("invalid chunk offset: %v", err)
		}

		var totalSize int64
		totalSizeStr := r.Header.Get("X-File-Total-Size")
		totalSize, err = strconv.ParseInt(totalSizeStr, 10, 64)
		if err != nil {
			slog.Debug("invalid total size: %v", err)
			return http.StatusBadRequest, fmt.Errorf("invalid total size: %v", err)
		}
		// On the first chunk, check for conflicts or handle override
		if offset == 0 {
			// Check for file/folder conflicts for chunked uploads
			if stat, statErr := os.Stat(realPath); statErr == nil {
				existingIsDir := stat.IsDir()
				requestingDir := false // Files are never directories

				// If type mismatch (existing dir vs requesting file) and not overriding
				if existingIsDir != requestingDir && r.URL.Query().Get("override") != "true" {
					slog.Debug("Type conflict detected in chunked: existing is dir=%v, requesting dir=%v at path=%v", existingIsDir, requestingDir, realPath)
					return http.StatusConflict, nil
				}
			}

			var fileInfo *iteminfo.ExtendedFileInfo
			fileInfo, err = files.FileInfoFaster(fileOpts, accessStore)
			if err == nil { // File exists
				if r.URL.Query().Get("override") != "true" {
					slog.Debug("resource already exists: %v", fileInfo.RealPath)
					return http.StatusConflict, nil
				}
				// If overriding, delete existing thumbnails
				preview.DelThumbs(r.Context(), *fileInfo)
			}
		}

		// Use a temporary file in the cache directory for chunks.
		// Create a unique name for the temporary file to avoid collisions.
		hasher := md5.New()
		hasher.Write([]byte(realPath))
		uploadID := hex.EncodeToString(hasher.Sum(nil))
		tempFilePath := filepath.Join(settings.Config.Server.CacheDir, "uploads", uploadID)

		if err = os.MkdirAll(filepath.Dir(tempFilePath), fileutils.PermDir); err != nil {
			slog.Debug("could not create temp dir: %v", err)
			return http.StatusInternalServerError, fmt.Errorf("could not create temp dir: %v", err)
		}
		// Create or open the temporary file
		var outFile *os.File
		outFile, err = os.OpenFile(tempFilePath, os.O_CREATE|os.O_WRONLY, fileutils.PermFile)
		if err != nil {
			slog.Debug("could not open temp file: %v", err)
			return http.StatusInternalServerError, fmt.Errorf("could not open temp file: %v", err)
		}
		defer outFile.Close()

		// Seek to the correct offset to write the chunk
		_, err = outFile.Seek(offset, 0)
		if err != nil {
			slog.Debug("could not seek in temp file: %v", err)
			return http.StatusInternalServerError, fmt.Errorf("could not seek in temp file: %v", err)
		}

		// Write the request body (the chunk) to the file
		var chunkSize int64
		chunkSize, err = io.Copy(outFile, r.Body)
		if err != nil {
			slog.Debug("could not write chunk to temp file: %v", err)
			return http.StatusInternalServerError, fmt.Errorf("could not write chunk to temp file: %v", err)
		}
		// check if the file is complete
		if (offset + chunkSize) >= totalSize {
			// close file before moving
			outFile.Close()
			// Move the completed file from the temp location to the final destination
			err = fileutils.MoveFile(tempFilePath, realPath)
			if err != nil {
				slog.Debug("could not move file from %v to %v: %v", tempFilePath, realPath, err)
				return http.StatusInternalServerError, fmt.Errorf("could not move file from chunked folder to destination: %v", err)
			}
			go files.RefreshIndex(source, realPath, false, false) //nolint:errcheck
		}

		return http.StatusOK, nil
	}

	fileInfo, err := files.FileInfoFaster(fileOpts, accessStore)
	if err == nil { // File exists
		if r.URL.Query().Get("override") != "true" {
			slog.Debug("resource already exists: %v", fileInfo.RealPath)
			return http.StatusConflict, nil
		}
		// If overriding, delete existing thumbnails
		preview.DelThumbs(r.Context(), *fileInfo)
	}
	err = files.WriteFile(fileOpts.Source, fileOpts.Path, r.Body)
	if err != nil {
		slog.Debug("error writing file: %v", err)
		return errToStatus(err), err
	}
	return http.StatusOK, nil
}
