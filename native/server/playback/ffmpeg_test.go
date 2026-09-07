package playback

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
)

type fakeFFmpegRunner struct {
	calls  int
	binary string
	args   []string
	err    error
}

func (f *fakeFFmpegRunner) Run(_ context.Context, binary string, args []string) error {
	f.calls++
	f.binary = binary
	f.args = append([]string(nil), args...)
	return f.err
}

func TestBuildFFmpegArgsUsesAuthorizedPathsNotOpaqueSourceID(t *testing.T) {
	request := FFmpegExecutionRequest{
		Transform: TransformRequest{
			sourceID:     "opaque-source-id-not-a-path",
			requirements: []TransformRequirement{TransformMediaTranscode},
		},
		InputPath:  "/srv/goreecloud/video/input.mkv",
		OutputPath: "/srv/goreecloud/video/work/output.mp4",
	}
	args, err := BuildFFmpegArgs(request)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, "\x00")
	if strings.Contains(joined, request.Transform.SourceID()) {
		t.Fatalf("opaque source id leaked into ffmpeg argv: %q", joined)
	}
	if !slices.Contains(args, request.InputPath) || !slices.Contains(args, request.OutputPath) {
		t.Fatalf("authorized paths missing from argv: %v", args)
	}
	if !slices.Contains(args, "libx264") || !slices.Contains(args, "aac") || !slices.Contains(args, "mp4") {
		t.Fatalf("expected bounded transcode output args: %v", args)
	}
}

func TestBuildFFmpegArgsCombinesHDRToneMapAndSafeSubtitleBurnIn(t *testing.T) {
	request := FFmpegExecutionRequest{
		Transform: TransformRequest{
			sourceID: "source-1",
			requirements: []TransformRequirement{
				TransformSubtitleBurnIn,
				TransformHDRToneMap,
			},
		},
		InputPath:    "/srv/video/input.mkv",
		OutputPath:   "/srv/video/output.mp4",
		SubtitlePath: "/srv/video/subtitles/en.srt",
	}
	args, err := BuildFFmpegArgs(request)
	if err != nil {
		t.Fatal(err)
	}
	filterIndex := slices.Index(args, "-vf")
	if filterIndex < 0 || filterIndex+1 >= len(args) {
		t.Fatalf("missing filtergraph: %v", args)
	}
	filter := args[filterIndex+1]
	if !strings.HasPrefix(filter, hdrToSDRFilter+",") || !strings.Contains(filter, "subtitles=filename='/srv/video/subtitles/en.srt'") {
		t.Fatalf("filtergraph=%q", filter)
	}
}

func TestBuildFFmpegArgsRejectsUnsafeOrUnauthorizedPathShapes(t *testing.T) {
	base := FFmpegExecutionRequest{
		Transform: TransformRequest{
			sourceID:     "source-1",
			requirements: []TransformRequirement{TransformSubtitleBurnIn},
		},
		InputPath:    "/srv/video/input.mkv",
		OutputPath:   "/srv/video/output.mp4",
		SubtitlePath: "/srv/video/subtitles/en.srt",
	}

	unsafeSubtitle := base
	unsafeSubtitle.SubtitlePath = "/srv/video/subtitles/en,forced.srt"
	if _, err := BuildFFmpegArgs(unsafeSubtitle); !errors.Is(err, ErrUnsafeSubtitleFilterPath) {
		t.Fatalf("unsafe subtitle error=%v", err)
	}

	relativeInput := base
	relativeInput.InputPath = "input.mkv"
	if _, err := BuildFFmpegArgs(relativeInput); !errors.Is(err, ErrInvalidFFmpegExecution) {
		t.Fatalf("relative input error=%v", err)
	}

	samePath := base
	samePath.OutputPath = samePath.InputPath
	if _, err := BuildFFmpegArgs(samePath); !errors.Is(err, ErrInvalidFFmpegExecution) {
		t.Fatalf("same path error=%v", err)
	}

	extraSubtitle := base
	extraSubtitle.Transform.requirements = []TransformRequirement{TransformMediaTranscode}
	if _, err := BuildFFmpegArgs(extraSubtitle); !errors.Is(err, ErrInvalidFFmpegExecution) {
		t.Fatalf("unexpected subtitle error=%v", err)
	}
}

func TestFFmpegExecutorDelegatesExactlyOnceAndPropagatesRunnerFailure(t *testing.T) {
	runner := &fakeFFmpegRunner{}
	executor, err := NewFFmpegExecutor("/usr/bin/ffmpeg", runner)
	if err != nil {
		t.Fatal(err)
	}
	request := FFmpegExecutionRequest{
		Transform:  TransformRequest{sourceID: "source-1", requirements: []TransformRequirement{TransformMediaTranscode}},
		InputPath:  "/srv/video/input.mkv",
		OutputPath: "/srv/video/output.mp4",
	}
	if err := executor.Execute(context.Background(), request); err != nil || runner.calls != 1 || runner.binary != "/usr/bin/ffmpeg" {
		t.Fatalf("calls=%d binary=%q err=%v", runner.calls, runner.binary, err)
	}

	want := errors.New("runner failed")
	runner.err = want
	if err := executor.Execute(context.Background(), request); !errors.Is(err, want) || runner.calls != 2 {
		t.Fatalf("calls=%d err=%v", runner.calls, err)
	}
}

func TestFFmpegExecutorRejectsCanceledContextAndInvalidExecutable(t *testing.T) {
	runner := &fakeFFmpegRunner{}
	if _, err := NewFFmpegExecutor("ffmpeg", runner); !errors.Is(err, ErrInvalidFFmpegExecutor) {
		t.Fatalf("relative executable error=%v", err)
	}
	executor, _ := NewFFmpegExecutor("/usr/bin/ffmpeg", runner)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request := FFmpegExecutionRequest{
		Transform:  TransformRequest{sourceID: "source-1", requirements: []TransformRequirement{TransformMediaTranscode}},
		InputPath:  "/srv/video/input.mkv",
		OutputPath: "/srv/video/output.mp4",
	}
	if err := executor.Execute(ctx, request); !errors.Is(err, context.Canceled) || runner.calls != 0 {
		t.Fatalf("calls=%d err=%v", runner.calls, err)
	}
}
