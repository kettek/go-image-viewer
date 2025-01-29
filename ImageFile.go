package main

import (
	"fmt"
	"image"
	"os"
	"path/filepath"

	"gioui.org/op/paint"
	"github.com/qeesung/image2ascii/convert"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "github.com/kettek/apng"
	_ "github.com/kettek/xbm"
	_ "github.com/kettek/xpm"
)

type ImageFile struct {
	path    string
	invalid error
	image   image.Image
	format  string
}

func (i *ImageFile) load() error {
	r, err := os.Open(i.path)
	if err != nil {
		i.invalid = err
		return err
	}
	m, format, err := image.Decode(r)
	if err != nil {
		i.invalid = err
		return err
	}
	i.image = m
	i.format = format
	return nil
}

func (i *ImageFile) unload() {
	i.invalid = nil
	i.image = nil
	i.format = ""
}

func (i *ImageFile) asASCII() (string, error) {
	if i.image == nil {
		err := i.load()
		if err != nil {
			return "", err
		}
	}
	c := convert.NewImageConverter()
	s := c.Image2ASCIIString(i.image, &convert.Options{
		Ratio:       1,
		FixedWidth:  -1,
		FixedHeight: -1,
		FitScreen:   false,
		Colored:     true,
	})

	return s, nil
}

type ImageFiles struct {
	files        []*ImageFile
	currentFile  *ImageFile
	currentIndex int
}

func (f *ImageFiles) CurrentFile() *ImageFile {
	return f.currentFile
}

func (f *ImageFiles) addFiles(files []string) {
	for _, v := range files {

		// Check if file is a dir, and if so, add all the files within.
		fi, err := os.Stat(v)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if fi.Mode().IsDir() {
			entries, err := os.ReadDir(v)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if args.Recursive {
				var subFiles []string
				for _, entry := range entries {
					subFiles = append(subFiles, filepath.Join(v, entry.Name()))
				}
				f.addFiles(subFiles)
			} else {
				for _, entry := range entries {
					f.files = append(f.files, &ImageFile{
						path: filepath.Join(v, entry.Name()),
					})
				}
			}
		} else {
			f.files = append(f.files, &ImageFile{
				path: v,
			})
		}
	}
}

func (f *ImageFiles) next() error {
	startIndex := f.currentIndex
	for {
		if f.currentIndex < len(f.files)-1 {
			f.currentIndex++
		} else {
			f.currentIndex = 0
		}
		// Couldn't find anything...
		if startIndex == f.currentIndex {
			return fmt.Errorf("no valid images found")
		}
		if args.Cache == false {
			f.files[f.currentIndex].unload()
		}
		if f.files[f.currentIndex].image == nil {
			f.files[f.currentIndex].load()
		}
		if f.files[f.currentIndex].invalid != nil {
			continue
		} else {
			f.currentFile = f.files[f.currentIndex]
			return nil
		}
	}
}

func (f *ImageFiles) prev() error {
	startIndex := f.currentIndex
	for {
		if f.currentIndex > 0 {
			f.currentIndex--
		} else {
			f.currentIndex = len(f.files) - 1
		}
		// Couldn't find anything...
		if startIndex == f.currentIndex {
			return fmt.Errorf("no valid images found")
		}
		if args.Cache == false {
			f.files[f.currentIndex].unload()
		}
		if f.files[f.currentIndex].image == nil {
			f.files[f.currentIndex].load()
		}
		if f.files[f.currentIndex].invalid != nil {
			continue
		} else {
			f.currentFile = f.files[f.currentIndex]
			imgOp = paint.ImageOp{}
			return nil
		}
	}
}
