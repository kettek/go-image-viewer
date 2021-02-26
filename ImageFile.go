package main

import (
	"image"
	"os"

	"github.com/qeesung/image2ascii/convert"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "github.com/kettek/apng"
	_ "github.com/kettek/xbm"
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

func (i *ImageFile) asAscii() (string, error) {
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
