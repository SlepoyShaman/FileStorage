package http

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/SlepoyShaman/filebrowser/backend/common/errors"
	"github.com/SlepoyShaman/filebrowser/backend/common/settings"
	"github.com/SlepoyShaman/filebrowser/backend/common/utils"
	"github.com/SlepoyShaman/filebrowser/backend/database/share"
	"github.com/SlepoyShaman/filebrowser/backend/database/users"
	"github.com/SlepoyShaman/filebrowser/backend/indexing"
	"github.com/SlepoyShaman/go-logger/logger"
)

type ShareResponse struct {
	*share.Link
	Source     string `json:"source"`
	Username   string `json:"username,omitempty"`
	PathExists bool   `json:"pathExists"`
}

func convertToFrontendShareResponse(r *http.Request, shares []*share.Link) ([]*ShareResponse, error) {
	responses := make([]*ShareResponse, 0, len(shares))
	for _, s := range shares {
		user, err := store.Users.Get(s.UserID)
		username := ""
		if err == nil {
			username = user.Username
		}

		sourceInfo, ok := config.Server.SourceMap[s.Source]
		if !ok {
			sourceInfo, ok = config.Server.NameToSource[s.Source]
			if !ok {
				logger.Error("Invalid share - deleting", "hash", s.Hash, "source", s.Source)
				_ = store.Share.Delete(s.Hash)
				continue
			}
			logger.Warning("Share has corrupted source - fixing", "hash", s.Hash, "from", s.Source, "to", sourceInfo.Path)
			s.Source = sourceInfo.Path
			_ = store.Share.Save(s)
		}

		pathExists := utils.CheckPathExists(filepath.Join(sourceInfo.Path, s.Path))

		s.CommonShare.HasPassword = s.HasPassword()
		s.DownloadURL = getShareURL(r, s.Hash, true)
		s.ShareURL = getShareURL(r, s.Hash, false)

		responses = append(responses, &ShareResponse{
			Link:       s,
			Source:     sourceInfo.Name,
			Username:   username,
			PathExists: pathExists,
		})
	}
	return responses, nil
}

func shareListHandler(w http.ResponseWriter, r *http.Request, d *requestContext) (int, error) {
	var err error
	var shares []*share.Link
	if d.user.Permissions.Admin {
		shares, err = store.Share.All()
	} else {
		shares, err = store.Share.FindByUserID(d.user.ID)
	}
	if err != nil && err != errors.ErrNotExist {
		return http.StatusInternalServerError, err
	}
	shares = utils.NonNilSlice(shares)
	sharesWithUsernames, err := convertToFrontendShareResponse(r, shares)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return renderJSON(w, r, sharesWithUsernames)
}

func shareGetHandler(w http.ResponseWriter, r *http.Request, d *requestContext) (int, error) {
	encodedPath := r.URL.Query().Get("path")
	sourceName := r.URL.Query().Get("source")
	path, err := url.PathUnescape(encodedPath)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("invalid path encoding: %v", err)
	}
	sourceInfo, ok := config.Server.NameToSource[sourceName]
	if !ok {
		return http.StatusBadRequest, fmt.Errorf("invalid source name: %s", sourceName)
	}
	userscope, err := settings.GetScopeFromSourceName(d.user.Scopes, sourceName)
	if err != nil {
		return http.StatusForbidden, err
	}
	scopePath := utils.JoinPathAsUnix(userscope, path)
	scopePath = utils.AddTrailingSlashIfNotExists(scopePath)
	s, err := store.Share.Gets(scopePath, sourceInfo.Path, d.user.ID)
	if err == errors.ErrNotExist || len(s) == 0 {
		return renderJSON(w, r, []*ShareResponse{})
	}

	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("error getting share info from server")
	}
	sharesWithUsernames, err := convertToFrontendShareResponse(r, s)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return renderJSON(w, r, sharesWithUsernames)
}

func shareDeleteHandler(w http.ResponseWriter, r *http.Request, d *requestContext) (int, error) {
	hash := r.URL.Query().Get("hash")

	if hash == "" {
		return http.StatusBadRequest, nil
	}

	err := store.Share.Delete(hash)
	if err != nil {
		return errToStatus(err), err
	}

	return errToStatus(err), err
}

func sharePatchHandler(w http.ResponseWriter, r *http.Request, d *requestContext) (int, error) {
	var body struct {
		Hash string `json:"hash"`
		Path string `json:"path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return http.StatusBadRequest, fmt.Errorf("failed to decode body: %w", err)
	}
	defer r.Body.Close()

	if body.Hash == "" || body.Path == "" {
		return http.StatusBadRequest, fmt.Errorf("hash and path are required")
	}

	err := store.Share.UpdateSharePath(body.Hash, body.Path)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	updatedShare, err := store.Share.GetByHash(body.Hash)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	sharesWithUsernames, err := convertToFrontendShareResponse(r, []*share.Link{updatedShare})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, sharesWithUsernames[0])
}

func sharePostHandler(w http.ResponseWriter, r *http.Request, d *requestContext) (int, error) {
	var s *share.Link
	var err error
	var body share.CreateBody
	if r.Body != nil {
		if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
			return http.StatusBadRequest, fmt.Errorf("failed to decode body: %w", err)
		}
		defer r.Body.Close()
	}

	if body.Hash != "" {
		s, err = store.Share.GetByHash(body.Hash)
		if err != nil {
			return http.StatusBadRequest, fmt.Errorf("invalid hash provided")
		}
	}

	var expire int64 = 0

	if body.Expires != "" {
		var num int
		num, err = strconv.Atoi(body.Expires)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		var add time.Duration
		switch body.Unit {
		case "seconds":
			add = time.Second * time.Duration(num)
		case "minutes":
			add = time.Minute * time.Duration(num)
		case "days":
			add = time.Hour * 24 * time.Duration(num)
		default:
			add = time.Hour * time.Duration(num)
		}

		expire = time.Now().Add(add).Unix()
	}

	hash, status, err := getSharePasswordHash(body)
	if err != nil {
		return status, err
	}
	stringHash := ""
	var token string
	if len(hash) > 0 {
		tokenBuffer := make([]byte, 24)
		if _, err = rand.Read(tokenBuffer); err != nil {
			return http.StatusInternalServerError, err
		}
		token = base64.URLEncoding.EncodeToString(tokenBuffer)
		stringHash = string(hash)
	}
	if s != nil {
		shouldResetCounts := s.DownloadsLimit != body.DownloadsLimit || s.PerUserDownloadLimit != body.PerUserDownloadLimit

		s.Expire = expire
		s.PasswordHash = stringHash
		s.Token = token
		body.Path = s.Path
		body.Source = s.Source
		s.CommonShare = body.CommonShare
		if s.ShareType == "upload" && !body.AllowCreate {
			s.AllowCreate = true
		}

		if shouldResetCounts {
			s.ResetDownloadCounts()
		}

		if err = store.Share.Save(s); err != nil {
			return http.StatusInternalServerError, err
		}
		var user *users.User
		user, err = store.Users.Get(s.UserID)
		username := ""
		if err == nil {
			username = user.Username
		}
		response := &ShareResponse{
			Link:     s,
			Username: username,
		}
		return renderJSON(w, r, response)
	}

	source, ok := config.Server.NameToSource[body.Source]
	if !ok {
		return http.StatusForbidden, fmt.Errorf("source with name not found: %s", body.Source)
	}

	if source.Config.Private {
		return http.StatusForbidden, fmt.Errorf("the target source is private, sharing is not permitted")
	}

	secure_hash, err := generateShortUUID()
	if err != nil {
		return http.StatusInternalServerError, err
	}
	idx := indexing.GetIndex(source.Name)
	if idx == nil {
		return http.StatusForbidden, fmt.Errorf("source with name not found: %s", body.Source)
	}
	userscope, err := settings.GetScopeFromSourceName(d.user.Scopes, source.Name)
	if err != nil {
		return http.StatusForbidden, err
	}
	scopePath := utils.JoinPathAsUnix(userscope, body.Path)
	scopePath = utils.AddTrailingSlashIfNotExists(scopePath)
	body.Path = scopePath
	_, exists := idx.GetReducedMetadata(body.Path, true)
	if !exists {
		_, exists := idx.GetReducedMetadata(utils.GetParentDirectoryPath(body.Path), true)
		if !exists {
			return http.StatusForbidden, fmt.Errorf("path not found: %s", body.Path)
		}
	}
	if body.ShareType == "upload" && !body.AllowCreate {
		body.AllowCreate = true
	}
	body.Source = source.Path
	s = &share.Link{
		Expire:       expire,
		UserID:       d.user.ID,
		Hash:         secure_hash,
		PasswordHash: stringHash,
		Token:        token,
		CommonShare:  body.CommonShare,
		Version:      1,
	}
	if err = store.Share.Save(s); err != nil {
		return http.StatusInternalServerError, err
	}
	sharesWithUsernames, err := convertToFrontendShareResponse(r, []*share.Link{s})
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return renderJSON(w, r, sharesWithUsernames[0])
}

type DirectDownloadResponse struct {
	Status      string `json:"status"`
	Hash        string `json:"hash"`
	DownloadURL string `json:"url"`
	ShareURL    string `json:"shareUrl"`
}

func shareDirectDownloadHandler(w http.ResponseWriter, r *http.Request, d *requestContext) (int, error) {
	encodedPath := r.URL.Query().Get("path")
	source := r.URL.Query().Get("source")
	duration := r.URL.Query().Get("duration")
	downloadCountStr := r.URL.Query().Get("count")
	downloadSpeedStr := r.URL.Query().Get("speed")

	if encodedPath == "" || source == "" {
		return http.StatusBadRequest, fmt.Errorf("path and source are required")
	}

	path, err := url.QueryUnescape(encodedPath)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("invalid path encoding: %v", err)
	}

	sourceInfo, ok := config.Server.NameToSource[source]
	if !ok {
		return http.StatusBadRequest, fmt.Errorf("invalid source name: %s", source)
	}

	userscope, err := settings.GetScopeFromSourceName(d.user.Scopes, source)
	if err != nil {
		return http.StatusForbidden, err
	}

	idx := indexing.GetIndex(source)
	if idx == nil {
		return http.StatusForbidden, fmt.Errorf("source with name not found: %s", source)
	}

	metadata, exists := idx.GetReducedMetadata(path, false)
	if !exists {
		return http.StatusBadRequest, fmt.Errorf("path is either not a file or not found: %s", path)
	}

	if metadata.Type == "directory" {
		return http.StatusBadRequest, fmt.Errorf("path must be a file, not a directory: %s", path)
	}

	if duration == "" {
		duration = "60"
	}

	var downloadCount int
	if downloadCountStr != "" {
		downloadCount, err = strconv.Atoi(downloadCountStr)
		if err != nil {
			return http.StatusBadRequest, fmt.Errorf("invalid downloadCount: %v", err)
		}
	}

	var downloadSpeed int
	if downloadSpeedStr != "" {
		downloadSpeed, err = strconv.Atoi(downloadSpeedStr)
		if err != nil {
			return http.StatusBadRequest, fmt.Errorf("invalid downloadSpeed: %v", err)
		}
	}

	durationNum, err := strconv.Atoi(duration)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("invalid duration: %v", err)
	}
	expire := time.Now().Add(time.Minute * time.Duration(durationNum)).Unix()

	secureHash, err := generateShortUUID()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	scopePath := utils.JoinPathAsUnix(userscope, path)

	existingShares, err := store.Share.Gets(scopePath, sourceInfo.Path, d.user.ID)
	if err == nil && len(existingShares) > 0 {
		for _, existing := range existingShares {
			if existing.DownloadsLimit == downloadCount &&
				existing.MaxBandwidth == downloadSpeed &&
				existing.QuickDownload &&
				(existing.Expire == 0 || existing.Expire >= expire) {

				response := DirectDownloadResponse{
					Status:      "201",
					Hash:        existing.Hash,
					DownloadURL: getShareURL(r, existing.Hash, true),
					ShareURL:    getShareURL(r, existing.Hash, false),
				}
				return renderJSON(w, r, response)
			}
		}
	}

	shareLink := &share.Link{
		Expire:  expire,
		UserID:  d.user.ID,
		Hash:    secureHash,
		Version: 1,
		CommonShare: share.CommonShare{
			Path:           scopePath,
			Source:         idx.Path,
			DownloadsLimit: downloadCount,
			MaxBandwidth:   downloadSpeed,
			QuickDownload:  true,
		},
	}

	if err := store.Share.Save(shareLink); err != nil {
		return http.StatusInternalServerError, err
	}

	response := DirectDownloadResponse{
		Status:      "200",
		Hash:        secureHash,
		DownloadURL: getShareURL(r, secureHash, true),
		ShareURL:    getShareURL(r, secureHash, false),
	}

	return renderJSON(w, r, response)
}

func getShareURL(r *http.Request, hash string, isDirectDownload bool) string {
	var shareURL string

	if config.Server.ExternalUrl != "" && len(config.Server.ExternalUrl) > 0 {
		basePath := config.Server.BaseURL
		if isDirectDownload == true {
			shareURL = config.Server.ExternalUrl + basePath + "public/api/raw?" + "hash=" + hash
		} else {
			if hash != "" {
				shareURL = config.Server.ExternalUrl + basePath + "public/share/" + hash
			} else {
				shareURL = config.Server.ExternalUrl + basePath + "public/share/unknown"
			}
		}
	} else {
		host := r.Host
		scheme := getScheme(r)

		forwardedHosts := r.Header.Values("X-Forwarded-Host")
		if len(forwardedHosts) > 0 {
			host = forwardedHosts[0]
			forwardedProtos := r.Header.Values("X-Forwarded-Proto")
			if len(forwardedProtos) > 0 {
				scheme = forwardedProtos[0]
			} else {
				if strings.Contains(host, "localhost") {
					scheme = "http"
				} else {
					scheme = "https"
				}
			}
		}

		if isDirectDownload {
			shareURL = scheme + "://" + host + config.Server.BaseURL + "public/api/raw?hash=" + hash
		} else {
			shareURL = scheme + "://" + host + config.Server.BaseURL + "public/share/" + hash
		}
	}

	return shareURL
}

func shareInfoHandler(w http.ResponseWriter, r *http.Request, d *requestContext) (int, error) {
	hash := r.URL.Query().Get("hash")
	commonShare, err := store.Share.GetCommonShareByHash(hash)
	if err != nil {
		return http.StatusNotFound, fmt.Errorf("share hash not found")
	}
	commonShare.DownloadURL = getShareURL(r, hash, true)
	commonShare.ShareURL = getShareURL(r, hash, false)
	return renderJSON(w, r, commonShare)
}

func getSharePasswordHash(body share.CreateBody) (data []byte, statuscode int, err error) {
	if body.Password == "" {
		return nil, 0, nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to hash password")
	}

	return hash, 0, nil
}

func generateShortUUID() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	uuid := base64.RawURLEncoding.EncodeToString(bytes)

	return uuid[:22], nil
}

func redirectToShare(w http.ResponseWriter, r *http.Request, d *requestContext) (int, error) {
	sharePath := strings.TrimPrefix(r.URL.Path, config.Server.BaseURL+"share/")
	newURL := config.Server.BaseURL + "public/share/" + sharePath
	if r.URL.RawQuery != "" {
		newURL += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, newURL, http.StatusMovedPermanently)
	return http.StatusMovedPermanently, nil
}
