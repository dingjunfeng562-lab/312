package file

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/ledongthuc/pdf"
)

const maxUploadSize = 100 << 20

var textExtensions = map[string]bool{
	".txt": true, ".md": true, ".markdown": true, ".rtf": true,
	".c": true, ".cc": true, ".cpp": true, ".cxx": true, ".h": true, ".hpp": true,
	".cs": true, ".go": true, ".java": true, ".js": true, ".jsx": true,
	".ts": true, ".tsx": true, ".py": true, ".rb": true, ".rs": true, ".php": true,
	".swift": true, ".kt": true, ".kts": true, ".scala": true, ".sh": true,
	".bash": true, ".zsh": true, ".fish": true, ".ps1": true, ".bat": true,
	".cmd": true, ".sql": true, ".html": true, ".htm": true, ".css": true,
	".scss": true, ".sass": true, ".less": true, ".vue": true, ".svelte": true,
	".astro": true, ".dart": true, ".lua": true, ".csv": true, ".tsv": true,
	".json": true, ".jsonl": true, ".ndjson": true, ".xml": true, ".yaml": true,
	".yml": true, ".toml": true, ".ini": true, ".cfg": true, ".conf": true,
	".log": true, ".properties": true, ".env": true,
}

func response(c *gin.Context, content string, err error) {
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "content": "", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "content": content})
}

func UploadAPI(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)
	header, err := c.FormFile("file")
	if err != nil {
		response(c, "", errors.New("missing file or file exceeds the 100 MB limit"))
		return
	}

	content, err := parse(header, c.PostForm("enable_vision") == "true")
	response(c, strings.TrimSpace(content), err)
}

func parse(header *multipart.FileHeader, vision bool) (string, error) {
	f, err := header.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, maxUploadSize+1))
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("file is empty")
	}
	if len(data) > maxUploadSize {
		return "", errors.New("file exceeds the 100 MB limit")
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	contentType := header.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = http.DetectContentType(data)
	}

	if textExtensions[ext] || strings.HasPrefix(contentType, "text/") {
		if !utf8.Valid(data) {
			return "", errors.New("text file is not UTF-8 encoded")
		}
		return strings.TrimPrefix(string(data), "\ufeff"), nil
	}
	if strings.HasPrefix(contentType, "image/") {
		if !vision {
			return "", errors.New("this image requires a vision-capable model or a configured OCR service")
		}
		return fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(data)), nil
	}
	if strings.HasPrefix(contentType, "audio/") || map[string]bool{".mp3": true, ".wav": true, ".m4a": true, ".aac": true, ".ogg": true, ".flac": true, ".opus": true}[ext] {
		return "", errors.New("audio transcription service is not configured; configure an external file parsing service in System Settings")
	}

	switch ext {
	case ".pdf":
		return parsePDF(data)
	case ".docx", ".pptx", ".xlsx", ".odt", ".odp", ".ods":
		return parseOpenDocument(data)
	case ".doc", ".ppt", ".xls":
		return "", errors.New("legacy binary Office files are not supported locally; save the file as DOCX, PPTX, or XLSX")
	default:
		return "", fmt.Errorf("unsupported file format %q", ext)
	}
}

func parsePDF(data []byte) (string, error) {
	temp, err := os.CreateTemp("", "coai-*.pdf")
	if err != nil {
		return "", err
	}
	name := temp.Name()
	defer os.Remove(name)
	if _, err = temp.Write(data); err != nil {
		temp.Close()
		return "", err
	}
	temp.Close()

	f, reader, err := pdf.Open(name)
	if err != nil {
		return "", fmt.Errorf("invalid PDF: %w", err)
	}
	defer f.Close()
	plain, err := reader.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("failed to extract PDF text: %w", err)
	}
	data, err = io.ReadAll(plain)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(string(data)) == "" {
		return "", errors.New("PDF contains no extractable text; scanned PDFs require a configured OCR service")
	}
	return string(data), nil
}

func parseOpenDocument(data []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", errors.New("invalid or encrypted Office document")
	}
	files := make([]*zip.File, 0)
	for _, f := range zr.File {
		name := strings.ToLower(strings.ReplaceAll(f.Name, "\\", "/"))
		if strings.HasSuffix(name, ".xml") && (strings.HasPrefix(name, "word/") || strings.HasPrefix(name, "ppt/slides/") || strings.HasPrefix(name, "xl/") || name == "content.xml") {
			files = append(files, f)
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })

	var out strings.Builder
	for _, f := range files {
		r, err := f.Open()
		if err != nil {
			continue
		}
		decoder := xml.NewDecoder(r)
		for {
			token, err := decoder.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}
			if chars, ok := token.(xml.CharData); ok {
				text := strings.TrimSpace(string(chars))
				if text != "" {
					out.WriteString(text)
					out.WriteByte('\n')
				}
			}
		}
		r.Close()
	}
	if strings.TrimSpace(out.String()) == "" {
		return "", errors.New("document contains no extractable text or is encrypted")
	}
	return out.String(), nil
}
