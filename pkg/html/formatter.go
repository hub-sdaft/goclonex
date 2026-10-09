package html

import (
	"os"

	"github.com/hub-sdaft/goclonex/pkg/project"
	"github.com/yosssi/gohtml"
)

// FormatHTML will formart any given string of HTML
func FormatHTML(filePath string) {
	// TODO: Implement
	dat, err := os.ReadFile(filePath)	// XXX
	if err != nil {
		panic(err)
	}
	data := gohtml.Format(string(dat))
	b := []byte(data)
	error := os.WriteFile(filePath, b, 0777)	// XXX
	// handle this error
	if error != nil {
		// print it out
		panic(error)
	}
}

func ProjectFormatHTML(proj *project.Project) {
	content := proj.IndexFile.Content
	if content == nil {
		return
	}
	proj.IndexFile.Content = gohtml.FormatBytes(content)
}
