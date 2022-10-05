package file

import (
	"fmt"
	"os"
)

func checkPermission(src string) bool {
	_, err := os.Stat(src)
	return os.IsPermission(err)
}

func checkNotExist(src string) bool {
	_, err := os.Stat(src)
	return os.IsNotExist(err)
}

func isNotExistMkDir(src string) error {
	if !checkNotExist(src) {
		return nil
	}
	return os.MkdirAll(src, os.ModePerm)
}

func MustOpen(filename, dir string) (*os.File, error) {
	if checkPermission(dir) {
		return nil, fmt.Errorf("permission denied idr:%s", dir)
	}

	err := isNotExistMkDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error during make dir %s, err:%v", dir, err)
	}

	filePath := fmt.Sprintf("%s%s%s", dir, string(os.PathSeparator), filename)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}
