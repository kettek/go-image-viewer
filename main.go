package main

import (
	"fmt"
	"image"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"gioui.org/font/gofont"
	"github.com/alexflint/go-arg"
	"golang.org/x/image/draw"
)

var args struct {
	Files []string `arg:"positional,required"`
	Zoom  float32  `arg:"-z,--zoom" help:"Image zoom" default:"1.0"`
	Fit   bool     `arg:"-f,--fit" help:"Fit images to window size" default:"false"`
	Ascii bool     `arg:"-a,--ascii" help:"Render files to ASCII" default:"false"`
}

var (
	list = &layout.List{
		Axis: layout.Vertical,
	}
	images            []ImageFile
	imgOp             paint.ImageOp
	currentImageIndex int
	refreshImage      bool
	// scaler ??? // draw.NearestNeighbor, draw.ApproxBiLinear, draw.BiLinear, draw.CatmullRom
)

func main() {
	arg.MustParse(&args)

	for _, v := range args.Files {
		images = append(images, ImageFile{
			path: v,
		})
	}

	if args.Ascii {
		for _, i := range images {
			s, err := i.asAscii()
			if err != nil {
				fmt.Errorf("couldn't open file %s: %w)", s, err)
			} else {
				fmt.Printf("%s (%s)\n", i.path, i.format)
				fmt.Printf("%s", s)
			}
		}
		return
	}

	go func() {
		w := app.NewWindow(app.Title("go-image-viewer"))
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

type (
	D = layout.Dimensions
	C = layout.Context
)

func nextImage() (err error) {
	if currentImageIndex < len(images)-1 {
		currentImageIndex++
	} else {
		currentImageIndex = 0
	}
	if images[currentImageIndex].image == nil {
		images[currentImageIndex].load()
	}
	if images[currentImageIndex].invalid == nil {
		imgOp = paint.ImageOp{}
	}
	return
}
func prevImage() (err error) {
	if currentImageIndex > 0 {
		currentImageIndex--
	} else {
		currentImageIndex = len(images) - 1
	}
	if images[currentImageIndex].image == nil {
		images[currentImageIndex].load()
	}
	if images[currentImageIndex].invalid == nil {
		imgOp = paint.ImageOp{}
	}
	return
}

func loop(w *app.Window) (err error) {
	th := material.NewTheme(gofont.Collection())
	var ops op.Ops
	images[currentImageIndex].load()
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			render(gtx, th)
			e.Frame(gtx.Ops)
		case key.Event:
			if e.State == key.Press {
				switch e.Name {
				case key.NameEscape:
					os.Exit(0)
				case key.NameLeftArrow:
					prevImage()
					w.Invalidate()
				case key.NameRightArrow:
					nextImage()
					w.Invalidate()
				case "z":
					args.Fit = !args.Fit
					imgOp = paint.ImageOp{}
				case "-":
					if args.Zoom > 1 {
						args.Zoom -= 1
					}
					imgOp = paint.ImageOp{}
				case "+":
					args.Zoom += 1
					imgOp = paint.ImageOp{}
				}
			}
		}
	}
}

func render(gtx C, th *material.Theme) D {
	widgets := []layout.Widget{
		func(gtx C) D {
			sz := gtx.Constraints.Min.X
			if imgOp.Size().X == 0 {
				currentImage := images[currentImageIndex].image
				rect := currentImage.Bounds()
				if args.Fit {
					// TODO
					/*sz := gtx.Constraints
					if rect.Max.X < sz.Max.X {
					}*/
				} else {
					rect.Max.X = int(float32(rect.Max.X) * args.Zoom)
					rect.Max.Y = int(float32(rect.Max.Y) * args.Zoom)
				}
				m := image.NewRGBA(rect)
				draw.NearestNeighbor.Scale(m, m.Bounds(), currentImage, currentImage.Bounds(), draw.Src, nil)
				imgOp = paint.NewImageOp(m)
			}
			img := widget.Image{Src: imgOp}
			img.Scale = float32(sz) / float32(gtx.Px(unit.Dp(float32(sz))))
			return img.Layout(gtx)
		},
	}
	return list.Layout(gtx, len(widgets), func(gtx C, i int) D {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
	})
}
