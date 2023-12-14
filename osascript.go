package outbow

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type OsaScript struct {
	PageNumberContainer     PageNumberContainer
	Path                    string
	CommandResult           *CommandResult
	ClipboardContent        string
	ClipboardContentPath    string
	AllowReviewsLoadSeconds int
}

func (script *OsaScript) SaveClipboardContent() error {
	if err := os.WriteFile(script.ClipboardContentPath, []byte(script.ClipboardContent), 0o600); err != nil {
		fmt.Println("Error:", err)
		return err
	}

	return nil
}

func (script *OsaScript) WriteApplescript(goproModel string) error {
	tmpl, err := template.ParseFiles("gopro.scpt.tmpl")
	if err != nil {
		return fmt.Errorf("error reading template: %v", err)
	}

	data := struct {
		MyURL                   string
		AllowReviewsLoadSeconds int
	}{
		MyURL:                   script.PageNumberContainer.URL.String(),
		AllowReviewsLoadSeconds: script.AllowReviewsLoadSeconds,
	}

	var applescriptBuf bytes.Buffer
	if err := tmpl.Execute(&applescriptBuf, data); err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	p := fmt.Sprintf("gopro-%s-%s.scpt", strings.ToLower(goproModel), numberFormatSpecifier)

	err = os.MkdirAll(DataDirAbsPath, os.ModePerm)
	if err != nil {
		slog.Error("mkdir had error", "dir", DataDirAbsPath, "error", err)
		return err
	}

	fname := fmt.Sprintf(p, script.PageNumberContainer.PageNumber)
	script.Path = filepath.Join(DataDirAbsPath, fname)
	if err := writeToFile(script.Path, applescriptBuf.Bytes()); err != nil {
		slog.Error("writing applescript to file", "error", err)
		return err
	}

	args := []string{script.Path}
	script.CommandResult = &CommandResult{
		Command: "osascript",
		Args:    args,
	}

	return nil
}
