package file

import (
	"archive/zip"
	"bytes"
	"mime/multipart"
	"strings"
	"testing"
)

func header(name, contentType string, data []byte) *multipart.FileHeader {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", name)
	if err != nil {
		panic(err)
	}
	if _, err = part.Write(data); err != nil {
		panic(err)
	}
	if err = w.Close(); err != nil {
		panic(err)
	}
	r := multipart.NewReader(bytes.NewReader(body.Bytes()), w.Boundary())
	form, err := r.ReadForm(int64(body.Len()))
	if err != nil {
		panic(err)
	}
	h := form.File["file"][0]
	h.Header.Set("Content-Type", contentType)
	return h
}

func zippedXML(t *testing.T, name, content string) []byte {
	t.Helper()
	var data bytes.Buffer
	zw := zip.NewWriter(&data)
	w, err := zw.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = w.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err = zw.Close(); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}

func TestParseTextAndCode(t *testing.T) {
	content, err := parse(header("main.go", "text/plain", []byte("package main")), false)
	if err != nil || content != "package main" {
		t.Fatalf("content=%q err=%v", content, err)
	}
}

func TestParseModernOfficeDocument(t *testing.T) {
	data := zippedXML(t, "word/document.xml", `<w:document xmlns:w="w"><w:p><w:t>Hello Word</w:t></w:p></w:document>`)
	content, err := parse(header("example.docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", data), false)
	if err != nil || !strings.Contains(content, "Hello Word") {
		t.Fatalf("content=%q err=%v", content, err)
	}
}

func TestLegacyOfficeErrorIsActionable(t *testing.T) {
	_, err := parse(header("example.doc", "application/msword", []byte("binary")), false)
	if err == nil || !strings.Contains(err.Error(), "DOCX") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVisionImageBecomesDataURL(t *testing.T) {
	content, err := parse(header("pixel.png", "image/png", []byte("image")), true)
	if err != nil || !strings.HasPrefix(content, "data:image/png;base64,") {
		t.Fatalf("content=%q err=%v", content, err)
	}
}
