package processor

import (
    "bytes"
    "fmt"
    "os"
    "os/exec"

    "github.com/yuin/goldmark"
    "github.com/yuin/goldmark/extension"
    "github.com/yuin/goldmark/renderer/html"
)

func ConvertMDToPDF(mdContent []byte, outputPath string) error {
    // Step 1: Markdown → HTML using goldmark
    md := goldmark.New(
        goldmark.WithExtensions(
            extension.GFM,        // GitHub Flavored Markdown
            extension.Table,
            extension.Strikethrough,
        ),
        goldmark.WithRendererOptions(
            html.WithUnsafe(), // Allow raw HTML in markdown
        ),
    )

    var htmlBuf bytes.Buffer
    if err := md.Convert(mdContent, &htmlBuf); err != nil {
        return fmt.Errorf("markdown to HTML conversion failed: %w", err)
    }

    // Step 2: Wrap in a styled HTML document
    fullHTML := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body {
            font-family: 'Segoe UI', Arial, sans-serif;
            max-width: 860px;
            margin: 40px auto;
            padding: 0 30px;
            color: #333;
            line-height: 1.6;
        }
        h1, h2, h3 { color: #1a1a2e; border-bottom: 1px solid #eee; padding-bottom: 6px; }
        code {
            background: #f4f4f4;
            padding: 2px 6px;
            border-radius: 3px;
            font-family: monospace;
        }
        pre {
            background: #f8f8f8;
            border-left: 4px solid #4a90d9;
            padding: 14px;
            border-radius: 4px;
            overflow-x: auto;
        }
        table { border-collapse: collapse; width: 100%%; }
        th, td { border: 1px solid #ddd; padding: 8px 12px; }
        th { background: #f0f0f0; }
        blockquote { border-left: 4px solid #ccc; margin: 0; padding-left: 16px; color: #666; }
        img { max-width: 100%%; }
    </style>
</head>
<body>
%s
</body>
</html>`, htmlBuf.String())

    // Step 3: Write HTML to a temp file
    tmpHTML, err := os.CreateTemp("", "mdtopdf-*.html")
    if err != nil {
        return fmt.Errorf("failed to create temp HTML file: %w", err)
    }
    defer os.Remove(tmpHTML.Name())

    if _, err := tmpHTML.WriteString(fullHTML); err != nil {
        return fmt.Errorf("failed to write HTML: %w", err)
    }
    tmpHTML.Close()

    // Step 4: HTML → PDF using wkhtmltopdf binary
    cmd := exec.Command("wkhtmltopdf",
        "--quiet",
        "--margin-top", "15mm",
        "--margin-bottom", "15mm",
        "--margin-left", "15mm",
        "--margin-right", "15mm",
        "--page-size", "A4",
        "--encoding", "UTF-8",
        "--enable-local-file-access",
        tmpHTML.Name(),
        outputPath,
    )

    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("wkhtmltopdf error: %w | output: %s", err, string(output))
    }
    return nil
}