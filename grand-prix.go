package main

import (
	"math/rand"
	"os"
	"strconv"
	"time"
)

type racer chan<- Location

var (
	track       [][]string
	competitors = make(map[int]chan bool)
	requests    = make(chan Location)
	totalLaps   int
)

//Update : struct that contains essential elements for the broadcasters
type Update struct {
	id       int
	rail     int
	position int
	lap      int
}

//Location : struct that contains essential elements for the broadcasters
type Location struct {
	id         int
	rail       int
	position   int
	currentLap int
}

func main() {
	winners := []int{1, 2, 3}
	winners = winners[:0]
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
	initialPositions := [16]int{0, 0, 0, 0, 0, 0, 0, 0, 3, 3, 3, 3, 3, 3, 3, 3}

	// go run gran-prix.go -racers 16 - laps 20
	if len(args) == 4 {
		numOfRacers, _ = strconv.Atoi(os.Args[1])
		totalLaps, _ = strconv.Atoi(os.Args[3])
	}

	for i := 1; i < numOfRacers+1; i++ {
		tmpResponseChan := make(chan bool)
		competitors[i] = tmpResponseChan
		tmpMaxSpeed := float64(rand.Intn(900-400) + 400)
		tmpAcceleration := float64(rand.Intn(100-40) + 40)
		go racerDynamics(Location{i, i, initialPositions[i], 1}, tmpMaxSpeed, tmpAcceleration, requests, tmpResponseChan)
	}

	for {
		recievedRequest := <-requests
		if track[recievedRequest.rail][recievedRequest.position] == " " {
			track[recievedRequest.rail][recievedRequest.position] = strconv.Itoa(recievedRequest.id)
			competitors[recievedRequest.id] <- true
			if recievedRequest.currentLap == totalLaps && recievedRequest.position == 0 {
				winners = append(winners, recievedRequest.id)
				if len(winners) == 3 {
					break
					println(winners)

				}
			}
		} else {
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
func racerDynamics(initLocation Location, maxSpeed float64, acceleration float64, chanRequest chan Location, response chan bool) {
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

	start := time.Now() //tiempo al inicio de la carrera
	id := initLocation.id
	lap := initLocation.currentLap
	currentLocation := initLocation
	currentVelocity := 0.0
	currentVelocity += acceleration
	desaccelerationRacer := -400.0
	desaccelerationCurve := -100.0
	sleep := 1000.0

	nextLocation := Location{0, 0, 0, 0}
	nextAcceleration := 0.0
	lap := initLocation.currentLap
	for lap < totalLaps { //mientras el coche no haya terminado la carrera
		if currentVelocity < maxSpeed { //si el carro no ha llegado a su limite de velocidad
			time.Sleep(time.Duration(sleep-currentVelocity) * time.Millisecond)
		} else { //si el carro ya llego a su limite de velocidad
			time.Sleep(time.Duration(sleep-maxSpeed) * time.Millisecond)
		}
		for {
			firstThreat := false
			//se checan las siguientes 5 posiciones en busqueda de carros que estorben
			for i := currentLocation.position + 1; i < currentLocation.position+5; i++ {
				if track[currentLocation.rail][i] != " " { //si en esta posición hay un carro
					firstThreat = true
					//ver si se puede mover a los lados y rebasar el otro carro
					if currentLocation.rail == 0 {
						if track[currentLocation.rail+1][currentLocation.position] == " " {
							nextLocation = Location{id, currentLocation.rail + 1, currentLocation.position, lap}
							nextAcceleration = acceleration
						} else {
							nextLocation = Location{id, currentLocation.rail, currentLocation.position + 1, lap}
							nextAcceleration = desaccelerationRacer
						}
					} else if currentLocation.rail == 7 {
						if track[currentLocation.rail-1][currentLocation.position] == " " {
							nextLocation = Location{id, currentLocation.rail - 1, currentLocation.position, lap}
							nextAcceleration = acceleration
						} else {
							nextLocation = Location{id, currentLocation.rail, currentLocation.position + 1, lap}
							nextAcceleration = desaccelerationRacer
						}
					} else {
						if track[currentLocation.rail+1][currentLocation.position] == " " {
							nextLocation := Location{id, currentLocation.rail + 1, currentLocation.position, lap}
							nextAcceleration = acceleration
						} else if track[currentLocation.rail-1][currentLocation.position] == " " {
							nextLocation = Location{id, currentLocation.rail - 1, currentLocation.position, lap}
							nextAcceleration = acceleration
						} else {
							nextLocation = Location{id, currentLocation.rail, currentLocation.position + 1, lap}
							nextAcceleration = desaccelerationRacer
						}
					}
				}
				if firstThreat {
					break
				}
			}
			if firstThreat == false { //si no hay nada estorbando adelante, pone su siguiete ubicación recto.
				nextLocation = Location{id, currentLocation.rail, currentLocation.position + 1, lap}
				nextAcceleration = acceleration
			}
			//si el carro se encuentra en una zona de frenado (curvas)
			if currentLocation.position >= 40 || currentLocation.position <= 50 || currentLocation.position >= 80 || currentLocation.position <= 90 || currentLocation.position >= 120 || currentLocation.position <= 130 || currentLocation.position >= 160 || currentLocation.position <= 170 {
				nextAcceleration = desaccelerationCurve
			}
			chanRequest <- nextLocation

			if <-response == true {
				break
			}
		}
		if nextLocation.position == 100 {
			t := time.Now()
			elapsed := t.Sub(start)
			nextLocation.position = 0
			lap++
		}
		currentLocation = nextLocation
		if nextAcceleration > 0 {
			if currentVelocity < maxSpeed {
				currentVelocity += nextAcceleration
			}
		} else {
			currentVelocity += nextAcceleration
		}
	}
}
