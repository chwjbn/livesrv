package glib

import (
	"os"
	"path"
	"path/filepath"
)

func AppVersion() string {
	sRet := "202412251430"
	return sRet
}

func AppBaseDir() string {

	sRet := ""

	modFilePath, _ := os.Executable()
	if len(modFilePath) < 1 {
		modFilePath = os.Args[0]
	}

	xFilePath, xFilePathErr := filepath.Abs(modFilePath)
	if xFilePathErr != nil {
		return sRet
	}

	sRet = filepath.Dir(xFilePath)

	return sRet
}

func AppFileName() string {

	sRet := ""

	modFilePath, _ := os.Executable()
	if len(modFilePath) < 1 {
		modFilePath = os.Args[0]
	}

	xFilePath, xFilePathErr := filepath.Abs(modFilePath)
	if xFilePathErr != nil {
		return sRet
	}

	sRet = filepath.Base(xFilePath)

	return sRet

}

func AppFullPath() string {

	sRet := ""

	modFilePath, _ := os.Executable()
	if len(modFilePath) < 1 {
		modFilePath = os.Args[0]
	}

	xFilePath, xFilePathErr := filepath.Abs(modFilePath)
	if xFilePathErr != nil {
		return sRet
	}

	sRet = xFilePath

	return sRet

}

func AppDebugMode() bool {

	bRet := false

	debugFile := path.Join(AppBaseDir(), "debug.lock")
	if FileExists(debugFile) {
		bRet = true
	}

	return bRet
}
