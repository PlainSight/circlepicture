package main

import (
	"fmt"
	"image"
	"image/color"
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

	//minCircle := int(math.Min(float64(width), float64(height))/300) + 1

	outputImage := image.NewRGBA(inputImage.Bounds())
	zImage := image.NewGray16(inputImage.Bounds())

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

		//radius := minCircle
		//var averageDifference float32 = 0.0

		// calculateAverageDifference := func(x int, y int, r int) float32 {
		// 	diffSum := 1
		// 	diffCount := 0

		// 	for xx := x - r; xx <= x+r; xx++ {
		// 		for yy := y - r; yy <= y+r; yy++ {
		// 			if (x-xx)*(x-xx)+(y-yy)*(y-yy) <= r*r {
		// 				col := inputImage.RGBAAt(xx, yy)

		// 				diffr := Abs(int(pickedCol.R - col.R))
		// 				diffg := Abs(int(pickedCol.G - col.G))
		// 				diffb := Abs(int(pickedCol.B - col.B))
		// 				diffSum += diffr + diffg + diffb
		// 				diffCount += 3
		// 			}
		// 		}
		// 	}

		// 	return float32(diffSum) / float32(diffCount)
		// }

		fillCircle := func(x int, y int, r int) {
			for xx := x - r; xx <= x+r; xx++ {
				for yy := y - r; yy <= y+r; yy++ {
					if (x-xx)*(x-xx)+(y-yy)*(y-yy) <= r*r {
						existingRadius := int(zImage.Gray16At(xx, yy).Y)
						if r < existingRadius || existingRadius == 0 {
							outputImage.SetRGBA(xx, yy, pickedCol)
							zImage.SetGray16(xx, yy, color.Gray16{Y: uint16(r)})
						}
					}
				}
			}
		}

		colourDiff := func(c1 color.RGBA, c2 color.RGBA) int {
			diffr := Abs(int(c1.R - c2.R))
			diffg := Abs(int(c1.G - c2.G))
			diffb := Abs(int(c1.B - c2.B))
			return diffr + diffg + diffb
		}

		type vector struct {
			x int
			y int
		}

		calculateRadius := func(x int, y int) int {
			var stack []vector
			visited := make(map[vector]struct{})

			stack = append(stack, vector{x: x, y: y})

			count := 0

			for len(stack) > 0 {
				popped := stack[len(stack)-1]
				stack = stack[:len(stack)-1]

				_, exists := visited[popped]

				if exists || popped.x < 0 && popped.x >= width && popped.y < 0 || popped.y >= height || count > 10000 {
					continue
				}

				count++

				visited[popped] = struct{}{}

				if colourDiff(inputImage.RGBAAt(popped.x, popped.y), pickedCol) < 60 {
					stack = append(stack,
						vector{x: popped.x + 1, y: popped.y},
						vector{x: popped.x - 1, y: popped.y},
						vector{x: popped.x, y: popped.y + 1},
						vector{x: popped.x, y: popped.y - 1})
				}
			}

			return int(math.Sqrt(float64(count)))
		}

		// dr := minCircle

		// for {
		// 	averageDifference = calculateAverageDifference(rx, ry, radius+dr)
		// 	if averageDifference > 30 {
		// 		break
		// 	}
		// 	radius += dr
		// }

		fillCircle(rx, ry, calculateRadius(rx, ry))

		//if radius > 10 {
		// draw in circle
		i++
		//}
	}

	elapsed := time.Since(startTime)
	fmt.Printf("Generation took %s", elapsed)

	f, _ := os.Create(fmt.Sprintf("%d.png", time.Now().Unix()))
	png.Encode(f, outputImage)
	f.Close()
}
