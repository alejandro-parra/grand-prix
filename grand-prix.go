package main

import "time"

var grid [][]string

type Location struct {
	Carril int
	Pos    int
}
type Car struct {
	responseChan chan bool
	id           int
}

func main() {
	lugaresDeInicio := [16]int{1, 1, 1, 1, 1, 1, 1, 1, 4, 4, 4, 4, 4, 4, 4, 4}
	numCarros := 4
	numVueltas := 3
	updateCh := make(chan Location)
	//initialize the grid which will be the playable space
	listChans = make(map[chan Location]Car)
	//definir rango de randoms de aceleracion y max speed

	for i := 1; i < numCarros+1; i++ {
		tmpResponseChan := make(chan bool)
		tmpRequestChan := make(chan Location)
		listChans[tmpRequestChan] = Car{tmpResponseChan, i}
		go carro(i, Location{i, lugaresDeInicio[i]}, 0.1)
	}

	grid = make([][]string, 8)
	for i := range grid {
		grid[i] = make([]string, 100)
	}
	for i := range grid {
		for j := range grid[i] {
			grid[i][j] = " "
		}
	}
	printGrid()

}

func printGrid() {
	for i := range grid {
		for j := range grid[i] {
			print(grid[i][j])
		}
		println("")
		println("-------------------------------------------------------------------------------------------------------------------------------------------------------------")
	}
}
func carro(id int, initLocation Location, maxSpeed float32, acceleration float32, chanRequest chan Location, response chan bool) {
	//sleep
	start = time.Now()
	lap := 0
	currentLocation := initLocation
	currentVelocity := 0
	currentAcceleration := acceleration
	currentVelocity += acceleration
	desacceleration := 0
	sleep := time.Millisecond * 1000
	for {
		if velocity < maxSpeed {
			time.Sleep(sleep - velocity)
		} else {
			time.Sleep(sleep - maxSpeed)
		}

		//el carro se duerme

		/*el carro analiza el grid

		- si en los proximos 10mts hay un obstaculo:
			- si es un corredor, moverse a un lado
			- si es pared o si no se puede mover al lado, empezar a desacelerar

			- definir nextLocation
			- definir nextAcc, el valor que se actualizara al final

		el carro le pide al main que lo mueva
			- el main regresa su decision
			- si el main dijo que si, actualiza valores de acceleration, velocity, y pocision

		si no, recalcula y vuelve a solicitar movimiento
		*/
		velocity += acceleration
	}
}
