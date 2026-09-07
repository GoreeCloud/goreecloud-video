package playback

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

var (
	ErrInvalidFFmpegExecutor       = errors.New("invalid ffmpeg executor")
	ErrInvalidFFmpegExecution      = errors.New("invalid ffmpeg execution request")
	ErrUnsafeSubtitleFilterPath    = errors.New("unsafe subtitle filter path")
	ErrFFmpegExecutionFailed       = errors.New("ffmpeg execution failed")
)

const hdrToSDRFilter = "zscale=t=linear:npl=100,format=gbrpf32le,zscale=p=bt709,tonemap=hable:desat=0,zscale=t=bt709:m=bt709:r=tv,format=yuv420p"

// FFmpegExecutionRequest binds an already-validated opaque transform request to
// filesystem paths authorized by a higher-level storage adapter. SourceID is
// never interpreted as a path or URL here.
type FFmpegExecutionRequest struct {
	Transform    TransformRequest
	InputPath    string
	OutputPath   string
	SubtitlePath string
}

// FFmpegProcessRunner isolates process creation for deterministic tests. The
// production implementation below executes argv directly without a shell.
type FFmpegProcessRunner interface {
	Run(ctx context.Context, binary string, args []string) error
}

type OSFFmpegProcessRunner struct{}

func (OSFFmpegProcessRunner) Run(ctx context.Context, binary string, args []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	command := exec.CommandContext(ctx, binary, args...)
	command.Stdin = nil
	command.Stdout = io.Discard
	command.Stderr = io.Discard
	if err := command.Run(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return fmt.Errorf("%w: %v", ErrFFmpegExecutionFailed, err)
	}
	return nil
}

type FFmpegExecutor struct {
	binary string
	runner FFmpegProcessRunner
}

func NewFFmpegExecutor(binary string, runner FFmpegProcessRunner) (*FFmpegExecutor, error) {
	if !validAuthorizedExecutablePath(binary) || runner == nil {
		return nil, ErrInvalidFFmpegExecutor
	}
	return &FFmpegExecutor{binary: binary, runner: runner}, nil
}

// Execute runs one FFmpeg transformation. The caller owns authorization of the
// supplied paths and lifecycle of the output artifact; this method validates
// path shape, builds fixed argv, and executes without a shell.
func (e *FFmpegExecutor) Execute(ctx context.Context, request FFmpegExecutionRequest) error {
	if e == nil || e.runner == nil || !validAuthorizedExecutablePath(e.binary) {
		return ErrInvalidFFmpegExecutor
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	args, err := BuildFFmpegArgs(request)
	if err != nil {
		return err
	}
	return e.runner.Run(ctx, e.binary, args)
}

func BuildFFmpegArgs(request FFmpegExecutionRequest) ([]string, error) {
	if !validTransformRequest(request.Transform) ||
		!validAuthorizedMediaPath(request.InputPath) ||
		!validAuthorizedMediaPath(request.OutputPath) ||
		request.InputPath == request.OutputPath {
		return nil, ErrInvalidFFmpegExecution
	}

	requirements := request.Transform.Requirements()
	needsSubtitle := false
	needsHDRToneMap := false
	for _, requirement := range requirements {
		switch requirement {
		case TransformMediaTranscode:
		case TransformSubtitleBurnIn:
			needsSubtitle = true
		case TransformHDRToneMap:
			needsHDRToneMap = true
		default:
			return nil, ErrInvalidFFmpegExecution
		}
	}
	if needsSubtitle {
		if !validAuthorizedMediaPath(request.SubtitlePath) {
			return nil, ErrInvalidFFmpegExecution
		}
		if !safeSubtitleFilterPath(request.SubtitlePath) {
			return nil, ErrUnsafeSubtitleFilterPath
		}
	} else if request.SubtitlePath != "" {
		return nil, ErrInvalidFFmpegExecution
	}

	args := []string{
		"-nostdin",
		"-hide_banner",
		"-loglevel", "error",
		"-y",
		"-i", request.InputPath,
		"-map", "0:v:0",
		"-map", "0:a?",
	}

	filters := make([]string, 0, 2)
	if needsHDRToneMap {
		filters = append(filters, hdrToSDRFilter)
	}
	if needsSubtitle {
		filters = append(filters, "subtitles=filename='"+filepath.ToSlash(request.SubtitlePath)+"'")
	}
	if len(filters) > 0 {
		args = append(args, "-vf", strings.Join(filters, ","))
	}

	args = append(args,
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", "20",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-b:a", "192k",
		"-movflags", "+faststart",
		"-f", "mp4",
		request.OutputPath,
	)
	return args, nil
}

func validAuthorizedExecutablePath(path string) bool {
	return validAbsoluteCleanPath(path)
}

func validAuthorizedMediaPath(path string) bool {
	return validAbsoluteCleanPath(path)
}

func validAbsoluteCleanPath(path string) bool {
	return path != "" &&
		strings.TrimSpace(path) == path &&
		!strings.ContainsRune(path, '\x00') &&
		filepath.IsAbs(path) &&
		filepath.Clean(path) == path
}

// safeSubtitleFilterPath is deliberately stricter than ordinary authorized
// paths because the path is embedded inside FFmpeg's filtergraph grammar. This
// prevents commas, semicolons, brackets, quotes, backslashes, colons, control
// characters, and other filtergraph metacharacters from becoming syntax.
func safeSubtitleFilterPath(path string) bool {
	for _, r := range filepath.ToSlash(path) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		switch r {
		case '/', '.', '_', '-':
			continue
		default:
			return false
		}
	}
	return true
}
