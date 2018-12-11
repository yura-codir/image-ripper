package main

import (
	"fmt"
	"image"
	"bytes"
	"io/ioutil"
	"image/png"
	"bufio"
	"github.com/nfnt/resize"
	"os"
	"strings"
	"path/filepath"
	"encoding/json"
	"io"
	"image/jpeg"
	"errors"
)

func main() {
	fmt.Println("Image ripper. version 1.0")
	args := os.Args[1:]

	config := ReadConfig(GetConfigPath(args))
	for _, file := range config.Files {
		if IsImagePath(file) {
			err := Resize(file, config)
			if err != nil {
				fmt.Printf("Unable to resize %s: %s", file, err.Error())
				fmt.Println()
			}
		}
	}
}

func Resize(input string, config Config) error {
	srcScale := config.Sizes[config.DefaultSize]
	if srcScale == 0 {
		return errors.New(fmt.Sprintf("in %s scale must not be 0", config.DefaultSize))
	}
	fileName := input[strings.LastIndexFunc(
		input,
		func(r rune) bool {
			return r == filepath.Separator
		}) + 1:]

	img, err := ReadImage(input)
	if err != nil {
		return err
	}
	for size, scale := range config.Sizes {
		if scale == 0 {
			return errors.New(fmt.Sprintf("in %s scale must not be 0", size))
		}
		scaled := ScaleImage(img, scale/srcScale)
		format := string(config.Output)
		format = strings.Replace(format, "{size}", size, 1)
		format = strings.Replace(format, "{file}", fileName, 1)
		output := filepath.Join(format, fileName)
		err = SaveImage(output, scaled)
		if err != nil {
			return err
		}
	}
	return nil
}

func ReadFile(input string) ([]byte, error) {
	return ioutil.ReadFile(input)
}

func ReadImage(input string) (image.Image, error) {
	src, err := ReadFile(input)
	if err != nil {
		return nil, err
	}
	imageSrc, _, err := image.Decode(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}
	return imageSrc, nil
}

func SaveImage(output string, img image.Image) error {
	dir := filepath.Dir(output)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	err = Encode(file.Name(), writer, img)
	writer.Flush()
	if err != nil {
		return err
	}
	return nil
}

func ScaleImage(img image.Image, scale float32) (image.Image) {
	size := img.Bounds().Size()
	w := float32(size.X) * scale
	h := float32(size.Y) * scale
	return resize.Resize(uint(w), uint(h), img, resize.Lanczos3)
}

func ReadConfig(input *string) Config {
	config := Config{
		DefaultSize: "xxxhdpi",
		Output:      filepath.Join("res", "drawable"),
		Files:       os.Args[1:],
		Sizes: map[string]float32{
			"mdpi":    1,
			"hdpi":    1.5,
			"xhdpi":   2,
			"xxhdpi":  3,
			"xxxhdpi": 4,
		},
	}
	if input != nil {
		path := *input
		file, err := ReadFile(path)
		err = json.Unmarshal(file, &config)
		if err != nil {
			fmt.Printf("Unable to parse config %s: %s", path, err.Error())
		} else {
			fmt.Printf("Used config %s", path)
		}
	}
	return config
}

func GetConfigPath(args []string) *string {
	for _, arg := range args {
		if strings.HasSuffix(arg, ".config") {
			return &arg
		}
	}
	return nil
}

func IsImagePath(input string) bool {
	return strings.HasSuffix(input, ".png") ||
		strings.HasSuffix(input, ".jpg")

}

func Encode(filename string, w io.Writer, m image.Image) (error) {
	var err error
	if strings.HasSuffix(filename, "jpg") {
		err = jpeg.Encode(w, m, &jpeg.Options{Quality: 100})
	} else if strings.HasSuffix(filename, "png") {
		err = png.Encode(w, m)
	}
	return err
}

type Config struct {
	DefaultSize string             `json:"default_size"`
	Output      string             `json:"output"`
	Files       []string           `json:"files"`
	Sizes       map[string]float32 `json:"sizes"`
}
