package resources

import (
	"path/filepath"
)

func ReplaceRelativePathBase(path string, fileName string) string {
	path = NormalizeRelativePathForStorage(path)
	fileName = NormalizeRelativePathForStorage(fileName)
	if path == "" {
		return fileName
	}

	dir := NormalizeRelativePathForStorage(filepath.ToSlash(filepath.Dir(path)))
	if dir == "" {
		return fileName
	}
	return dir + "/" + fileName
}
