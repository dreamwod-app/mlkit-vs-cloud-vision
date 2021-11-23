package main

import (
	"bufio"
	vision "cloud.google.com/go/vision/apiv1"
	"context"
	"flag"
	"fmt"
	"github.com/llgcode/draw2d/draw2dimg"
	vision2 "google.golang.org/genproto/googleapis/cloud/vision/v1"
	image2 "image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

import _ "image/jpeg"

func main() {
	flag.Usage = func() {
		program := filepath.Base(os.Args[0])
		fmt.Printf("Usage: %s vision -image <path-to-image> -out <output-file-path> to do cloud vision on an image.\n", program)
		fmt.Printf("Usage: %s draw -image <path-to-image> -coords <path-to-coordindate-file> -out <output-file-path> to draw coordinates from a file.\n", program)
	}
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	visionCmd := flag.NewFlagSet("vision", flag.ExitOnError)
	var output string
	visionCmd.StringVar(&output, "out", "", "output file path")
	var image string
	visionCmd.StringVar(&image, "image", "", "image file path")

	drawCmd := flag.NewFlagSet("draw", flag.ExitOnError)
	drawCmd.StringVar(&output, "out", "", "output file path")
	drawCmd.StringVar(&image, "image", "", "image file path")
	var coordinates string
	drawCmd.StringVar(&coordinates, "coords", "", "coordinates file path")

	switch os.Args[1] {
	case "vision":
		err := visionCmd.Parse(flag.Args()[1:])

		if err != nil {
			log.Fatalln(err)
		}
	case "draw":
		err := drawCmd.Parse(flag.Args()[1:])

		if err != nil {
			log.Fatalln(err)
		}
	}

	if visionCmd.Parsed() {
		if image == "" {
			visionCmd.PrintDefaults()
			os.Exit(1)
		}

		if output == "" {
			visionCmd.PrintDefaults()
			os.Exit(1)
		}

		err := ExecuteCloudVision(image, output)

		if err != nil {
			log.Fatalln(err)
		}
	}

	if drawCmd.Parsed() {
		if image == "" {
			drawCmd.PrintDefaults()
			os.Exit(1)
		}

		if output == "" {
			drawCmd.PrintDefaults()
			os.Exit(1)
		}

		if coordinates == "" {
			drawCmd.PrintDefaults()
			os.Exit(1)
		}

		err := DrawCoordinates(image, coordinates, output)

		if err != nil {
			log.Fatalln(err)
		}

	}
}

// ExecuteCloudVision calls cloud vision to get the detected text areas.
func ExecuteCloudVision(filepath string, output string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	client, err := vision.NewImageAnnotatorClient(context.Background())
	if err != nil {
		return err
	}
	defer client.Close()

	image, err := vision.NewImageFromReader(file)
	if err != nil {
		return err
	}

	texts, err := client.DetectTexts(context.Background(), image, nil, 0)
	if err != nil {
		log.Fatalf("Failed to detect texts: %v", err)
	}

	var areas = make([][]*vision2.Vertex, 0)

	fmt.Println("text returned from cloud vision:")
	for _, text := range texts {
		fmt.Println(text.Description)

		areas = append(areas, text.BoundingPoly.GetVertices())
	}

	// Draw the areas with detected texts
	return DrawAreas(filepath, areas, output)
}

func DrawCoordinates(filename, coordinateFile, output string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var coordinates = make([][]*vision2.Vertex, 0)

	bytes, _ := ioutil.ReadFile(coordinateFile)

	for _, s := range strings.Split(string(bytes), "\n") {
		values := strings.Split(s, ",")

		var area = make([]*vision2.Vertex, 0)

		x1, _ := strconv.ParseFloat(values[0], 32)
		y1, _ := strconv.ParseFloat(values[1], 32)
		x2, _ := strconv.ParseFloat(values[2], 32)
		y2, _ := strconv.ParseFloat(values[3], 32)
		x3, _ := strconv.ParseFloat(values[4], 32)
		y3, _ := strconv.ParseFloat(values[5], 32)
		x4, _ := strconv.ParseFloat(values[6], 32)
		y4, _ := strconv.ParseFloat(values[7], 32)

		area = append(area, &vision2.Vertex{
			X: int32(x1),
			Y: int32(y1),
		})
		area = append(area, &vision2.Vertex{
			X: int32(x2),
			Y: int32(y2),
		})
		area = append(area, &vision2.Vertex{
			X: int32(x3),
			Y: int32(y3),
		})
		area = append(area, &vision2.Vertex{
			X: int32(x4),
			Y: int32(y4),
		})

		coordinates = append(coordinates, area)

	}

	return DrawAreas(filename, coordinates, output)
}

// DrawAreas draws rectangular areas on the image and save the result to output
func DrawAreas(filename string, areas [][]*vision2.Vertex, output string) error {
	file, err := os.Open(filename)

	if err != nil {
		return err
	}

	reader := bufio.NewReader(file)

	img, _, err := image2.Decode(reader)

	if err != nil {
		return err
	}

	b := img.Bounds()

	surf := image2.NewRGBA(image2.Rect(0, 0, b.Max.X-b.Min.X, b.Max.Y-b.Min.Y))

	draw.Draw(surf, surf.Bounds(), img, img.Bounds().Min, draw.Src)

	gc := draw2dimg.NewGraphicContext(surf)

	gc.SetStrokeColor(color.RGBA{R: 255, G: 0x44, B: 0x44, A: 0xff})
	gc.SetLineWidth(6)

	for _, area := range areas {
		for i, v := range area {
			if i == 0 {
				gc.MoveTo(float64(v.X), float64(v.Y))
			} else {
				gc.LineTo(float64(v.X), float64(v.Y))
			}
		}
		gc.Close()

	}
	gc.Stroke()

	return draw2dimg.SaveToPngFile(output, surf)
}
