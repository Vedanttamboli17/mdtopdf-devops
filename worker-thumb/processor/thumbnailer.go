package processor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GenerateThumbnail renders the first page of a PDF to a JPEG at 150 DPI
// using pdftoppm (poppler-utils) and writes the result to outputPath.
func GenerateThumbnail(pdfContent []byte, outputPath string) error {
	tmpPDF, err := os.CreateTemp("/tmp", "docplatform-*.pdf")
	if err != nil {
		return fmt.Errorf("create temp pdf: %w", err)
	}
	tmpPDFPath := tmpPDF.Name()
	defer os.Remove(tmpPDFPath)

	if _, err := tmpPDF.Write(pdfContent); err != nil {
		tmpPDF.Close()
		return fmt.Errorf("write temp pdf: %w", err)
	}
	tmpPDF.Close()

	// pdftoppm writes {prefix}-<N>.jpg where N is zero-padded by page count
	thumbPrefix := strings.TrimSuffix(tmpPDFPath, ".pdf") + "-thumb"

	cmd := exec.Command("pdftoppm",
		"-jpeg",
		"-r", "150",
		"-f", "1",
		"-l", "1",
		tmpPDFPath,
		thumbPrefix,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pdftoppm failed: %w | output: %s", err, string(output))
	}

	// Locate the generated file — pdftoppm zero-pads based on total page count
	matches, err := filepath.Glob(thumbPrefix + "*.jpg")
	if err != nil || len(matches) == 0 {
		return fmt.Errorf("pdftoppm output not found at %s*.jpg", thumbPrefix)
	}
	thumbFile := matches[0]
	defer os.Remove(thumbFile)

	src, err := os.Open(thumbFile)
	if err != nil {
		return fmt.Errorf("open thumbnail: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("copy thumbnail: %w", err)
	}

	return nil
}
