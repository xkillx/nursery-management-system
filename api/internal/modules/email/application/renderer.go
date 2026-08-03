package application

import (
	"embed"
	"fmt"
	"html/template"
	"strings"
	texttemplate "text/template"
)

//go:embed templates/*
var templatesFS embed.FS

type Renderer struct{}

func NewRenderer() *Renderer {
	return &Renderer{}
}

func (r *Renderer) Render(templateName string, version int, data map[string]interface{}) (htmlBody, textBody string, err error) {
	htmlFilename := fmt.Sprintf("templates/%s_v%d.html", templateName, version)
	textFilename := fmt.Sprintf("templates/%s_v%d.txt", templateName, version)

	var htmlTmpl *template.Template
	if version == 2 {
		// v2 HTML templates share the layout shell (KTD-5): parse the layout
		// plus the per-type content file, then execute the layout by name, which
		// pulls the content in via the mandatory {{define "content"}} block.
		htmlTmpl, err = template.ParseFS(templatesFS, "templates/layout_v2.html", htmlFilename)
	} else {
		htmlTmpl, err = template.ParseFS(templatesFS, htmlFilename)
	}
	if err != nil {
		return "", "", fmt.Errorf("parse html template %s: %w", htmlFilename, err)
	}

	var htmlBuf strings.Builder
	if version == 2 {
		if err := htmlTmpl.ExecuteTemplate(&htmlBuf, "layout_v2.html", data); err != nil {
			return "", "", fmt.Errorf("execute html template: %w", err)
		}
	} else {
		if err := htmlTmpl.Execute(&htmlBuf, data); err != nil {
			return "", "", fmt.Errorf("execute html template: %w", err)
		}
	}

	textTmpl, err := texttemplate.ParseFS(templatesFS, textFilename)
	if err != nil {
		return "", "", fmt.Errorf("parse text template %s: %w", textFilename, err)
	}

	var textBuf strings.Builder
	if err := textTmpl.Execute(&textBuf, data); err != nil {
		return "", "", fmt.Errorf("execute text template: %w", err)
	}

	return htmlBuf.String(), textBuf.String(), nil
}
