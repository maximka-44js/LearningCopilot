package folder

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"mime"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	storageentity "github.com/maximka-44js/lc/internal/storage/entity"
	storagerepo "github.com/maximka-44js/lc/internal/storage/repo"
)

type FolderReader interface {
	GetByID(ctx context.Context, userID, folderID string) (storageentity.Folder, error)
	ListByLibraryAndParent(ctx context.Context, userID, libID, parentID string) ([]storageentity.Folder, error)
	Store(ctx context.Context, folder *storageentity.Folder) error
}

type LibraryReader interface {
	GetByID(ctx context.Context, userID, libID string) (storageentity.Library, error)
}

type FileReaderWriter interface {
	ListByLibraryAndFolder(ctx context.Context, userID, libID, folderID string) ([]storageentity.File, error)
	Store(ctx context.Context, file *storageentity.File) error
}

type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type ExtractUseCase interface {
	ExtractFolder(ctx context.Context, userID, folderID string) error
}

type extractUseCase struct {
	folders FolderReader
	libs    LibraryReader
	files   FileReaderWriter
	storage storagerepo.ObjectStorage
	reader  ObjectReader
	tx      TxManager
}

func NewExtractUseCase(
	libs LibraryReader,
	folders FolderReader,
	files FileReaderWriter,
	storage storagerepo.ObjectStorage,
	reader ObjectReader,
	tx TxManager,
) ExtractUseCase {
	return &extractUseCase{
		folders: folders,
		libs:    libs,
		files:   files,
		storage: storage,
		reader:  reader,
		tx:      tx,
	}
}

func (uc *extractUseCase) ExtractFolder(ctx context.Context, userID, folderID string) error {
	if uc == nil {
		return fmt.Errorf("folder extract usecase: usecase is nil")
	}
	if uc.libs == nil || uc.folders == nil || uc.files == nil || uc.tx == nil {
		return fmt.Errorf("folder extract usecase: repositories are not configured")
	}
	if uc.storage == nil || uc.reader == nil {
		return fmt.Errorf("folder extract usecase: storage is not configured")
	}

	userID = strings.TrimSpace(userID)
	folderID = strings.TrimSpace(folderID)
	if userID == "" || folderID == "" {
		return fmt.Errorf("folder extract usecase: user_id and folder_id are required")
	}

	targetFolder, err := uc.folders.GetByID(ctx, userID, folderID)
	if err != nil {
		return err
	}

	lib, err := uc.libs.GetByID(ctx, userID, targetFolder.LibID)
	if err != nil {
		return err
	}
	if !lib.CanRead(userID) || !lib.CanWrite(userID) {
		return fmt.Errorf("folder extract usecase: access denied")
	}

	candidates, err := uc.files.ListByLibraryAndFolder(ctx, userID, targetFolder.LibID, targetFolder.ID)
	if err != nil {
		return err
	}

	zipFiles := make([]storageentity.File, 0)
	for _, f := range candidates {
		if isZipFile(f) {
			zipFiles = append(zipFiles, f)
		}
	}
	if len(zipFiles) == 0 {
		return fmt.Errorf("folder extract usecase: no zip files found in folder")
	}

	for _, zipFile := range zipFiles {
		if err := uc.extractArchiveIntoFolder(ctx, userID, targetFolder, zipFile); err != nil {
			return err
		}
	}

	return nil
}

func (uc *extractUseCase) extractArchiveIntoFolder(ctx context.Context, userID string, targetFolder storageentity.Folder, archiveFile storageentity.File) error {
	key := strings.TrimSpace(archiveFile.ObjectKey)
	if key == "" {
		return fmt.Errorf("folder extract usecase: zip file object key is empty")
	}

	reader, err := uc.reader.Open(ctx, key)
	if err != nil {
		return fmt.Errorf("folder extract usecase: open zip object: %w", err)
	}
	defer func() {
		_ = reader.Close()
	}()

	size := archiveFile.Size
	if size <= 0 {
		off, err := reader.Seek(0, io.SeekEnd)
		if err != nil {
			return fmt.Errorf("folder extract usecase: get zip size: %w", err)
		}
		size = off
		if _, err := reader.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("folder extract usecase: rewind zip: %w", err)
		}
	}

	zr, err := zip.NewReader(reader, size)
	if err != nil {
		return fmt.Errorf("folder extract usecase: open zip reader: %w", err)
	}

	rootName := archiveRootName(archiveFile.Name)
	rootFolder, err := uc.createUniqueChildFolder(ctx, userID, targetFolder.LibID, targetFolder.ID, rootName)
	if err != nil {
		return err
	}

	dirCache := map[string]string{"": rootFolder.ID}
	fileNameCache := make(map[string]map[string]struct{})

	for _, zf := range zr.File {
		if err := ctx.Err(); err != nil {
			return err
		}

		cleanPath, ok := sanitizeZipPath(zf.Name)
		if !ok {
			continue
		}

		if zf.FileInfo().IsDir() {
			if _, err := uc.ensureDirPath(ctx, userID, targetFolder.LibID, rootFolder.ID, dirCache, cleanPath); err != nil {
				return err
			}
			continue
		}

		dirPath, fileName := path.Split(cleanPath)
		fileName = strings.TrimSpace(fileName)
		if fileName == "" {
			continue
		}

		parentFolderID, err := uc.ensureDirPath(ctx, userID, targetFolder.LibID, rootFolder.ID, dirCache, strings.TrimSuffix(dirPath, "/"))
		if err != nil {
			return err
		}

		mimeType, ok := detectAllowedMimeType(fileName)
		if !ok {
			continue
		}

		uniqueName, err := uc.ensureUniqueFileName(ctx, userID, targetFolder.LibID, parentFolderID, fileName, fileNameCache)
		if err != nil {
			return err
		}

		entryReader, err := zf.Open()
		if err != nil {
			return fmt.Errorf("folder extract usecase: open zip entry: %w", err)
		}

		fileID := uuid.NewString()
		objectKey := buildExtractedObjectKey(userID, targetFolder.LibID, parentFolderID, fileID, uniqueName)
		contentSize := int64(zf.UncompressedSize64)

		putErr := uc.storage.PutObject(ctx, storagerepo.PutObjectRequest{
			Key:      objectKey,
			MimeType: mimeType,
			Size:     contentSize,
			Body:     entryReader,
		})
		closeErr := entryReader.Close()
		if putErr != nil {
			return fmt.Errorf("folder extract usecase: store extracted object: %w", putErr)
		}
		if closeErr != nil {
			return fmt.Errorf("folder extract usecase: close zip entry reader: %w", closeErr)
		}

		head, err := uc.storage.Head(ctx, objectKey)
		if err != nil {
			_ = uc.storage.Delete(context.Background(), objectKey)
			return fmt.Errorf("folder extract usecase: head extracted object: %w", err)
		}

		newFile, err := storageentity.NewFile(fileID, userID, targetFolder.LibID, parentFolderID, uniqueName, mimeType, head.Size, "")
		if err != nil {
			_ = uc.storage.Delete(context.Background(), objectKey)
			return err
		}

		if err := newFile.ValidateStatusTransition(storageentity.FileStatusUploaded); err != nil {
			_ = uc.storage.Delete(context.Background(), objectKey)
			return err
		}

		newFile.Status = storageentity.FileStatusUploaded
		newFile.ObjectKey = objectKey
		newFile.Checksum = strings.TrimSpace(head.Checksum)
		newFile.UpdatedAt = time.Now().UTC()

		if err := uc.tx.WithinTx(ctx, func(txCtx context.Context) error {
			if err := uc.files.Store(txCtx, &newFile); err != nil {
				return err
			}
			return nil
		}); err != nil {
			_ = uc.storage.Delete(context.Background(), objectKey)
			return err
		}

		if _, ok := fileNameCache[parentFolderID]; !ok {
			fileNameCache[parentFolderID] = make(map[string]struct{})
		}
		fileNameCache[parentFolderID][uniqueName] = struct{}{}
	}

	return nil
}

func (uc *extractUseCase) ensureDirPath(ctx context.Context, userID, libID, rootFolderID string, dirCache map[string]string, relativePath string) (string, error) {
	relativePath = strings.TrimSpace(strings.Trim(relativePath, "/"))
	if relativePath == "" {
		return rootFolderID, nil
	}

	parts := strings.Split(relativePath, "/")
	parentID := rootFolderID
	currentPath := ""

	for _, rawPart := range parts {
		name := strings.TrimSpace(rawPart)
		if name == "" {
			continue
		}

		if currentPath == "" {
			currentPath = name
		} else {
			currentPath = currentPath + "/" + name
		}

		if existingID, ok := dirCache[currentPath]; ok {
			parentID = existingID
			continue
		}

		folder, err := uc.ensureChildFolder(ctx, userID, libID, parentID, name)
		if err != nil {
			return "", err
		}

		dirCache[currentPath] = folder.ID
		parentID = folder.ID
	}

	return parentID, nil
}

func (uc *extractUseCase) ensureChildFolder(ctx context.Context, userID, libID, parentID, name string) (storageentity.Folder, error) {
	children, err := uc.folders.ListByLibraryAndParent(ctx, userID, libID, parentID)
	if err != nil {
		return storageentity.Folder{}, err
	}
	for _, ch := range children {
		if ch.Name == name {
			return ch, nil
		}
	}

	newFolder, err := storageentity.NewFolder(uuid.NewString(), userID, libID, name, parentID)
	if err != nil {
		return storageentity.Folder{}, err
	}

	if err := uc.tx.WithinTx(ctx, func(txCtx context.Context) error {
		return uc.folders.Store(txCtx, &newFolder)
	}); err != nil {
		return storageentity.Folder{}, err
	}

	return newFolder, nil
}

func (uc *extractUseCase) createUniqueChildFolder(ctx context.Context, userID, libID, parentID, baseName string) (storageentity.Folder, error) {
	children, err := uc.folders.ListByLibraryAndParent(ctx, userID, libID, parentID)
	if err != nil {
		return storageentity.Folder{}, err
	}

	used := make(map[string]struct{}, len(children))
	for _, ch := range children {
		used[ch.Name] = struct{}{}
	}

	name := baseName
	if name == "" {
		name = "extracted"
	}
	if _, ok := used[name]; ok {
		i := 2
		for {
			candidate := fmt.Sprintf("%s (%d)", baseName, i)
			if _, exists := used[candidate]; !exists {
				name = candidate
				break
			}
			i++
		}
	}

	newFolder, err := storageentity.NewFolder(uuid.NewString(), userID, libID, name, parentID)
	if err != nil {
		return storageentity.Folder{}, err
	}

	if err := uc.tx.WithinTx(ctx, func(txCtx context.Context) error {
		return uc.folders.Store(txCtx, &newFolder)
	}); err != nil {
		return storageentity.Folder{}, err
	}

	return newFolder, nil
}

func (uc *extractUseCase) ensureUniqueFileName(ctx context.Context, userID, libID, folderID, fileName string, cache map[string]map[string]struct{}) (string, error) {
	if _, ok := cache[folderID]; !ok {
		cache[folderID] = make(map[string]struct{})
		items, err := uc.files.ListByLibraryAndFolder(ctx, userID, libID, folderID)
		if err != nil {
			return "", err
		}
		for _, item := range items {
			cache[folderID][item.Name] = struct{}{}
		}
	}

	if _, exists := cache[folderID][fileName]; !exists {
		return fileName, nil
	}

	base := fileName
	ext := ""
	if dot := strings.LastIndex(fileName, "."); dot > 0 {
		base = fileName[:dot]
		ext = fileName[dot:]
	}

	i := 2
	for {
		candidate := fmt.Sprintf("%s (%d)%s", base, i, ext)
		if _, exists := cache[folderID][candidate]; !exists {
			return candidate, nil
		}
		i++
	}
}

func isZipFile(f storageentity.File) bool {
	if strings.EqualFold(strings.TrimSpace(f.MimeType), "application/zip") {
		return true
	}
	return strings.HasSuffix(strings.ToLower(strings.TrimSpace(f.Name)), ".zip")
}

func archiveRootName(name string) string {
	base := strings.TrimSpace(name)
	base = strings.TrimSuffix(base, ".zip")
	if base == "" {
		return "archive"
	}
	return base
}

func sanitizeZipPath(raw string) (string, bool) {
	value := strings.TrimSpace(strings.ReplaceAll(raw, "\\", "/"))
	if value == "" {
		return "", false
	}

	clean := path.Clean(value)
	if clean == "." || strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") {
		return "", false
	}

	parts := strings.Split(clean, "/")
	for _, p := range parts {
		if p == ".." {
			return "", false
		}
	}

	return clean, true
}

func buildExtractedObjectKey(userID, libID, folderID, fileID, fileName string) string {
	return "files/" + userID + "/" + libID + "/" + folderID + "/" + fileID + "/" + fileName
}

func detectAllowedMimeType(name string) (string, bool) {
	ext := strings.ToLower(path.Ext(strings.TrimSpace(name)))
	if ext == "" {
		return "", false
	}

	known := map[string]string{
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".txt":  "text/plain",
		".md":   "text/markdown",
		".csv":  "text/csv",
		".json": "application/json",
		".xml":  "application/xml",
		".zip":  "application/zip",
	}

	if mt, ok := known[ext]; ok {
		return mt, true
	}

	mt := strings.TrimSpace(mime.TypeByExtension(ext))
	if mt == "" {
		return "", false
	}
	if semi := strings.Index(mt, ";"); semi >= 0 {
		mt = strings.TrimSpace(mt[:semi])
	}

	allowed := map[string]struct{}{
		"application/pdf":    {},
		"application/msword": {},
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {},
		"application/vnd.ms-excel": {},
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         {},
		"application/vnd.ms-powerpoint":                                             {},
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": {},
		"text/plain":       {},
		"text/markdown":    {},
		"text/csv":         {},
		"application/json": {},
		"application/xml":  {},
		"application/zip":  {},
	}

	_, ok := allowed[mt]
	return mt, ok
}
