package http

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/SlepoyShaman/filebrowser/backend/adapters/fs/files"
	"github.com/SlepoyShaman/filebrowser/backend/adapters/fs/fileutils"
	"github.com/SlepoyShaman/filebrowser/backend/common/settings"
	"github.com/SlepoyShaman/filebrowser/backend/common/utils"
	"github.com/SlepoyShaman/filebrowser/backend/indexing"
	"github.com/SlepoyShaman/filebrowser/backend/indexing/iteminfo"
	"github.com/SlepoyShaman/go-logger/logger"
	"golang.org/x/time/rate"
)

type throttledReadSeeker struct {
	rs      io.ReadSeeker
	limiter *rate.Limiter
	ctx     context.Context
}

func newThrottledReadSeeker(rs io.ReadSeeker, limit rate.Limit, burst int, ctx context.Context) *throttledReadSeeker {
	return &throttledReadSeeker{
		rs:      rs,
		limiter: rate.NewLimiter(limit, burst),
		ctx:     ctx,
	}
}

func (r *throttledReadSeeker) Read(p []byte) (n int, err error) {
	n, err = r.rs.Read(p)
	if n > 0 {
		if waitErr := r.limiter.WaitN(r.ctx, n); waitErr != nil {
			if err == nil {
				err = waitErr
			}
		}
	}
	return
}

func (r *throttledReadSeeker) Seek(offset int64, whence int) (int64, error) {
	return r.rs.Seek(offset, whence)
}

func toASCIIFilename(fileName string) string {
	var result strings.Builder
	for _, r := range fileName {
		if r > 127 {
			result.WriteRune('_')
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func setContentDisposition(w http.ResponseWriter, r *http.Request, fileName string) {
	dispositionType := "attachment"
	if r.URL.Query().Get("inline") == "true" {
		dispositionType = "inline"
	}

	asciiFileName := toASCIIFilename(fileName)
	encodedFileName := url.PathEscape(fileName)

	w.Header().Set("Content-Disposition", fmt.Sprintf("%s; filename=%q; filename*=utf-8''%s", dispositionType, asciiFileName, encodedFileName))
}

// @Tags Resources
// @Accept json
// @Param files query string true "a list of files in the following format 'source::filename' and separated by '||' with additional items in the list. (required)"
// @Param inline query bool false "If true, sets 'Content-Disposition' to 'inline'. Otherwise, defaults to 'attachment'."
// @Param algo query string false "Compression algorithm for archiving multiple files or directories. Options: 'zip' and 'tar.gz'. Default is 'zip'."
// @Success 200 {file} file "Raw file or directory content, or archive for multiple files"
// @Failure 202 {object} map[string]string "Modify permissions required"
// @Failure 400 {object} map[string]string "Invalid request path"
// @Failure 404 {object} map[string]string "File or directory not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/raw [get]
func rawHandler(w http.ResponseWriter, r *http.Request, d *requestContext) (int, error) {
	files := r.URL.Query().Get("files")
	fileList := strings.Split(files, "||")
	return rawFilesHandler(w, r, d, fileList)
}

func addFile(path string, d *requestContext, tarWriter *tar.Writer, zipWriter *zip.Writer, flatten bool) error {
	splitFile := strings.Split(path, "::")
	if len(splitFile) != 2 {
		return fmt.Errorf("invalid file in files requested: %v", splitFile)
	}
	source := splitFile[0]
	path = splitFile[1]

	var err error
	if d.share == nil {
		var userScope string
		userScope, err = settings.GetScopeFromSourceName(d.user.Scopes, source)
		if err != nil {
			return fmt.Errorf("source %s is not available for user %s", source, d.user.Username)
		}
		path = utils.JoinPathAsUnix(userScope, path)
	}

	idx := indexing.GetIndex(source)
	if idx == nil {
		return fmt.Errorf("source %s is not available", source)
	}

	if d.share == nil && store.Access != nil {
		if !store.Access.Permitted(idx.Path, path, d.user.Username) {
			return nil
		}
	}

	_, err = files.FileInfoFaster(utils.FileOptions{
		Path:   path,
		Source: source,
		Expand: false,
	}, nil)
	if err != nil {
		return err
	}
	realPath, _, _ := idx.GetRealPath(path)
	info, err := os.Stat(realPath)
	if err != nil {
		return err
	}

	baseName := filepath.Base(realPath)

	if info.IsDir() {
		return filepath.Walk(realPath, func(filePath string, fileInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(realPath, filePath)
			if err != nil {
				return err
			}

			relPath = filepath.ToSlash(relPath)

			if relPath == "." {
				return nil
			}

			if d.share == nil {
				indexRelPath := filepath.Join(path, relPath)
				indexRelPath = filepath.ToSlash(indexRelPath)

				if !store.Access.Permitted(idx.Path, indexRelPath, d.user.Username) {
					if fileInfo.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}

			if !flatten {
				relPath = filepath.Join(baseName, relPath)
				relPath = filepath.ToSlash(relPath)
			}

			if fileInfo.IsDir() {
				if tarWriter != nil {
					header := &tar.Header{
						Name:     relPath + "/",
						Mode:     int64(fileutils.PermDir),
						Typeflag: tar.TypeDir,
						ModTime:  fileInfo.ModTime(),
					}
					return tarWriter.WriteHeader(header)
				}
				if zipWriter != nil {
					_, err := zipWriter.Create(relPath + "/")
					return err
				}
				return nil
			}
			return addSingleFile(filePath, relPath, zipWriter, tarWriter)
		})
	} else {
		return addSingleFile(realPath, baseName, zipWriter, tarWriter)
	}
}

func addSingleFile(realPath, archivePath string, zipWriter *zip.Writer, tarWriter *tar.Writer) error {
	file, err := os.Open(realPath)
	if err != nil {
		if strings.Contains(err.Error(), "is a directory") {
			return nil
		}
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	if info.IsDir() {
		return nil
	}

	if tarWriter != nil {
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(archivePath)
		if err = tarWriter.WriteHeader(header); err != nil {
			return err
		}
		_, err = io.Copy(tarWriter, file)
		return err
	}

	if zipWriter != nil {
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = archivePath
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, file)
		return err
	}

	return nil
}

func rawFilesHandler(w http.ResponseWriter, r *http.Request, d *requestContext, fileList []string) (int, error) {
	if !d.user.Permissions.Download && d.share == nil {
		return http.StatusForbidden, fmt.Errorf("user is not allowed to download")
	}
	splitFile := strings.Split(fileList[0], "::")
	if len(splitFile) != 2 {
		return http.StatusBadRequest, fmt.Errorf("invalid file in files request: %v", fileList[0])
	}

	firstFileSource := splitFile[0]
	firstFilePath := splitFile[1]
	var err error
	var userscope string
	fileName := filepath.Base(firstFilePath)

	isOnlyOffice := isOnlyOfficeCompatibleFile(fileName) && config.Integrations.OnlyOffice.Url != ""
	var documentId string
	var logContext *OnlyOfficeLogContext

	if d.share == nil {
		userscope, err = settings.GetScopeFromSourceName(d.user.Scopes, firstFileSource)
		if err != nil {
			// Send OnlyOffice error log if this was an OnlyOffice file
			if isOnlyOffice {
				// Try to get document ID for error logging
				idx := indexing.GetIndex(firstFileSource)
				if idx != nil {
					tempPath := utils.JoinPathAsUnix(userscope, firstFilePath)
					if realPath, _, realErr := idx.GetRealPath(tempPath); realErr == nil {
						if docId, _ := getOnlyOfficeId(realPath); docId != "" {
							if ctx := getOnlyOfficeLogContext(docId); ctx != nil {
								sendOnlyOfficeLogEvent(ctx, "ERROR", "download",
									fmt.Sprintf("OnlyOffice download failed - source not available: %s - %v", firstFilePath, err))
							}
						}
					}
				}
			}
			return http.StatusForbidden, err
		}
		firstFilePath = utils.JoinPathAsUnix(userscope, firstFilePath)
	}
	idx := indexing.GetIndex(firstFileSource)
	if idx == nil {
		if isOnlyOffice {
			sendOnlyOfficeLogEvent(logContext, "ERROR", "download",
				fmt.Sprintf("OnlyOffice download failed - source index not available: %s", firstFileSource))
		}
		return http.StatusInternalServerError, fmt.Errorf("source %s is not available", firstFileSource)
	}
	realPath, isDir, err := idx.GetRealPath(firstFilePath)
	if err != nil {
		// Send OnlyOffice error log if this was an OnlyOffice file
		if isOnlyOffice {
			if docId, _ := getOnlyOfficeId(realPath); docId != "" {
				if ctx := getOnlyOfficeLogContext(docId); ctx != nil {
					sendOnlyOfficeLogEvent(ctx, "ERROR", "download",
						fmt.Sprintf("OnlyOffice download failed - could not resolve path: %s - %v", firstFilePath, err))
				}
			}
		}
		return http.StatusInternalServerError, err
	}
	estimatedSize, err := computeArchiveSize(fileList, d)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if len(fileList) == 1 && !isDir {
		if isOnlyOffice {
			documentId, _ = getOnlyOfficeId(realPath)
			if documentId != "" {
				logContext = getOnlyOfficeLogContext(documentId)
			}
		}

		if d.share == nil && store.Access != nil {
			if !store.Access.Permitted(idx.Path, firstFilePath, d.user.Username) {
				logger.Debugf("user %s denied access to path %s", d.user.Username, firstFilePath)
				if isOnlyOffice && logContext != nil {
					sendOnlyOfficeLogEvent(logContext, "ERROR", "download",
						fmt.Sprintf("OnlyOffice download failed - access denied by rule: %s", firstFilePath))
				}
				return http.StatusForbidden, fmt.Errorf("access denied to path %s", firstFilePath)
			}
		}

		fd, err2 := os.Open(realPath)
		if err2 != nil {
			if isOnlyOffice && logContext != nil {
				sendOnlyOfficeLogEvent(logContext, "ERROR", "download",
					fmt.Sprintf("OnlyOffice download failed - could not open file: %s - %v", firstFilePath, err2))
			}
			return http.StatusInternalServerError, err2
		}
		defer fd.Close()

		fileInfo, err2 := fd.Stat()
		if err2 != nil {
			if isOnlyOffice && logContext != nil {
				sendOnlyOfficeLogEvent(logContext, "ERROR", "download",
					fmt.Sprintf("OnlyOffice download failed - could not get file info: %s - %v", firstFilePath, err2))
			}
			return http.StatusInternalServerError, err2
		}

		if isOnlyOffice && logContext != nil {
			logger.Infof("OnlyOffice Server is downloading file: %s (documentId: %s)",
				firstFilePath, documentId)

			sendOnlyOfficeLogEvent(logContext, "INFO", "download",
				fmt.Sprintf("OnlyOffice Server downloading file: %s", firstFilePath))
		}

		setContentDisposition(w, r, fileName)
		w.Header().Set("Cache-Control", "private")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
		sizeInMB := estimatedSize / 1024 / 1024
		if sizeInMB > 500 {
			logger.Debugf("User %v is downloading large (%d MB) file: %v", d.user.Username, sizeInMB, fileName)
		}
		var reader io.ReadSeeker = fd
		if d.share != nil && d.share.MaxBandwidth > 0 {
			limit := rate.Limit(d.share.MaxBandwidth * 1024)
			burst := d.share.MaxBandwidth * 1024
			reader = newThrottledReadSeeker(fd, limit, burst, r.Context())
		}
		http.ServeContent(w, r, fileName, fileInfo.ModTime(), reader)
		return 200, nil
	}

	if config.Server.MaxArchiveSizeGB > 0 {
		maxSize := config.Server.MaxArchiveSizeGB * 1024 * 1024 * 1024
		if estimatedSize > maxSize {
			return http.StatusRequestEntityTooLarge, fmt.Errorf("pre-archive combined size of files exceeds maximum limit of %d GB", config.Server.MaxArchiveSizeGB)
		}
	}
	algo := r.URL.Query().Get("algo")
	var extension string
	switch algo {
	case "zip", "true", "":
		extension = ".zip"
	case "tar.gz":
		extension = ".tar.gz"
	default:
		return http.StatusInternalServerError, errors.New("format not implemented")
	}

	baseDirName := filepath.Base(filepath.Dir(firstFilePath))
	if baseDirName == "" || baseDirName == "/" {
		baseDirName = "download"
	}
	if len(fileList) == 1 && isDir {
		baseDirName = filepath.Base(realPath)
	}
	originalFileName := baseDirName + extension

	archiveData := filepath.Join(config.Server.CacheDir, utils.InsecureRandomIdentifier(10))
	if extension == ".zip" {
		archiveData = archiveData + ".zip"
		err = createZip(d, archiveData, fileList...)
	} else {
		archiveData = archiveData + ".tar.gz"
		err = createTarGz(d, archiveData, fileList...)
	}
	if err != nil {
		return http.StatusInternalServerError, err
	}

	fd, err := os.Open(archiveData)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer fd.Close()

	fileInfo, err := fd.Stat()
	if err != nil {
		os.Remove(archiveData)
		return http.StatusInternalServerError, err
	}

	sizeInMB := fileInfo.Size() / 1024 / 1024
	if sizeInMB > 500 {
		logger.Debugf("User %v is downloading large (%d MB) file: %v", d.user.Username, sizeInMB, originalFileName)
	}

	setContentDisposition(w, r, originalFileName)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	w.Header().Set("Content-Type", "application/octet-stream")

	var reader io.Reader = fd
	if d.share != nil && d.share.MaxBandwidth > 0 {
		limit := rate.Limit(d.share.MaxBandwidth * 1024)
		burst := d.share.MaxBandwidth * 1024
		reader = newThrottledReadSeeker(fd, limit, burst, r.Context())
	}
	_, err = io.Copy(w, reader)
	os.Remove(archiveData)
	if err != nil {
		logger.Errorf("Failed to copy archive data to response: %v", err)
		return http.StatusInternalServerError, err
	}

	return 0, nil
}

func computeArchiveSize(fileList []string, d *requestContext) (int64, error) {
	var estimatedSize int64
	for _, fname := range fileList {
		splitFile := strings.Split(fname, "::")
		if len(splitFile) != 2 {
			return http.StatusBadRequest, fmt.Errorf("invalid file in files request: %v", fileList[0])
		}
		source := splitFile[0]
		path := splitFile[1]
		var err error
		idx := indexing.GetIndex(source)
		if idx == nil {
			return 0, fmt.Errorf("source %s is not available", source)
		}
		var userScope string
		if d.share == nil {
			userScope, err = settings.GetScopeFromSourceName(d.user.Scopes, source)
			if err != nil {
				return 0, fmt.Errorf("source %s is not available for user %s", source, d.user.Username)
			}
			path = utils.JoinPathAsUnix(userScope, path)

			if store.Access != nil && !store.Access.Permitted(idx.Path, path, d.user.Username) {
				continue
			}
		}
		realPath, isDir, err := idx.GetRealPath(path)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		indexPath := idx.MakeIndexPath(realPath)
		info, ok := idx.GetReducedMetadata(indexPath, isDir)
		if !ok {
			info, err = idx.GetFsDirInfo(indexPath)
			if err != nil {
				return 0, fmt.Errorf("failed to get file info for %s : %v", path, err)
			}
		}
		estimatedSize += info.Size
	}
	return estimatedSize, nil
}

func createZip(d *requestContext, tmpDirPath string, filenames ...string) error {
	file, err := os.Create(tmpDirPath)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	for _, fname := range filenames {
		err := addFile(fname, d, nil, zipWriter, false)
		if err != nil {
			logger.Errorf("Failed to add %s to ZIP: %v", fname, err)
			return err
		}
	}

	return nil
}

func createTarGz(d *requestContext, tmpDirPath string, filenames ...string) error {
	file, err := os.Create(tmpDirPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	for _, fname := range filenames {
		err := addFile(fname, d, tarWriter, nil, false)
		if err != nil {
			logger.Errorf("Failed to add %s to TAR.GZ: %v", fname, err)
			return err
		}
	}

	return nil
}

func isOnlyOfficeCompatibleFile(fileName string) bool {
	return iteminfo.IsOnlyOffice(fileName)
}
