// Mocking os FS interface, through the usage of a map
package memos

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"
)

type memosNode interface {
	os.FileInfo
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type memosFile struct {
	name    string
	mode    os.FileMode
	modTime time.Time
	content []byte
}

func (f *memosFile) update(data []byte) {
	f.content = data
	f.modTime = time.Now()
}

func (f *memosFile) Write(b []byte) (n int, err error) {
	if f.mode|fs.ModeAppend == 1 {
		f.content = append(f.content, b...)
	} else {
		f.content = b
	}
	return len(b), nil
}
func (f *memosFile) Close() error { return nil }

func (f memosFile) Name() string       { return f.name }
func (f memosFile) Size() int64        { return int64(len(f.content)) }
func (f memosFile) ModTime() time.Time { return f.modTime }
func (f memosFile) Mode() os.FileMode  { return f.mode }
func (f memosFile) IsDir() bool        { return false }
func (f memosFile) Sys() any           { return nil }

var _ File = &memosFile{}
var _ memosNode = &memosFile{}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type memosDir struct {
	name     string
	mode     os.FileMode
	modTime  time.Time
	children map[string]memosNode
}

func (d memosDir) Name() string       { return d.name }
func (d memosDir) Size() int64        { return 0 }
func (d memosDir) ModTime() time.Time { return d.modTime }
func (d memosDir) Mode() os.FileMode  { return d.mode }
func (d memosDir) IsDir() bool        { return true }
func (d memosDir) Sys() any           { return nil }

var _ memosNode = &memosDir{}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type InMemoryOS struct {
	// fs map[string]memosNode
	Root memosDir
	wd string
}

func NewInMemoryOS(workingDir string) *InMemoryOS {
	return &InMemoryOS{
		Root: memosDir{
			name: "/",
			mode: 0o777,
			modTime: time.Now(),
			children: make(map[string]memosNode),
		},
		wd: workingDir,
	}
}

func (InMemoryOS) splitPath(path string) []string {
	return strings.Split(path, "/")
}

// Resolves a path. Has following possible output combos
// - Existing file: memosFile{...}, true, nil
// - Existing dir: memosDir{...}, false, nil
// - Invalid internal state: (nil), (false), error
// - File not found: nil, (false), (nil)
// - Parent directory does not exist
//
// (values in parentheses can be ignored for that specific state)
func (r *InMemoryOS) resolve(name string, notFoundOk bool) (node memosNode, isFile bool, err error) {
	pathParts := r.splitPath(name)
	fmt.Printf("%#v\n", pathParts)
	dir := &r.Root.children

	// TODO: generalize this with mkdirall
	for i, pathPart := range pathParts {
		if i == 0 && pathPart == "" {
			continue
		}
		
		fmt.Printf("[%d] resolving %s for %s\n", i, pathPart, name)
		isLeaf := (i == len(pathParts) - 1)
		fmt.Printf("  isLeaf=%v\n", isLeaf)
		
		next, exists := (*dir)[pathPart]
		if !exists {
			fmt.Println("  not found")
			if notFoundOk {
				fmt.Println("  notFoundOk = true")
				return nil, false, nil
			}
			err = fmt.Errorf(`parent dir "%s" for %s does not exist`, pathPart, name)
			return
		}
		
		fmt.Println("  found")

		switch nextTyped := next.(type) {
		case memosFile:
			isFile = true
			if isLeaf {
				node = &nextTyped
			} else {
				node = nil
			}
			return

		case memosDir:
			if isLeaf {
				isFile = false
				node = &nextTyped
				return
			}
			dir = &nextTyped.children

		default:
			err = fmt.Errorf("memOS node contains invalid value %v (%v)", next, r)
			return
		}
	}

	return nil, false, nil
}

func (r *InMemoryOS) resolveParent(name string) (dir *memosDir, err error) {
	node, _, err := r.resolve(path.Dir(name), false)
	if err != nil {
		return nil, err
	}

	dir, ok := node.(*memosDir)
	if !ok {
		return nil, fmt.Errorf("parent of path %s is somehow not directory... how??", name)
	}

	return
}

// OS interface --------------------------------------------------------------------------

func (r *InMemoryOS) Create(name string) (File, error) {
	node, isFile, err := r.resolve(name, true)
	if err != nil {
		return nil, err
	}
	// file exists: truncate
	if node != nil {
		if isFile {
			file, _ := node.(*memosFile)
			file.content = nil
			return file, nil
		}
		return nil, fmt.Errorf("%s is directory and already exists", name)
	}

	// find parent
	parent, err := r.resolveParent(name)
	if err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, fmt.Errorf("cannot create %s: parent does not exist", name)
	}

	// create file inside parent
	filename := path.Base(name)
	newFile := memosFile{
		name:    filename,
		mode:    0o666, // defined by spec
		modTime: time.Now(),
		content: nil,
	}
	parent.children[filename] = newFile

	return &newFile, nil

}

func (r *InMemoryOS) MkdirAll(path string, perm os.FileMode) (err error) {
	node, isFile, err := r.resolve(path, true)
	if err != nil {
		return
	}

	// dir exists: do nothing
	if node != nil {
		if isFile {
			err = fmt.Errorf("path %s already exists and is file", path)
		} else {
			err = nil
		}
		return
	}

	// create dir
	pathParts := r.splitPath(path)
	dir := &r.Root.children

	for _, pathPart := range pathParts[1:] {
		next, exists := (*dir)[pathPart]
		// create if not exists
		if !exists {
			fmt.Printf("creating %s\n", pathPart)
			next = memosDir{
				name:     pathPart,
				mode:     perm | fs.ModeDir,
				modTime: time.Now(),
				children: make(map[string]memosNode),
			}
			(*dir)[pathPart] = next.(memosDir)
			fmt.Printf("  partial: %v\n", (*dir)[pathPart])
			fmt.Printf("  total: %#v\n", r.Root)
		}

		switch nextTyped := next.(type) {
		case memosFile:
			return fmt.Errorf("cannot create path %s because child %s is not directory", path, next)

		case memosDir:
			dir = &nextTyped.children

		default:
			return fmt.Errorf("memOS node contains invalid value %v", next)
		}
	}

	return
}

func (r *InMemoryOS) Getwd() (dir string, err error) { return r.wd, nil }

func (r *InMemoryOS) Stat(name string) (os.FileInfo, error) {
	node, _, err := r.resolve(name, false)
	if err != nil {
		return nil, err
	}

	// memosNode embeds os.FileInfo
	return node, nil
}

func (r *InMemoryOS) ReadFile(name string) ([]byte, error) {
	node, isFile, err := r.resolve(name, false)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, fmt.Errorf("file %s does not exist", name)
	}
	if !isFile {
		return nil, fmt.Errorf("%s is not file", name)
	}

	file, ok := node.(*memosFile)
	if !ok {
		return nil, fmt.Errorf("file %s does not type assert to memosFile", name)
	}

	return file.content, err
}

func (r *InMemoryOS) WriteFile(name string, data []byte, perm os.FileMode) (err error) {
	node, isFile, err := r.resolve(name, false)
	if err != nil {
		return err
	}

	// cannot write to directory
	if !isFile {
		return fmt.Errorf("%s is not file", name)
	}

	var file *memosFile
	// if file does NOT exist, create it and set perm
	if node == nil {
		newFile, err := r.Create(name)
		if err != nil {
			return err
		}
		file = newFile.(*memosFile)
		file.mode = perm
	} else
	// if file exists, truncate it
	{
		typedFile, ok := node.(*memosFile)
		if !ok {
			return fmt.Errorf("file %s does not type assert to memosFile", name)
		}
		file = typedFile
	}

	file.update(data)
	return err
}

func (r *InMemoryOS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	node, isFile, err := r.resolve(name, false)
	if err != nil {
		return nil, err
	}
	if !isFile {
		return nil, fmt.Errorf("%s is not file", name)
	}

	var file *memosFile
	// If file does not exist
	if node == nil {
		// ... and O_CREATE is passed, create file with mode perm
		if flag&os.O_CREATE == 1 {
			newFile, err := r.Create(name)
			if err != nil {
				return nil, err
			}
			file = newFile.(*memosFile)
			file.mode = perm
		} else
		// otherwise, cannot open it
		{
			return nil, fmt.Errorf("cannot open non-existing file %s w/o O_CREATE", name)
		}
	} else {
		typedFile, ok := node.(*memosFile)
		if !ok {
			return nil, fmt.Errorf("file %s does not type assert to memosFile", name)
		}
		file = typedFile
	}

	return file, err
}

var _ OS = &InMemoryOS{}
