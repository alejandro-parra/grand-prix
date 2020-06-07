package main

import (
	"image/png"
	"io"
	"os"
	"os/exec"
	"runtime"
)

var grid [][]string
var clear map[string]func() //create a map for storing clear funcs

func init() {
	clear = make(map[string]func()) //Initialize it
	clear["linux"] = func() {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func main() {

	//initialize the grid which will be the playable space
	grid = make([][]string, 50)
	for i := range grid {
		grid[i] = make([]string, 100)
	}
	//load the image of the course and read the pixels to make the grid
	gridImageFile, err := os.Open("grid.png")
	if err != nil {
		println("error reading image")
		os.Exit(1)
	}

	setupGrid(gridImageFile)
	grid[20][25] = "O"
	printGrid()
	for i := 20; i < 35; i++ {
		grid[i][25] = " "
		grid[i+1][25] = "O"
		CallClear()
		printGrid()
	}
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
				grid[y][x] = "x"
			} else {
				grid[y][x] = " "
			}

		}
	}
}

func printGrid() {
	for i := range grid {
		for j := range grid[i] {
			print(grid[i][j])
		}
		println("")
	}
}

func CallClear() {
	value, ok := clear[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok {                          //if we defined a clear func for that platform:
		value() //we execute it
	} else { //unsupported platform
		panic("Your platform is unsupported! I can't clear terminal screen :(")
	}
}
