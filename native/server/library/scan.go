package library

import (
	"path/filepath"
	"strings"
)

var supportedVideoExtensions = map[string]struct{}{
	".avi": {}, ".m2ts": {}, ".m4v": {}, ".mkv": {}, ".mov": {}, ".mp4": {}, ".mpeg": {}, ".mpg": {}, ".ts": {}, ".webm": {},
}

type ScanCandidate struct {
	Path string
}

func NewScanCandidate(path string) (ScanCandidate, bool) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ScanCandidate{}, false
	}
	ext := strings.ToLower(filepath.Ext(trimmed))
	if _, ok := supportedVideoExtensions[ext]; !ok {
		return ScanCandidate{}, false
	}
	return ScanCandidate{Path: trimmed}, true
}

func SupportedVideoExtensions() []string {
	return []string{".avi", ".m2ts", ".m4v", ".mkv", ".mov", ".mp4", ".mpeg", ".mpg", ".ts", ".webm"}
}
