package local

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/maximka-44js/lc/internal/storage/repo"
)

const (
	multipartDirName = ".multipart"
	partFilePrefix   = "part-"
)

type Config struct {
	RootDir   string
	PublicURL string
	SignKey   string
	URLTTL    time.Duration
}

type Storage struct {
	rootDir   string
	publicURL string
	signKey   []byte
	urlTTL    time.Duration
}

type LocalObjectTokenClaims struct {
	Op         string
	Key        string
	UploadID   string
	PartNumber int
	Checksum   string
	ExpiresAt  time.Time
}

type multipartMeta struct {
	Key       string    `json:"key"`
	MimeType  string    `json:"mime_type"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

type objectMeta struct {
	ContentType string    `json:"content_type"`
	Checksum    string    `json:"checksum"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewStorage(_ context.Context, cfg Config) (*Storage, error) {
	root := strings.TrimSpace(cfg.RootDir)
	if root == "" {
		return nil, fmt.Errorf("local storage: root dir is required")
	}
	absRoot, err := resolveRootDir(root)
	if err != nil {
		return nil, fmt.Errorf("local storage: resolve root dir: %w", err)
	}
	if err := os.MkdirAll(absRoot, 0o755); err != nil {
		return nil, fmt.Errorf("local storage: create root dir: %w", err)
	}

	publicURL := strings.TrimSpace(cfg.PublicURL)
	if publicURL == "" {
		return nil, fmt.Errorf("local storage: public url is required")
	}
	if _, err := url.ParseRequestURI(publicURL); err != nil {
		return nil, fmt.Errorf("local storage: invalid public url: %w", err)
	}

	signKey := strings.TrimSpace(cfg.SignKey)
	if signKey == "" {
		return nil, fmt.Errorf("local storage: sign key is required")
	}

	ttl := cfg.URLTTL
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}

	return &Storage{
		rootDir:   absRoot,
		publicURL: strings.TrimRight(publicURL, "/"),
		signKey:   []byte(signKey),
		urlTTL:    ttl,
	}, nil
}

func resolveRootDir(root string) (string, error) {
	if filepath.IsAbs(root) {
		return filepath.Clean(root), nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	base := findProjectRoot(cwd)
	return filepath.Abs(filepath.Join(base, root))
}

func findProjectRoot(start string) string {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return start
		}
		dir = parent
	}
}

func (s *Storage) PutObject(ctx context.Context, req repo.PutObjectRequest) error {
	if strings.TrimSpace(req.Key) == "" {
		return fmt.Errorf("local storage: key is required")
	}
	if req.Body == nil {
		return fmt.Errorf("local storage: body is required")
	}
	return s.WriteObject(ctx, req.Key, req.MimeType, req.Body)
}

func (s *Storage) Copy(ctx context.Context, sourceKey, destinationKey, mimeType string) error {
	sourceKey = strings.TrimSpace(sourceKey)
	destinationKey = strings.TrimSpace(destinationKey)
	if sourceKey == "" || destinationKey == "" {
		return fmt.Errorf("local storage: source and destination keys are required")
	}

	srcPath := s.objectPath(sourceKey)
	src, err := os.Open(srcPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return wrapStorageErr(repo.ErrStorageNotFound, err)
		}
		return fmt.Errorf("local storage: open source object: %w", err)
	}
	defer func() {
		_ = src.Close()
	}()

	dstPath := s.objectPath(destinationKey)
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return fmt.Errorf("local storage: mkdir destination dir: %w", err)
	}

	tmpPath := dstPath + ".tmp"
	dst, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("local storage: create destination object: %w", err)
	}

	hasher := sha256.New()
	mw := io.MultiWriter(dst, hasher)
	if _, err := io.Copy(mw, readerWithContext(ctx, src)); err != nil {
		_ = dst.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("local storage: copy object: %w", err)
	}
	if err := dst.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("local storage: close destination object: %w", err)
	}
	if err := os.Rename(tmpPath, dstPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("local storage: finalize destination object: %w", err)
	}

	meta, _ := s.readObjectMeta(sourceKey)
	contentType := strings.TrimSpace(mimeType)
	if contentType == "" {
		contentType = meta.ContentType
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	return s.writeObjectMeta(destinationKey, objectMeta{
		ContentType: contentType,
		Checksum:    checksum,
		UpdatedAt:   time.Now().UTC(),
	})
}

func (s *Storage) PresignPut(_ context.Context, req repo.PresignPutRequest) (repo.PresignPutResult, error) {
	if strings.TrimSpace(req.Key) == "" {
		return repo.PresignPutResult{}, fmt.Errorf("local storage: key is required")
	}
	claims := LocalObjectTokenClaims{Op: "put", Key: req.Key, ExpiresAt: time.Now().UTC().Add(s.urlTTL)}
	token, err := s.signToken(claims)
	if err != nil {
		return repo.PresignPutResult{}, err
	}
	return repo.PresignPutResult{URL: s.objectURL(token)}, nil
}

func (s *Storage) CreateMultipart(_ context.Context, req repo.CreateMultipartRequest) (repo.CreateMultipartResult, error) {
	if strings.TrimSpace(req.Key) == "" {
		return repo.CreateMultipartResult{}, fmt.Errorf("local storage: key is required")
	}
	uploadID := uuid.NewString()
	dir := s.multipartPath(uploadID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return repo.CreateMultipartResult{}, fmt.Errorf("local storage: create multipart dir: %w", err)
	}
	meta := multipartMeta{Key: req.Key, MimeType: req.MimeType, Size: req.Size, CreatedAt: time.Now().UTC()}
	if err := s.writeJSON(filepath.Join(dir, "meta.json"), meta); err != nil {
		return repo.CreateMultipartResult{}, err
	}
	return repo.CreateMultipartResult{UploadID: uploadID}, nil
}

func (s *Storage) PresignUploadPart(_ context.Context, req repo.PresignUploadPartRequest) (repo.PresignUploadPartResult, error) {
	if strings.TrimSpace(req.Key) == "" || strings.TrimSpace(req.UploadID) == "" || req.PartNumber <= 0 {
		return repo.PresignUploadPartResult{}, fmt.Errorf("local storage: key, upload_id and part_number are required")
	}
	meta, err := s.readMultipartMeta(req.UploadID)
	if err != nil {
		return repo.PresignUploadPartResult{}, err
	}
	if meta.Key != req.Key {
		return repo.PresignUploadPartResult{}, fmt.Errorf("local storage: upload key mismatch")
	}
	claims := LocalObjectTokenClaims{
		Op:         "part",
		Key:        req.Key,
		UploadID:   req.UploadID,
		PartNumber: req.PartNumber,
		Checksum:   strings.TrimSpace(req.ChecksumSHA256),
		ExpiresAt:  time.Now().UTC().Add(s.urlTTL),
	}
	token, err := s.signToken(claims)
	if err != nil {
		return repo.PresignUploadPartResult{}, err
	}
	partURL := s.objectPartURL(token)
	return repo.PresignUploadPartResult{URL: partURL}, nil
}

func (s *Storage) CompleteMultipart(ctx context.Context, req repo.CompleteMultipartRequest) error {
	if strings.TrimSpace(req.Key) == "" || strings.TrimSpace(req.UploadID) == "" {
		return fmt.Errorf("local storage: key and upload_id are required")
	}
	meta, err := s.readMultipartMeta(req.UploadID)
	if err != nil {
		return err
	}
	if meta.Key != req.Key {
		return fmt.Errorf("local storage: upload key mismatch")
	}

	parts := make([]repo.CompletedPart, 0, len(req.CompletedParts))
	parts = append(parts, req.CompletedParts...)
	sort.Slice(parts, func(i, j int) bool { return parts[i].PartNumber < parts[j].PartNumber })

	target := s.objectPath(req.Key)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("local storage: mkdir target dir: %w", err)
	}

	tmpTarget := target + ".tmp"
	out, err := os.OpenFile(tmpTarget, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("local storage: open target: %w", err)
	}

	hasher := sha256.New()
	mw := io.MultiWriter(out, hasher)
	copyErr := func() error {
		for _, part := range parts {
			if err := ctx.Err(); err != nil {
				return err
			}
			partPath := filepath.Join(s.multipartPath(req.UploadID), partFilename(part.PartNumber))
			f, err := os.Open(partPath)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return wrapStorageErr(repo.ErrStorageNotFound, fmt.Errorf("local storage: part missing: %s", partPath))
				}
				return fmt.Errorf("local storage: open part: %w", err)
			}
			if _, err := io.Copy(mw, f); err != nil {
				_ = f.Close()
				return fmt.Errorf("local storage: copy part: %w", err)
			}
			_ = f.Close()
		}
		return nil
	}()
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(tmpTarget)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmpTarget)
		return fmt.Errorf("local storage: close target: %w", closeErr)
	}
	if err := os.Rename(tmpTarget, target); err != nil {
		_ = os.Remove(tmpTarget)
		return fmt.Errorf("local storage: finalize target: %w", err)
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	if err := s.writeObjectMeta(req.Key, objectMeta{
		ContentType: meta.MimeType,
		Checksum:    checksum,
		UpdatedAt:   time.Now().UTC(),
	}); err != nil {
		return err
	}

	if err := os.RemoveAll(s.multipartPath(req.UploadID)); err != nil {
		return fmt.Errorf("local storage: cleanup multipart dir: %w", err)
	}

	return nil
}

func (s *Storage) AbortMultipart(_ context.Context, _ string, uploadID string) error {
	if strings.TrimSpace(uploadID) == "" {
		return fmt.Errorf("local storage: upload_id is required")
	}
	if err := os.RemoveAll(s.multipartPath(uploadID)); err != nil {
		return fmt.Errorf("local storage: abort multipart: %w", err)
	}
	return nil
}

func (s *Storage) ListMultipartUploads(_ context.Context, req repo.ListMultipartUploadsRequest) ([]repo.MultipartUpload, error) {
	base := s.multipartBasePath()
	entries, err := os.ReadDir(base)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("local storage: list multipart base dir: %w", err)
	}

	result := make([]repo.MultipartUpload, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		uploadID := entry.Name()
		meta, err := s.readMultipartMeta(uploadID)
		if err != nil {
			continue
		}
		if strings.TrimSpace(req.Prefix) != "" && !strings.HasPrefix(meta.Key, req.Prefix) {
			continue
		}
		if !meta.CreatedAt.Before(req.OlderThan) {
			continue
		}
		result = append(result, repo.MultipartUpload{
			Key:         meta.Key,
			UploadID:    uploadID,
			InitiatedAt: meta.CreatedAt,
		})
	}

	return result, nil
}

func (s *Storage) Head(_ context.Context, key string) (repo.ObjectHead, error) {
	fullPath := s.objectPath(key)
	st, err := os.Stat(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return repo.ObjectHead{}, wrapStorageErr(repo.ErrStorageNotFound, err)
		}
		return repo.ObjectHead{}, fmt.Errorf("local storage: head: %w", err)
	}

	meta, _ := s.readObjectMeta(key)

	return repo.ObjectHead{
		Size:         st.Size(),
		ETag:         "",
		LastModified: st.ModTime().UTC(),
		ContentType:  meta.ContentType,
		Checksum:     meta.Checksum,
	}, nil
}

func (s *Storage) PresignGet(_ context.Context, key string) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", fmt.Errorf("local storage: key is required")
	}
	claims := LocalObjectTokenClaims{Op: "get", Key: key, ExpiresAt: time.Now().UTC().Add(s.urlTTL)}
	token, err := s.signToken(claims)
	if err != nil {
		return "", err
	}
	return s.objectURL(token), nil
}

func (s *Storage) Delete(_ context.Context, key string) error {
	fullPath := s.objectPath(key)
	if err := os.Remove(fullPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return wrapStorageErr(repo.ErrStorageNotFound, err)
		}
		return fmt.Errorf("local storage: delete: %w", err)
	}
	_ = os.Remove(s.objectMetaPath(key))
	return nil
}

func (s *Storage) DeleteByPrefix(_ context.Context, prefix string) error {
	if strings.TrimSpace(prefix) == "" {
		return nil
	}

	cleanPrefix := filepath.FromSlash(strings.TrimPrefix(prefix, "/"))
	basePrefixPath := filepath.Join(s.rootDir, cleanPrefix)
	if err := os.RemoveAll(basePrefixPath); err != nil {
		return fmt.Errorf("local storage: delete prefix: %w", err)
	}

	// Also cleanup metadata files for deleted objects.
	err := filepath.WalkDir(s.rootDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".meta.json") {
			return nil
		}
		rel, err := filepath.Rel(s.rootDir, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		rel = strings.TrimSuffix(rel, ".meta.json")
		if strings.HasPrefix(rel, strings.TrimPrefix(prefix, "/")) {
			_ = os.Remove(path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("local storage: cleanup meta by prefix: %w", err)
	}

	return nil
}

func (s *Storage) WriteObject(ctx context.Context, key, contentType string, body io.Reader) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("local storage: key is required")
	}
	fullPath := s.objectPath(key)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return fmt.Errorf("local storage: mkdir object dir: %w", err)
	}

	tmpPath := fullPath + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("local storage: create object file: %w", err)
	}

	hasher := sha256.New()
	mw := io.MultiWriter(f, hasher)
	if _, err := io.Copy(mw, readerWithContext(ctx, body)); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("local storage: write object body: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("local storage: close object file: %w", err)
	}
	if err := os.Rename(tmpPath, fullPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("local storage: move object file: %w", err)
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	return s.writeObjectMeta(key, objectMeta{
		ContentType: contentType,
		Checksum:    checksum,
		UpdatedAt:   time.Now().UTC(),
	})
}

func (s *Storage) ValidatePutToken(token string) (LocalObjectTokenClaims, error) {
	claims, err := s.verifyToken(token)
	if err != nil {
		return LocalObjectTokenClaims{}, err
	}
	if claims.Op != "put" {
		return LocalObjectTokenClaims{}, fmt.Errorf("local storage: token op mismatch")
	}
	if strings.TrimSpace(claims.Key) == "" {
		return LocalObjectTokenClaims{}, fmt.Errorf("local storage: token key is empty")
	}
	return claims, nil
}

func (s *Storage) ValidateGetToken(token string) (LocalObjectTokenClaims, error) {
	claims, err := s.verifyToken(token)
	if err != nil {
		return LocalObjectTokenClaims{}, err
	}
	if claims.Op != "get" {
		return LocalObjectTokenClaims{}, fmt.Errorf("local storage: token op mismatch")
	}
	if strings.TrimSpace(claims.Key) == "" {
		return LocalObjectTokenClaims{}, fmt.Errorf("local storage: token key is empty")
	}
	return claims, nil
}

func (s *Storage) ValidatePartToken(token string) (LocalObjectTokenClaims, error) {
	claims, err := s.verifyToken(token)
	if err != nil {
		return LocalObjectTokenClaims{}, err
	}
	if claims.Op != "part" {
		return LocalObjectTokenClaims{}, fmt.Errorf("local storage: token op mismatch")
	}
	if strings.TrimSpace(claims.Key) == "" || strings.TrimSpace(claims.UploadID) == "" || claims.PartNumber <= 0 {
		return LocalObjectTokenClaims{}, fmt.Errorf("local storage: token fields are invalid")
	}
	if strings.TrimSpace(claims.Checksum) == "" {
		return LocalObjectTokenClaims{}, fmt.Errorf("local storage: token checksum is empty")
	}
	return claims, nil
}

func (s *Storage) WriteMultipartPart(ctx context.Context, key, uploadID string, partNumber int, expectedChecksum string, body io.Reader) error {
	if strings.TrimSpace(key) == "" || strings.TrimSpace(uploadID) == "" || partNumber <= 0 {
		return fmt.Errorf("local storage: key, upload_id and part_number are required")
	}

	meta, err := s.readMultipartMeta(uploadID)
	if err != nil {
		return err
	}
	if meta.Key != key {
		return fmt.Errorf("local storage: upload key mismatch")
	}

	partPath := filepath.Join(s.multipartPath(uploadID), partFilename(partNumber))
	f, err := os.OpenFile(partPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("local storage: create multipart part: %w", err)
	}

	hasher := sha256.New()
	mw := io.MultiWriter(f, hasher)
	if _, err := io.Copy(mw, readerWithContext(ctx, body)); err != nil {
		_ = f.Close()
		_ = os.Remove(partPath)
		return fmt.Errorf("local storage: write multipart part: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(partPath)
		return fmt.Errorf("local storage: close multipart part: %w", err)
	}

	actual := hex.EncodeToString(hasher.Sum(nil))
	if expected := strings.TrimSpace(expectedChecksum); expected != "" {
		if !strings.EqualFold(actual, expected) {
			_ = os.Remove(partPath)
			return fmt.Errorf("local storage: part checksum mismatch: expected %s actual %s", expected, actual)
		}
	}

	return nil
}

func (s *Storage) OpenObject(key string) (*os.File, string, error) {
	fullPath := s.objectPath(key)
	f, err := os.Open(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, "", wrapStorageErr(repo.ErrStorageNotFound, err)
		}
		return nil, "", fmt.Errorf("local storage: open object: %w", err)
	}
	meta, _ := s.readObjectMeta(key)
	return f, meta.ContentType, nil
}

func (s *Storage) GetObject(_ context.Context, key string) (io.ReadCloser, error) {
	f, _, err := s.OpenObject(key)
	return f, err
}

func (s *Storage) objectURL(token string) string {
	v := url.Values{}
	v.Set("token", token)
	return s.publicURL + "/v1/storage/object?" + v.Encode()
}

func (s *Storage) objectPartURL(token string) string {
	v := url.Values{}
	v.Set("token", token)
	return s.publicURL + "/v1/storage/object/part?" + v.Encode()
}

func (s *Storage) signToken(claims LocalObjectTokenClaims) (string, error) {
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("local storage: marshal token payload: %w", err)
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, s.signKey)
	_, _ = mac.Write([]byte(payloadB64))
	sig := mac.Sum(nil)
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)
	return payloadB64 + "." + sigB64, nil
}

func (s *Storage) verifyToken(token string) (LocalObjectTokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return LocalObjectTokenClaims{}, fmt.Errorf("local storage: invalid token format")
	}
	payloadB64 := parts[0]
	sigB64 := parts[1]

	sig, err := base64.RawURLEncoding.DecodeString(sigB64)
	if err != nil {
		return LocalObjectTokenClaims{}, fmt.Errorf("local storage: invalid token signature encoding")
	}
	mac := hmac.New(sha256.New, s.signKey)
	_, _ = mac.Write([]byte(payloadB64))
	expectedSig := mac.Sum(nil)
	if !hmac.Equal(sig, expectedSig) {
		return LocalObjectTokenClaims{}, fmt.Errorf("local storage: invalid token signature")
	}

	payload, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return LocalObjectTokenClaims{}, fmt.Errorf("local storage: invalid token payload encoding")
	}

	var claims LocalObjectTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return LocalObjectTokenClaims{}, fmt.Errorf("local storage: invalid token payload")
	}
	if claims.ExpiresAt.IsZero() || time.Now().UTC().After(claims.ExpiresAt) {
		return LocalObjectTokenClaims{}, fmt.Errorf("local storage: token expired")
	}

	return claims, nil
}

func (s *Storage) objectPath(key string) string {
	cleanKey := filepath.FromSlash(strings.TrimPrefix(strings.TrimSpace(key), "/"))
	return filepath.Join(s.rootDir, cleanKey)
}

func (s *Storage) objectMetaPath(key string) string {
	return s.objectPath(key) + ".meta.json"
}

func (s *Storage) writeObjectMeta(key string, meta objectMeta) error {
	return s.writeJSON(s.objectMetaPath(key), meta)
}

func (s *Storage) readObjectMeta(key string) (objectMeta, error) {
	var meta objectMeta
	err := s.readJSON(s.objectMetaPath(key), &meta)
	return meta, err
}

func (s *Storage) multipartBasePath() string {
	return filepath.Join(s.rootDir, multipartDirName)
}

func (s *Storage) multipartPath(uploadID string) string {
	return filepath.Join(s.multipartBasePath(), uploadID)
}

func (s *Storage) readMultipartMeta(uploadID string) (multipartMeta, error) {
	var meta multipartMeta
	err := s.readJSON(filepath.Join(s.multipartPath(uploadID), "meta.json"), &meta)
	if err != nil {
		return multipartMeta{}, err
	}
	return meta, nil
}

func (s *Storage) writeJSON(filePath string, v any) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return fmt.Errorf("local storage: mkdir json dir: %w", err)
	}
	payload, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("local storage: marshal json: %w", err)
	}
	if err := os.WriteFile(filePath, payload, 0o644); err != nil {
		return fmt.Errorf("local storage: write json: %w", err)
	}
	return nil
}

func (s *Storage) readJSON(filePath string, v any) error {
	payload, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return wrapStorageErr(repo.ErrStorageNotFound, err)
		}
		return fmt.Errorf("local storage: read json: %w", err)
	}
	if err := json.Unmarshal(payload, v); err != nil {
		return fmt.Errorf("local storage: unmarshal json: %w", err)
	}
	return nil
}

func partFilename(partNumber int) string {
	return partFilePrefix + strconv.Itoa(partNumber)
}

func wrapStorageErr(kind error, err error) error {
	if err == nil {
		return kind
	}
	return fmt.Errorf("%w: %v", kind, err)
}

type ctxReader struct {
	ctx context.Context
	r   io.Reader
}

func readerWithContext(ctx context.Context, r io.Reader) io.Reader {
	return &ctxReader{ctx: ctx, r: r}
}

func (r *ctxReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}
