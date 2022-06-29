package cmd

import (
	"errors"
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
var bytes []byte

func Execute() {
	dumpPath = rootCmd.Flags().StringP("input", "i", "", "path to dump file")
	outputPath = rootCmd.Flags().StringP("output", "o", ".", "output path")
	rawAlpha = rootCmd.Flags().BoolP("raw-alpha", "r", false, "raw green colour instead of transparency")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
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

	sprites := make([]*image.Image, 0)
	for i, idx := 0, 0; i < len(bytes); i++ {
		img := scanImage(idx)
		if img != nil {
			idx += img.Size

			img.DecodePalette()
			err = img.DecodeImage()
			if err == nil {
				sprites = append(sprites, img)
			}
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
		idx := strconv.Itoa(i)
		var filename = pads[len(idx):] + idx + ".png"

		path := *outputPath + string(os.PathSeparator) + filename
		log.Println(path)
		if err := s.DrawImage(path, *rawAlpha); err != nil {
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
