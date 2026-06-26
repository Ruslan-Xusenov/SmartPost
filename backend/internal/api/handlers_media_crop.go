package api

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func cropVideoToSquare(ctx context.Context, input io.Reader) (*os.File, error) {
	// Create temp files
	inTemp, err := os.CreateTemp("", "in-*.mp4")
	if err != nil {
		return nil, err
	}
	defer func() {
		inTemp.Close()
		os.Remove(inTemp.Name())
	}()

	_, err = io.Copy(inTemp, input)
	if err != nil {
		return nil, err
	}

	outTempPath := filepath.Join(os.TempDir(), filepath.Base(inTemp.Name())+"-out.mp4")

	// Run ffmpeg
	// crop='min(iw,ih)':'min(iw,ih)'
	// scale=384:384 (Telegram recommends this size for video notes)
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", inTemp.Name(),
		"-vf", "crop='min(iw,ih)':'min(iw,ih)',scale=384:384",
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-c:a", "aac",
		"-f", "mp4",
		"-y", outTempPath,
	)

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return os.Open(outTempPath)
}
