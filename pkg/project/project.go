package project

import (
	"fmt"
	"iter"
	"maps"
	"os"
	"path"
	"sync"

	pkgFile "github.com/hub-sdaft/goclonex/pkg/file"
)

// --------------------------------------------------------------

type File struct {
	Name    string
	Content []byte
}

func (f *File) Save(dir string) (err error) {
	filePath := path.Join(dir, f.Name)

	osFile, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", filePath, err)
	}
	defer osFile.Close()

	_, err = osFile.Write(f.Content)
	if err != nil {
		return fmt.Errorf("error writing file %s: %w", filePath, err)
	}

	return
}

func SaveFiles(files iter.Seq[File], dir string) (err error) {
	for file := range files {
		err = file.Save(dir)
		if err != nil {
			return fmt.Errorf("Error saving %s: %w", file.Name, err)
		}
	}
	return
}

// ------------------------------------------------------------------


type rwMap[K comparable, V any] struct {
	data map[K]V
	lock sync.RWMutex
}

func (m *rwMap[K, V]) Get(key K) (V, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if m.data == nil {
		m.data = make(map[K]V)
	}
	v, ok := m.data[key]
	return v, ok
}

func (m *rwMap[K, V]) Set(key K, val V) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.data[key] = val
}

func (m *rwMap[K, V]) Len() int {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return len(m.data)
}

func (m *rwMap[K, V]) Values() iter.Seq[V] {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return maps.Values(m.data)
}


var ProjectFolders = [...]string{
	pkgFile.CSSFolder,
	pkgFile.JSFolder,
	pkgFile.ImgFolder,
	pkgFile.OtherFolder,
}

type Project struct {
	name    string
	url     string

	IndexFile File

	cssFiles   rwMap[string, File]
	jsFiles    rwMap[string, File]
	imgFiles   rwMap[string, File]
	otherFiles rwMap[string, File]
}

func NewProject(name, url string) Project {
	return Project{
		name:       name,
		url:        url,
	}
}

func (p *Project) Name() string { return p.name }
func (p *Project) Url() string  { return p.url }

func (p *Project) Debug() string {
	return fmt.Sprintf(`
		URL:  %s
		Name: %s

		CSS: %d files
		JS: %d files
		IMG: %d files
		`,
		p.Url(),
		p.Name(),
		p.cssFiles.Len(),
		p.jsFiles.Len(),
		p.imgFiles.Len(),
	)
}

func (p *Project) AddFile(file File, folder string) (err error) {
	var fileMap *rwMap[string, File]
	switch folder {
	case pkgFile.CSSFolder:
		fileMap = &p.cssFiles
	case pkgFile.JSFolder:
		fileMap = &p.jsFiles
	case pkgFile.ImgFolder:
		fileMap = &p.imgFiles
	case pkgFile.OtherFolder:
		fileMap = &p.otherFiles
	default:
		return fmt.Errorf("cannot add file %s path to invalid folder %s", file.Name, folder)
	}

	if _, exists := fileMap.Get(file.Name); exists {
		return fmt.Errorf("cannot add file %s to folder %s, because it exists already", file.Name, folder)
	}

	fileMap.Set(file.Name, file)
	return
}

func (p *Project) Save(rootDir string) (err error) {
	// Create root dir
	err = os.MkdirAll(rootDir, 0777)
	if err != nil {
		return
	}

	// Create all sub folders
	for _, folder := range ProjectFolders {
		err = os.MkdirAll(path.Join(rootDir, folder), 0777)
		if err != nil {
			return
		}
	}

	// Create index file
	err = p.IndexFile.Save(rootDir)
	if err != nil {
		return
	}

	// Save the file maps to the corresponding folders
	mapsToFolders := map[*rwMap[string, File]]string{
		&p.cssFiles:   pkgFile.CSSFolder,
		&p.jsFiles:    pkgFile.JSFolder,
		&p.imgFiles:   pkgFile.ImgFolder,
		&p.otherFiles: pkgFile.OtherFolder,
	}
	for fileMap, folder := range mapsToFolders {
		SaveFiles(fileMap.Values(), path.Join(rootDir, folder))
	}

	return
}
