package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"golang.org/x/image/draw"
)

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func main() {
	startTime := time.Now()

	filePath := os.Args[1]

	file, _ := os.Open(filePath)

	var im image.Image

	if strings.HasSuffix(filePath, "jpg") {
		im, _ = jpeg.Decode(file)
	} else {
		im, _ = png.Decode(file)
	}

	newBounds := im.Bounds()
	minDimension := int(math.Min(float64(newBounds.Max.X), float64(newBounds.Max.Y)))
	if minDimension < 1000 {
		multiple := 1000.0 / float32(minDimension)
		newBounds = image.Rect(0, 0, int(float32(newBounds.Max.X)*multiple), int(float32(newBounds.Max.Y)*multiple))
	}

	fmt.Printf(newBounds.String())

	inputImage := image.NewRGBA(newBounds)

	draw.NearestNeighbor.Scale(inputImage, inputImage.Bounds(), im, im.Bounds(), draw.Over, nil)

	imageBounds := inputImage.Bounds()

	width := imageBounds.Max.X
	height := imageBounds.Max.Y

	minCircle := int(math.Min(float64(width), float64(height))/300) + 1

	outputImage := image.NewRGBA(inputImage.Bounds())

	limit := width * height / 200

	for i := 0; i < limit; {
		if i%100 == 0 {
			fmt.Printf("%d%%\n", i*100/limit)
		}

		rx := rand.Intn(width)
		ry := rand.Intn(height)

		pickedCol := inputImage.RGBAAt(rx, ry)

		// base colour
		// start with a radius of 2, then increase from there
		// determine average difference of all colours within radius to initial colour and stop at a certain cut off

		radius := minCircle
		var averageDifference float32 = 0.0

		calculateAverageDifference := func(x int, y int, r int) float32 {
			diffSum := 1
			diffCount := 0

			for xx := x - r; xx <= x+r; xx++ {
				for yy := y - r; yy <= y+r; yy++ {
					if (x-xx)*(x-xx)+(y-yy)*(y-yy) < r*r {
						col := inputImage.RGBAAt(xx, yy)

						diffr := Abs(int(pickedCol.R - col.R))
						diffg := Abs(int(pickedCol.G - col.G))
						diffb := Abs(int(pickedCol.B - col.B))
						diffSum += diffr + diffg + diffb
						diffCount += 3
					}
				}
			}

			return float32(diffSum) / float32(diffCount)
		}

		fillCircle := func(x int, y int, r int) {
			for xx := x - r; xx <= x+r; xx++ {
				for yy := y - r; yy <= y+r; yy++ {
					if (x-xx)*(x-xx)+(y-yy)*(y-yy) < r*r {
						outputImage.SetRGBA(xx, yy, pickedCol)
					}
				}
			}
		}

		dr := minCircle

		for {
			averageDifference = calculateAverageDifference(rx, ry, radius+dr)
			if averageDifference > 30 {
				break
			}
			radius += dr
		}

		fillCircle(rx, ry, radius)

		if radius > 10 {
			// draw in circle
			i++
		}
	}

	elapsed := time.Since(startTime)
	fmt.Printf("Generation took %s", elapsed)

	f, _ := os.Create(fmt.Sprintf("%d.png", time.Now().Unix()))
	png.Encode(f, outputImage)
	f.Close()
}
