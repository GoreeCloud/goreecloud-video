package library

import "testing"

func TestVideoItemValidate(t *testing.T) {
    item := VideoItem{ID: "video-1", LibraryID: "library-1", Title: "Example", OriginalURI: "file:///media/example.mkv", MIMEType: "video/x-matroska", SizeBytes: 100}
    if err := item.Validate(); err != nil {
        t.Fatalf("expected valid item: %v", err)
    }
}

func TestVideoItemRejectsAudio(t *testing.T) {
    item := VideoItem{ID: "audio-1", LibraryID: "library-1", Title: "Example", OriginalURI: "file:///media/example.flac", MIMEType: "audio/flac", SizeBytes: 100}
    if err := item.Validate(); err == nil {
        t.Fatal("expected non-video MIME type to be rejected")
    }
}
