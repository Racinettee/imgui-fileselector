package imguifileselector

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/inkyblackness/imgui-go/v4"
)

type SelectorPurpose byte

const (
	PurposeOpen SelectorPurpose = iota
	PurposeSave
)

// Sets the text of the chooser button when the dialog is for opening files
var OpenButtonText string = "Open"

// Sets the text of the chooser button when the dialog is for saving files
var SaveButtonText string = "Save"

// Sets the text of the close dialog button
var CloseButtonText string = "Close"

// FsReader is an interface to implement non default file system functions for the file selector
type FsReader interface {
	Root() string
	IsDir(string) bool
	ReadDir(string) []string
	PathSep() string
}

type defaultFsReader struct{}

func (*defaultFsReader) Root() string {
	return os.Getenv("SystemDrive") + string(os.PathSeparator)
}

func (*defaultFsReader) IsDir(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

func (*defaultFsReader) ReadDir(directory string) []string {
	items, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil
	}
	var results []string
	for _, item := range items {
		results = append(results, item.Name())
	}
	return results
}

func (*defaultFsReader) PathSep() string {
	return string(os.PathSeparator)
}

// DefaultFsReader returns an object that uses default filesystem functions
func DefaultFsReader() FsReader {
	return &defaultFsReader{}
}

// An object representing all the state of a file selector dialog
type FileSelector struct {
	// The path the selector was created at
	Path string
	// The current selection
	Selection string
	// A functoin that will be called when open or save is pressed
	OnChoosePressed func(dir, file string)
	// A function that will be called if the close button is pressed
	OnClosePressed func()
	// The intention that the dialog was created for
	SelectorPurpose  SelectorPurpose
	listing          []string
	currentDirectory string
	currentSelection int32
	fsReader         FsReader
}

func OpenFileSelectorWithReader(path string, fsReader FsReader) (FileSelector, error) {
	result := FileSelector{
		Path:            path,
		SelectorPurpose: PurposeOpen,
		OnChoosePressed: func(_, _ string) {},
		OnClosePressed:  func() {},
		fsReader:        fsReader,
	}
	err := result.buildListing(path)
	return result, err
}

func OpenFileSelector(path string) (FileSelector, error) {
	return OpenFileSelectorWithReader(path, DefaultFsReader())
}

func SaveFileSelectorWithReader(path string, fsReader FsReader) (FileSelector, error) {
	result := FileSelector{
		Path:            path,
		SelectorPurpose: PurposeSave,
		OnChoosePressed: func(_, _ string) {},
		OnClosePressed:  func() {},
		fsReader:        fsReader,
	}
	err := result.buildListing(path)
	return result, err
}

func SaveFileSelector(path string) (FileSelector, error) {
	return SaveFileSelectorWithReader(path, DefaultFsReader())
}

func (fileSelector FileSelector) DialogLabel() string {
	if fileSelector.SelectorPurpose == PurposeOpen {
		return fmt.Sprintf("%v File", OpenButtonText)
	} else {
		return fmt.Sprintf("%v File", SaveButtonText)
	}
}

func (fileSelector *FileSelector) buildListing(path string) (err error) {
	fileSelector.listing = make([]string, 0)
	fileSelector.currentDirectory, err = filepath.Abs(path)
	if err != nil {
		return
	}
	if path != fileSelector.fsReader.Root() {
		fileSelector.listing = append(fileSelector.listing, "..")
	}
	fileSelector.listing = append(fileSelector.listing, fileSelector.fsReader.ReadDir(fileSelector.currentDirectory)...)

	return
}

// Call Update between your imgui begin/end frame calls
func (fileSelector *FileSelector) Update() {
	if imgui.BeginPopupModal(fileSelector.DialogLabel()) {
		imgui.Text(fileSelector.currentDirectory)
		label := OpenButtonText
		if fileSelector.SelectorPurpose == PurposeSave {
			label = SaveButtonText
		}
		if imgui.ListBox(label, &fileSelector.currentSelection, fileSelector.listing) {
			fileSelector.Selection = fileSelector.listing[fileSelector.currentSelection]
			fpathSelection := filepath.Join(fileSelector.currentDirectory, fileSelector.Selection)
			if fileSelector.fsReader.IsDir(fpathSelection) {
				fileSelector.buildListing(fpathSelection)
			}
		}
		if imgui.Button(label) {
			fileSelector.OnChoosePressed(fileSelector.currentDirectory, fileSelector.Selection)
			imgui.CloseCurrentPopup()
		}
		if imgui.Button(CloseButtonText) {
			fileSelector.OnClosePressed()
			imgui.CloseCurrentPopup()
		}
		imgui.EndPopup()
	}
}
