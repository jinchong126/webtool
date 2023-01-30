package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func CopyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	// copy source file permission
	destination, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, sourceFileStat.Mode())
	if err != nil {
		return 0, err
	}

	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func RelativePath(oldDirectory, oldFilePath, newDirectory, newExt string) string {
	fileName := filepath.Base(oldFilePath)

	fileName = fileName[:len(fileName)-len(filepath.Ext(fileName))] + newExt
	fileDir := filepath.Dir(oldFilePath)

	newFilePath := ""

	if oldDirectory != "" { // 当初是通过 LoadFromDirectory 加载的
		if strings.HasPrefix(fileDir, oldDirectory) {
			if fileDir == oldDirectory { // 没有多余的路径
				newFilePath = filepath.Join(newDirectory, fileName)
			} else {
				newFilePath = filepath.Join(newDirectory, fileDir[len(oldDirectory)+1:], fileName)
			}
		} else {
			panic(fmt.Sprintf("%s don't have the prefix %s, this could not happened", oldFilePath, oldDirectory))
		}
	} else {
		newFilePath = filepath.Join(fileDir, fileName)
	}

	newFilePath = filepath.Clean(newFilePath)

	return newFilePath
}
