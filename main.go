package main

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"
	"sync"
	"syscall/js"

	"github.com/disintegration/imaging"
)

var asciiChar = []byte("@ND8OZ$7I?+=~:,  ")

func main() {
	c := make(chan struct{}, 0)
	println("WASM Initialized")
	js.Global().Set("process", js.NewCallback(process))
	<-c
}

func process(args []js.Value) {
	if len(args) <= 0 || args[0].Length() == 0 {
		return
	}
	file := args[0].Index(0)
	mime := file.Get("type").String()
	fmt.Printf("new %s file '%s'\n", mime, file.Get("name").String())
	fileReader := js.Global().Get("FileReader").New()
	var wg sync.WaitGroup
	var reader io.Reader
	wg.Add(1)
	onload := js.NewCallback(func(args []js.Value) {
		defer wg.Done()
		data := fileReader.Get("result")
		// strip of data:image/png;base64,
		if index := strings.Index(data.String(), ","); index != -1 {
			reader = base64.NewDecoder(base64.StdEncoding, strings.NewReader(data.String()[index+1:]))
		}
		//out := js.Global().Get("document").Call("getElementById", "output")
		//out.Set("src", data)
	})
	fileReader.Set("onload", onload)
	fileReader.Call("readAsDataURL", args[0].Index(0))
	go func() {
		wg.Wait()
		if reader != nil {
			var img image.Image
			var err error
			switch mime {
			case "image/gif":
				img, err = gif.Decode(reader)
				if err != nil {
					panic(err)
				}
			case "image/png":
				img, err = png.Decode(reader)
				if err != nil {
					panic(err)
				}
			case "image/jpeg":
				img, err = jpeg.Decode(reader)
				if err != nil {
					panic(err)
				}
			default:
				panic("unsupported type: " + mime)
			}
			asciiArt(img)
		}
	}()
}

func resize(img image.Image, w int) image.Image {
	h := (img.Bounds().Max.Y * w * 10) / (img.Bounds().Max.X * 16)
	return imaging.Resize(img, w, h, imaging.Lanczos)
}

func asciiArt(img image.Image) {
	var rect = img.Bounds()
	var width = rect.Max.X
	var height = rect.Max.Y
	fmt.Printf("image demension: %dx%d\n", width, height)
	if width > 600 {
		fmt.Println("resize image")
		asciiArt(resize(img, 600))
		return
	}
	var document = js.Global().Get("document")
	var pre = document.Call("createElement", "pre")
	var lines = make([]string, height)
	var wg sync.WaitGroup
	var queue = make(chan int, height)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go asciiArtWorker(queue, &wg, &rect, img, &lines)
	}
	for i := 0; i < rect.Max.Y; i++ {
		queue <- i
	}
	close(queue)
	wg.Wait()
	pre.Set("innerHTML", strings.Join(lines, "\n"))
	document.Call("getElementById", "result").Set("innerHTML", "")
	document.Call("getElementById", "result").Call("appendChild", pre)
}

func asciiArtWorker(queue <-chan int, wg *sync.WaitGroup, rect *image.Rectangle, img image.Image, lines *[]string) {
	defer wg.Done()
	for i := range queue {
		var line string
		for j := 0; j < rect.Max.X; j++ {
			p := img.At(j, i)
			g := color.GrayModel.Convert(p)
			y, _, _, _ := g.RGBA()
			pos := int(y * 16 / 1 >> 16)
			line += string(asciiChar[pos])
		}
		(*lines)[i] = line
	}
}
