package image

import (
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

type Image struct {
	Offset      int
	Size        int
	Bytes       []byte
	Byte5       byte
	Byte6       byte
	Byte7       byte
	Width       int
	Height      int
	ColorsCount int
	PaletteData []byte
	Palette     [][]int
	ImageData   []byte
	Pixels      []int
}

const headerSize = 5

func New(offset int, bytes []byte) *Image {
	var img = Image{
		Offset:      offset,
		Size:        len(bytes),
		Bytes:       bytes,
		Byte5:       bytes[3],
		Byte6:       bytes[4],
		Byte7:       bytes[5],
		Width:       int(bytes[0]),
		Height:      int(bytes[1]),
		ColorsCount: int(bytes[2]),
	}

	var paletteSize = img.ColorsCount * 2

	img.PaletteData = bytes[(headerSize + 1):(headerSize + paletteSize + 1)]
	img.ImageData = bytes[(headerSize + paletteSize + 1):]

	return &img
}

func (img *Image) DecodePalette() {
	var bytes = img.PaletteData

	palette := make([][]int, 0)

	// each color encoded by 16 bit big-endian word in order: 5 blue, 6 green, 5 red
	// a palette decoding was spied on the @ianling fork
	// https://github.com/ianling/t-on/blob/42a01b3d80fb89c8a4227d0dc7d33225a7d6505b/extract.py#L90-L104
	// which was inspired (as noted) by MyMeets https://tamatown.com/downloads
	for i := 0; i < len(bytes); i += 2 {
		var color16 = binary.BigEndian.Uint16(bytes[i : i+2])

		var blue = int(math.Round(float64(((color16&0xF800)>>11)*255) / float64(31)))
		var green = int(math.Round(float64(((color16&0x7E0)>>5)*255) / float64(63)))
		var red = int(math.Round(float64((color16&0x1F)*255) / float64(31)))

		c := []int{red, green, blue, 255}
		palette = append(palette, c)
	}

	img.Palette = palette
}

func (img *Image) DecodeImage() error {
	var bytes = img.ImageData
	var palette = img.Palette
	var halfBytePixel = len(palette) <= 16

	var pixelsCount = len(bytes)
	if halfBytePixel {
		pixelsCount *= 2
	}

	var pixels = make([]int, pixelsCount*4)

	if halfBytePixel {
		for i, b := range bytes {
			var idx = i * 8
			var bound = b & 0xf
			if int(bound) > len(palette)-1 {
				return errors.New("cannot decode this image")
			}
			var p1 = palette[bound]
			for i, p := range p1 {
				pixels[idx+i] = p
			}

			bound = b >> 4
			if int(bound) > len(palette)-1 {
				return errors.New("cannot decode this image")
			}
			var p2 = palette[bound]
			for i, p := range p2 {
				pixels[idx+i+4] = p
			}
		}
	} else {
		for i, b := range bytes {
			var idx = i * 4
			var p1 = palette[b]
			for i, p := range p1 {
				pixels[idx+i] = p
			}
		}
	}

	img.Pixels = pixels
	return nil
}

func (img *Image) DrawImage(path string, rawAlpha bool) error {
	if path == "" {
		return errors.New("path cannot be empty")
	}

	f := image.NewRGBA(image.Rect(0, 0, img.Width, img.Height))

	var c = 0
	for y := 0; y < img.Height; y++ {
		for x := 0; x < img.Width; x++ {
			var r, g, b, a int
			if c+3 <= len(img.Pixels)-1 {
				r = img.Pixels[c]
				g = img.Pixels[c+1]
				b = img.Pixels[c+2]
				a = img.Pixels[c+3]

				if !rawAlpha && r == 0 && g == 255 && a == 255 {
					a = 0
				}
			}

			var p = color.RGBA{
				R: uint8(r),
				G: uint8(g),
				B: uint8(b),
				A: uint8(a),
			}

			f.Set(x, y, p)

			c += 4
		}
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	return png.Encode(out, f)
}
