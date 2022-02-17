package atomic

import (
	"fmt"
	"os"
)

type AtomicResult struct {
	Err               error
	IsWriteSuccessful bool
	IsTempFileDeleted bool
	TempFilePath      string
}

func NewAtomicError(msg string, tempFilePath string) *AtomicResult {
	isTempFileDeleted, tempFilePath := tryDeleteTempFile(tempFilePath)

	return &AtomicResult{
		Err:               fmt.Errorf(msg),
		IsWriteSuccessful: false,
		IsTempFileDeleted: isTempFileDeleted,
		TempFilePath:      tempFilePath,
	}
}

func NewAtomicResult(tempFilePath string) *AtomicResult {
	isTempFileDeleted, tempFilePath := tryDeleteTempFile(tempFilePath)

	return &AtomicResult{
		IsWriteSuccessful: true,
		IsTempFileDeleted: isTempFileDeleted,
		TempFilePath:      tempFilePath,
	}
}

func tryDeleteTempFile(tempFilePath string) (bool, string) {
	// Temp file path not provided, nothing to delete.
	if tempFilePath == "" {
		return true, ""
	}

	// Try to remove the temp file.
	err := os.Remove(tempFilePath)
	if err == nil || os.IsNotExist(err) {
		// Deleted successfully or file doesn't exist.
		return true, ""
	}

	// Failed to delete the temp file.
	return false, tempFilePath
}
