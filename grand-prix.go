package main

import (
	"image/png"
	"io"
	"os"
)

var grid [][]bool

func main() {
	//initialize the grid which will be the playable space
	grid = make([][]bool, 50)
	for i := range grid {
		grid[i] = make([]bool, 100)
	}
	//load the image of the course and read the pixels to make the grid
	gridImageFile, err := os.Open("grid.png")
	if err != nil {
		println("error reading image")
		os.Exit(1)
	}

	setupGrid(gridImageFile)

	/*for i := range grid {
		for j := range grid[i] {
			grid[i][j] = false
			print(grid[i][j])
		}
		println("")
	}*/

}

// Get the bi-dimensional pixel array
func setupGrid(gridImageFile io.Reader) {
	gridImage, err := png.Decode(gridImageFile)
	if err != nil {
		println("error decoding image")
		os.Exit(2)
	}

	bounds := gridImage.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			_, _, _, a := gridImage.At(x, y).RGBA()
			if a == 65535 {
				grid[y][x] = true
				print("X")
			} else {
				grid[y][x] = false
				print(" ")
			}

		}
		println("")
	}
}
