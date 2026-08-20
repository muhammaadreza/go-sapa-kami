package utils

import (
	"os"
	"io"
	"fmt"
	"time"
	"errors"
	"strings"
	"net/http"
	"path/filepath"
	"mime/multipart"

	"go-sapa-kami/internal/model"
)

type Validate struct {
	AllowedTypes []string
	MaxFiles     int
	MaxSize      int64
	MaxTotalSize int64
}

// Validate File
func ValidateFile(request *multipart.FileHeader, validate Validate) error {
	if request == nil { return errors.New("invalid file") }
	if validate.MaxSize != 0 && request.Size > validate.MaxSize {
		return errors.New("file size exceeds the maximum limit")
	}

	if len(validate.AllowedTypes) > 0 {
		file, err := request.Open()
		if err != nil { return err }
		defer file.Close()

		mimeType, err := getMimeType(file)
		if err != nil { return err }

		if !isAllowedType(mimeType, validate.AllowedTypes) {
			return errors.New("invalid file type: " + mimeType)
		}
	}

	return nil
}

// Validate Files
func ValidateFiles(requests []*multipart.FileHeader, validate Validate) error {
	if validate.MaxFiles > 0 && len(requests) > validate.MaxFiles {
		return errors.New("number of files exceeds the maximum limit")
	}

	var totalSize int64
	for _, request := range requests {
		if request == nil { return errors.New("invalid file") }

		totalSize += request.Size
		if validate.MaxTotalSize > 0 && totalSize > validate.MaxTotalSize {
			return errors.New("total file size exceeds the maximum limit")
		}

		err := ValidateFile(request, validate)
		if err != nil { return err }
	}

	return nil
}

// Save File
func SaveFile(request *multipart.FileHeader, directory string) (*model.Attachment, error) {
	if request == nil { return nil, errors.New("invalid file") }

	safeDir, err := safePath(directory)
	if err != nil { return nil, err }

	err = os.MkdirAll(safeDir, 0755)
	if err != nil { return nil, err }

	file, err := request.Open()
	if err != nil { return nil, err }
	defer file.Close()

	fileType, err := getMimeType(file)
	if err != nil { return nil, err }

	_, err = file.Seek(0, io.SeekStart)
	if err != nil { return nil, err }

	safeName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(request.Filename))
	filePath := filepath.Join(safeDir, safeName)

	out, err := os.Create(filePath)
	if err != nil { return nil, err }

	_, err = io.Copy(out, file)
	if err != nil {
		_ = out.Close()
		_ = os.Remove(filePath)
		return nil, err
	}

	_ = out.Close()
	return &model.Attachment{
		Path: filePath,
		Size: request.Size,
		Name: request.Filename,
		Type: fileType,
	}, nil
}

// Save Files
func SaveFiles(requests []*multipart.FileHeader, directory string) ([]model.Attachment, error) {
	var files []model.Attachment
	var paths []string

	for _, request := range requests {
		savedFile, err := SaveFile(request, directory)
		if err != nil {
			cleanupErr := DeleteFiles(paths);
			if cleanupErr != nil { return nil, errors.Join(err, fmt.Errorf("cleanup saved files: %w", cleanupErr)) }
			return nil, err
		}

		files = append(files, *savedFile)
		paths = append(paths, savedFile.Path)
	}

	return files, nil
}

// Delete File
func DeleteFile(path string) error {
	cleanPath, err := safePath(path)
	if err != nil { return err }

	err = os.Remove(cleanPath)
	if errors.Is(err, os.ErrNotExist) { return nil }
	return err
}

// Delete Files
func DeleteFiles(paths []string) error {
	var deleteErrors []error
	for _, path := range paths {
		err := DeleteFile(path);
		if err != nil { deleteErrors = append(deleteErrors, fmt.Errorf("delete %q: %w", path, err)) }
	}

	return errors.Join(deleteErrors...)
}

// Helper Functions
// Get MIME Type
func getMimeType(r io.Reader) (string, error) {
	buffer := make([]byte, 512)
	bytesRead, err := r.Read(buffer)
	if err != nil && err != io.EOF { return "", err }

	return http.DetectContentType(buffer[:bytesRead]), nil
}

// Check Allowed Type
func isAllowedType(fileType string, allowedTypes []string) bool {
	for _, allowed := range allowedTypes {
		if fileType == allowed { return true }
	}

	return false
}

// Build Safe Path
func safePath(path string) (string, error) {
	if path == "" { return "", errors.New("path cannot be empty") }

	cleanPath := filepath.Clean(path)
	if !strings.HasPrefix(cleanPath, "uploads/") && cleanPath != "uploads" {
		cleanPath = filepath.Join("uploads", path)
	}

	if !strings.HasPrefix(cleanPath, "uploads"+string(filepath.Separator)) {
		return "", errors.New("invalid path: attempts to escape base directory")
	}

	return cleanPath, nil
}
