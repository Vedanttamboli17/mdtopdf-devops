package processor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// ConvertDOCXToPDF converts a DOCX byte slice to PDF using LibreOffice headless mode
// and writes the result to outputPath.
func ConvertDOCXToPDF(docxContent []byte, outputPath string) error {
	// Write DOCX bytes to a temp file so LibreOffice can read it
	tmpDocx, err := os.CreateTemp("/tmp", "docplatform-*.docx")
	if err != nil {
		return fmt.Errorf("create temp docx: %w", err)
	}
	tmpDocxPath := tmpDocx.Name()
	defer os.Remove(tmpDocxPath)

	if _, err := tmpDocx.Write(docxContent); err != nil {
		tmpDocx.Close()
		return fmt.Errorf("write temp docx: %w", err)
	}
	tmpDocx.Close()

	// LibreOffice names its output file {stem}.pdf inside --outdir
	tmpPDFPath := strings.TrimSuffix(tmpDocxPath, ".docx") + ".pdf"
	defer os.Remove(tmpPDFPath)

	cmd := exec.Command("libreoffice",
		"--headless",
		"--norestore",
		"--nofirststartwizard",
		"--convert-to", "pdf",
		"--outdir", "/tmp",
		tmpDocxPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("libreoffice failed: %w | output: %s", err, string(output))
	}

	// Copy the LibreOffice-generated PDF to the caller-specified outputPath
	src, err := os.Open(tmpPDFPath)
	if err != nil {
		return fmt.Errorf("open libreoffice output: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("copy to output: %w", err)
	}

	return nil
}
