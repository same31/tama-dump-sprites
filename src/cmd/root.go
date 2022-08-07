package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"tama-dump-sprites/src/image"
)

var rootCmd = &cobra.Command{
	Short: "sprite extract",
	Long:  "Command line utility to extract sprites from a Tamagotchi dump",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ExtractSprites(); err != nil {
			log.Fatalln(err)
		}
	},
}

var dumpPath, outputPath *string
var rawAlpha *bool
var filenameAddress *bool
var bytes []byte

func Execute() {
	dumpPath = rootCmd.Flags().StringP("input", "i", "", "path to dump file")
	outputPath = rootCmd.Flags().StringP("output", "o", ".", "output path")
	rawAlpha = rootCmd.Flags().BoolP("raw-alpha", "r", false, "raw green colour instead of transparency")
	filenameAddress = rootCmd.Flags().BoolP("address", "a", false, "output files contain address of image in file")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

type Sprite struct {
	Address int
	Size    int
	Img     *image.Image
}

func ExtractSprites() error {
	if len(*outputPath) <= 0 {
		var p = "."
		outputPath = &p
	} else if (*outputPath)[len(*outputPath)-1] == os.PathSeparator {
		var p = (*outputPath)[0 : len(*outputPath)-1]
		outputPath = &p
	}

	info, err := os.Stat(*outputPath)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return errors.New((*outputPath) + " is not a directory")
	}

	bytes, err = ioutil.ReadFile(*dumpPath)
	if err != nil {
		return err
	}

	sprites := make([]Sprite, 0)
	for i, idx := 0, 0; i < len(bytes); i++ {
		img := scanImage(idx)
		if img != nil {
			img.DecodePalette()
			err = img.DecodeImage()
			if err == nil {
				sprite := Sprite{
					Address: idx,
					Img:     img,
				}
				sprites = append(sprites, sprite)
			}

			idx += img.Size
		} else {
			idx++
		}
	}

	charLen := len(strconv.Itoa(len(sprites)))
	pads := ""
	for i := 0; i < charLen; i++ {
		pads += "0"
	}

	for i, s := range sprites {
		var prefix = strconv.Itoa(i)
		prefix = pads[len(prefix):] + prefix
		if *filenameAddress {
			prefix += fmt.Sprintf("_0x%x-0x%x", s.Address, s.Address+s.Img.Size-1)
		}

		var filename = prefix + ".png"

		path := *outputPath + string(os.PathSeparator) + filename
		log.Println(path)
		if err := s.Img.DrawImage(path, *rawAlpha); err != nil {
			return err
		}
	}

	return nil
}

func scanImage(offset int) *image.Image {
	if len(bytes)-1 < offset+2 {
		return nil
	}
	var width = int(bytes[offset])
	var height = int(bytes[offset+1])
	var paletteSize = int(bytes[offset+2])

	if len(bytes)-offset > 10 && width > 0 && width <= 128 && height > 0 && height <= 128 && paletteSize > 0 && bytes[offset+3] == 0 && bytes[offset+4] == 1 && bytes[offset+5] == 255 {
		var headerSize = 6 + paletteSize*2
		var pixelPerByte int
		if paletteSize > 16 {
			pixelPerByte = 1
		} else {
			pixelPerByte = 2
		}

		// less than 16 colors per pixel encoded using 4 bits, so 2 pixels encoded by 1 byte
		var size = headerSize + int(math.Ceil(float64((width*height)/pixelPerByte)))
		return image.New(offset, bytes[offset:(offset+size)])
	}
	return nil
}
