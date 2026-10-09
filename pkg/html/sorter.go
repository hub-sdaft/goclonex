package html

import "github.com/hub-sdaft/goclonex/pkg/project"

// LinkRestructure grabs all html files in project directory
// reorganizes each file with local links (css js images)
func LinkRestructure(projectDir string) error {
	// Redirect JS/CSS/Img tags to the correct place :)
	return arrange(projectDir)
}

func ProjectLinkRestructure(proj *project.Project) error {
	return projectArrange(proj)
}
