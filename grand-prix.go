package main

import (
	"time"
	"math/rand"
	"os"
	"strconv"
)

type racer chan <-Location

var (
	track [][]string
	competitors = make(map[int]chan bool)
	requests =  make(chan Location)
	totalLaps int
)


type Location struct {
	id int
	rail int
	position int
	currentLap int
}

func main() {
	var winners []int
	track = make([][]string, 8)

	for i := range track {
		track[i] = make([]string, 200)
	}

	for i := range track {
		for j := range track[i] {
			track[i][j] = " "
		}
	}

	args := os.Args[1:]
	numOfRacers := 4
	totalLaps = 3
	initialPositions := [16]int{1, 1, 1, 1, 1, 1, 1, 1, 4, 4, 4, 4, 4, 4, 4, 4}

	// go run gran-prix.go -racers 16 - laps 20
	if len(args) == 4 {
		numOfRacers = args[1]
		totalLaps = args[3]
	}

	for i := 1; i < racers + 1; i++ {
		tmpResponseChan := make(chan bool) 
		competitors[i] = tmpResponseChan
		tmpMaxSpeed := rand.Intn(900 - 400) + 400
		tmpAcceleration := rand.Intn(100 - 40) + 40
		go racerDynamics(i, Location{i, i, initialPositions[i], 1}, tmpMaxSpeed, tmpAcceleration, requests, tmpResponseChan)
	}

	for {
		recievedRequest := <- requests
		if track[recievedRequest.rail][recievedRequest.position] == " "{
			track[recievedRequest.rail][recievedRequest.position] = strconv.Itoa(recievedRequest.id)
			competitors[recievedRequest.id] <- true
			if recievedRequest.currentLap == totalLaps && recievedRequest.position == 0{
				append(winners, recievedRequest.id)
				if len(winners) == 3{
					break
				}
			}
		}
		else{
			competitors[recievedRequest.id] <- false
		}
	}
}

func printTrack() {
	for i := range track {
		for j := range track[i] {
			print(track[i][j])
		}
		println("")
		println("-------------------------------------------------------------------------------------------------------------------------------------------------------------")
	}
}
func racerDynamics(initLocation Location, maxSpeed float32, acceleration float32, chanRequest chan Location, response chan bool) {
	start = time.Now()
	id := initLocation.id
	

	currentLocation := initLocation
	currentVelocity := 0
	currentAcceleration := acceleration
	currentVelocity += acceleration
	desaccelerationRacer := - 400
	desaccelerationCurve := - 100
	
	sleep := 1000
	for lap := initLocation.currentLap; lap < totalLaps;{
		if currentVelocity < maxSpeed {
			time.Sleep((sleep - currentVelocity) * time.Millisecond)
		} else {
			time.Sleep((sleep - maxSpeed) * time.Millisecond)
		}

		for i := currentLocation.position + 1; i < currentLocation.position + 10; i++{
			if track[currentLocation.rail][i] != " "{
				
			}
		}
		// zonas de frenado {40, 80, 120, 160}

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
		currentVelocity += acceleration
	}
}
