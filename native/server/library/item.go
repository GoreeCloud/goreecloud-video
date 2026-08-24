package library

import (
    "errors"
    "strings"
    "time"
)

type VideoItem struct {
    ID          string
    LibraryID   string
    Title       string
    OriginalURI string
    MIMEType    string
    Duration    time.Duration
    SizeBytes   int64
}

func (v VideoItem) Validate() error {
    if strings.TrimSpace(v.ID) == "" || strings.TrimSpace(v.LibraryID) == "" {
        return errors.New("id and library id are required")
    }
    if strings.TrimSpace(v.Title) == "" || strings.TrimSpace(v.OriginalURI) == "" {
        return errors.New("title and original uri are required")
    }
    if !strings.HasPrefix(v.MIMEType, "video/") {
        return errors.New("video mime type required")
    }
    if v.Duration < 0 || v.SizeBytes < 0 {
        return errors.New("duration and size must be non-negative")
    }
    return nil
}
