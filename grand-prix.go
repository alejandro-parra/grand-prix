package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type racer chan<- Location

var (
	track         [][]string
	competitors   = make(map[int]chan bool) //the reference to each communication with the racers
	requests      = make(chan Location)		//a channel that all racers use to ask main to move
	destroy       = make(chan Location, 60)
	updateChan    = make(chan Update, 60)	//channel to provide the printing system each racer's stats
	totalLaps     int
	numOfRacers   int
	totalDistance int
	winners       []int
	clear         map[string]func() //create a map for storing clear funcs
)

//Update : struct that contains essential elements for the broadcasters
type Update struct {
	id         int
	rail       int
	position   int
	lap        int
	speed      float64
	lapTime    string
	racingTime string
	lastUpdate string
}

//Location : struct that contains essential elements for the broadcasters
type Location struct {
	id         int
	rail       int
	position   int
	currentLap int
}

//code from: https://stackoverflow.com/questions/22891644/how-can-i-clear-the-terminal-screen-in-go
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
func callClear() {
	value, ok := clear[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok {                          //if we defined a clear func for that platform:
		value() //we execute it
	} else { //unsupported platform
		panic("Your platform is unsupported! I can't clear terminal screen :(")
	}
}

func main() {
	totalDistance = 150
	winners = []int{1, 2, 3}
	winners = winners[:0]
	track = make([][]string, 8)

	//initialize empty track array
	for i := range track {
		track[i] = make([]string, totalDistance)
	}
	for i := range track {
		for j := range track[i] {
			track[i][j] = " "
		}
	}

	//args := os.Args[1:]
	numOfRacers = 4
	totalLaps = 3

	//array used to specify starting positions of racers at start of race
	initialPositions := [16]int{0, 0, 0, 0, 0, 0, 0, 0, 3, 3, 3, 3, 3, 3, 3, 3}

	// usage: go run gran-prix.go -racers 16 -laps 20
	if len(os.Args) == 5 || len(os.Args) == 1 {
		if len(os.Args) == 1 {
		} else if os.Args[1] == "-racers" && os.Args[3] == "-laps" {
			numOfRacers, _ = strconv.Atoi(os.Args[2])
			totalLaps, _ = strconv.Atoi(os.Args[4])
		} else {
			fmt.Println("Wrong parameters. Usage: go run grand-prix.go -racers X -laps X")
			os.Exit(1)
		}

	} else {
		fmt.Println("Wrong parameters. Usage: go run grand-prix.go -racers X -laps X")
		os.Exit(2)
	}

	//create random generator for velocity and acceleration of cars
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	//initialize racers
	for i := 1; i < numOfRacers+1; i++ {
		tmpResponseChan := make(chan bool)
		competitors[i] = tmpResponseChan
		tmpMaxSpeed := float64(r.Intn(700-500) + 500)
		tmpAcceleration := float64(r.Intn(150-75) + 75)
		go racerDynamics(Location{i, i % 8, initialPositions[i-1], 1}, tmpMaxSpeed, tmpAcceleration, requests, tmpResponseChan)
	}
	killPrint := make(chan struct{})
	go prints(killPrint)
	for {
		select {

		//a car sends a request to move to a certain location
		case recievedRequest := <-requests:
			if track[recievedRequest.rail][recievedRequest.position] == " " {
				track[recievedRequest.rail][recievedRequest.position] = strconv.Itoa(recievedRequest.id)
				//track[recievedRequest.rail][recievedRequest.position] = " "
				competitors[recievedRequest.id] <- true //change accepted
				//update track
				if recievedRequest.currentLap == totalLaps && recievedRequest.position == 0 { //declare winners
					//fmt.Println("WINNER WINNER CHICKEN DINNER", recievedRequest.id)
					winners = append(winners, recievedRequest.id)
					if numOfRacers < 3 {
						if len(winners) == numOfRacers {
							killPrint <- struct{}{}
							fmt.Println("Race is over!")
							fmt.Println("THE WINNERS ARE:")
							for i := 0; i < numOfRacers; i++ {
								fmt.Println(i, ") ", winners[i])
							}
							return
						}
					} else {
						if len(winners) == 3 {
							killPrint <- struct{}{}
							fmt.Println("Race is over!")
							fmt.Println("THE WINNERS ARE:")
							fmt.Println("1) ", winners[0])
							fmt.Println("2) ", winners[1])
							fmt.Println("3) ", winners[2])
							return
						}
					}
				}
			} else {
				competitors[recievedRequest.id] <- false
			}
		//a call to destroy an object from the track
		case recievedRequest := <-destroy:
			if track[recievedRequest.rail][recievedRequest.position] == strconv.Itoa(recievedRequest.id) {
				track[recievedRequest.rail][recievedRequest.position] = " "
			} else {
				fmt.Println("Error, destroy request", track[recievedRequest.rail][recievedRequest.position])
			}
		}

	}

}

//function to print the track at any moment in race
func printTrack() {
	fmt.Println("")
	breakzone := "| Curve  |"
	fmt.Println(strings.Repeat(" ", 23), breakzone, strings.Repeat(" ", 18), breakzone, strings.Repeat(" ", 23), breakzone, strings.Repeat(" ", 18), breakzone, strings.Repeat(" ", 18))
	for i := range track {
		print("|")
		for j := range track[i] {
			print(track[i][j])
		}
		fmt.Print("|")
		fmt.Println("")
		fmt.Println("|", strings.Repeat("-", totalDistance), "|")
	}
}
func racerDynamics(initLocation Location, maxSpeed float64, acceleration float64, chanRequest chan Location, response chan bool) {
	// zonas de frenado {25-35, 55-65, 90-100, 120-130}

	start := time.Now() //tiempo al inicio de la carrera
	startLap := time.Now()
	elapsed := startLap.Sub(start)
	id := initLocation.id
	currentLocation := initLocation
	currentVelocity := 0.0
	currentVelocity += acceleration
	desaccelerationRacer := -400.0
	r2 := rand.New(rand.NewSource(time.Now().UnixNano()))
	desaccelerationCurve := float64(-(r2.Intn(35-10) + 10))
	sleep := 800.0

	nextLocation := Location{0, 0, 0, 0}
	nextAcceleration := 0.0
	lastUpdateCar := ""
	lap := initLocation.currentLap
	for lap < totalLaps+1 { //mientras el coche no haya terminado la carrera
		if currentVelocity < maxSpeed { //si el carro no ha llegado a su limite de velocidad
			time.Sleep(time.Duration(sleep-currentVelocity) * time.Millisecond)
		} else { //si el carro ya llego a su limite de velocidad
			time.Sleep(time.Duration(sleep-maxSpeed) * time.Millisecond)
		}
		for {
			firstThreat := false
			indexOutOfRange := false
			//se checan las siguientes 5 posiciones en busqueda de carros que estorben
			afk := 0
			for i := (currentLocation.position + 1) % totalDistance; i != (currentLocation.position+5)%totalDistance; i = (i + 1) % totalDistance {
				afk++
				if track[currentLocation.rail][i] != " " { //si en esta posición hay un carro
					firstThreat = true
					lastUpdateCar = fmt.Sprintf("obstacle rail %d", currentLocation.rail)
					//fmt.Println("obstacle seen at rail", currentLocation.rail, afk, "spaces away")
					//ver si se puede mover a los lados y rebasar el otro carro
					if currentLocation.rail == 0 {
						if track[currentLocation.rail+1][(currentLocation.position+1)%totalDistance] == " " {
							nextLocation = Location{id, currentLocation.rail + 1, currentLocation.position + 1, lap}
							nextAcceleration = acceleration
							lastUpdateCar = fmt.Sprintf("i am at left")
							//fmt.Println(id, "i am at far left")
						} else {
							nextLocation = Location{id, currentLocation.rail, currentLocation.position + 1, lap}
							nextAcceleration = desaccelerationRacer
							lastUpdateCar = fmt.Sprintf("deaccelerating3")
							//fmt.Println("deaccelerating3")
						}
					} else if currentLocation.rail == 7 {
						if track[currentLocation.rail-1][(currentLocation.position+1)%totalDistance] == " " {
							nextLocation = Location{id, currentLocation.rail - 1, currentLocation.position + 1, lap}
							nextAcceleration = acceleration
							lastUpdateCar = fmt.Sprintf("i am at right")
							//fmt.Println(id, "i am at far right")
						} else {
							nextLocation = Location{id, currentLocation.rail, currentLocation.position + 1, lap}
							nextAcceleration = desaccelerationRacer
							lastUpdateCar = fmt.Sprintf("deaccelerating2")
							//fmt.Println("deaccelerating2")
						}
					} else {
						if track[currentLocation.rail+1][(currentLocation.position+1)%totalDistance] == " " {
							nextLocation = Location{id, currentLocation.rail + 1, currentLocation.position + 1, lap}
							nextAcceleration = acceleration
							lastUpdateCar = fmt.Sprintf("i went right")
							//fmt.Println(id, "i went to the right")
						} else if track[currentLocation.rail-1][(currentLocation.position+1)%totalDistance] == " " {
							nextLocation = Location{id, currentLocation.rail - 1, currentLocation.position + 1, lap}
							nextAcceleration = acceleration
							lastUpdateCar = fmt.Sprintf("i went left")
							//fmt.Println(id, "i went to the left")
						} else {
							nextLocation = Location{id, currentLocation.rail, currentLocation.position + 1, lap}
							nextAcceleration = desaccelerationRacer
							lastUpdateCar = fmt.Sprintf("deaccelerating")
							//fmt.Println("deaccelerating")
						}
					}
				}
				//si ya encontro algo adelante, no sigas buscando
				if firstThreat || indexOutOfRange {
					break
				}
			}
			if firstThreat == false { //si no hay nada estorbando adelante, pone su siguiete ubicación recto.
				nextLocation = Location{id, currentLocation.rail, currentLocation.position + 1, lap}
				nextAcceleration = acceleration
			}
			//si el carro se encuentra en una zona de frenado (curvas)
			if (currentLocation.position >= 25 && currentLocation.position <= 35) || (currentLocation.position >= 55 && currentLocation.position <= 65) || (currentLocation.position >= 90 && currentLocation.position <= 100) || (currentLocation.position >= 120 && currentLocation.position <= 130) {
				nextAcceleration = desaccelerationCurve
			}
			if nextLocation.position >= totalDistance {
				nextLocation.position = 0
				//fmt.Println(id, "my pos was total distance", nextLocation)
			}
			chanRequest <- nextLocation

			if <-response == true {
				destroy <- currentLocation
				break
			}
		}
		if nextLocation.position == 0 {
			lastUpdateCar = "completed lap"
			//fmt.Println(id, "a lap is done!", nextLocation)
			t := time.Now()
			elapsed = t.Sub(startLap)
			startLap = time.Now()
			lap++

		}
		currentLocation = nextLocation
		if nextAcceleration > 0 {
			if currentVelocity < maxSpeed {
				currentVelocity += nextAcceleration
			}
		} else {
			if currentVelocity >= 0 {
				if currentVelocity+nextAcceleration < 0 {
					currentVelocity = 0
				} else {
					currentVelocity += nextAcceleration
				}

			}
		}
		updateChan <- Update{id, currentLocation.rail, currentLocation.position, lap, currentVelocity, elapsed.String(), time.Now().Sub(start).String(), lastUpdateCar}
	}
}
func prints(killT chan struct{}) {
	start := time.Now()
	updateList := make([]Update, numOfRacers)
	numSpaces := 25
	info := [8]string{"Player ", "Rail: ", "Position: ", "Lap: ", "Speed: ", "Lap Time: ", "GlobalTime: ", "LastUpdate: "}
	for {
		space := 20
		if len(winners) >= 1 {
			space = 5
		} else if numOfRacers < 5 {
			space = 10
		}
		tmpUpdateList := make([]Update, space)
		for i := 0; i < len(tmpUpdateList); i++ {
			select {
			case x := <-updateChan:
				tmpUpdateList[i] = x
			case <-killT:
				return
			}
		}

		for i := 0; i < len(tmpUpdateList); i++ {
			tmpid := tmpUpdateList[i].id
			updateList[tmpid-1] = tmpUpdateList[i]
		}
		tmpString := ""
		callClear()

		if numOfRacers < 9 {
			for j := 0; j < numOfRacers; j++ {
				tmptmpstring := info[0] + strconv.Itoa(updateList[j].id)
				if len(tmptmpstring) < numSpaces {
					tmptmpstring += strings.Repeat(" ", (numSpaces - len(tmptmpstring)))
				}
				tmpString += tmptmpstring
			}
			fmt.Println(tmpString)
			tmpString = ""
			for j := 0; j < numOfRacers; j++ {
				tmptmpstring := info[1] + strconv.Itoa(updateList[j].rail)
				if len(tmptmpstring) < numSpaces {
					tmptmpstring += strings.Repeat(" ", (numSpaces - len(tmptmpstring)))
				}
				tmpString += tmptmpstring
			}
			fmt.Println(tmpString)
			tmpString = ""
			for j := 0; j < numOfRacers; j++ {
				tmptmpstring := info[2] + strconv.Itoa(updateList[j].position)
				if len(tmptmpstring) < numSpaces {
					tmptmpstring += strings.Repeat(" ", (numSpaces - len(tmptmpstring)))
				}
				tmpString += tmptmpstring
			}
			fmt.Println(tmpString)
			tmpString = ""
			for j := 0; j < numOfRacers; j++ {
				tmptmpstring := ""
				if updateList[j].lap == totalLaps+1 {
					tmptmpstring = info[3] + "Finished!"
				} else {
					tmptmpstring = info[3] + strconv.Itoa(updateList[j].lap)
				}
				if len(tmptmpstring) < numSpaces {
					tmptmpstring += strings.Repeat(" ", (numSpaces - len(tmptmpstring)))
				}
				tmpString += tmptmpstring
			}
			fmt.Println(tmpString)
			tmpString = ""
			for j := 0; j < numOfRacers; j++ {
				s := fmt.Sprintf("%f", updateList[j].speed)
				tmptmpstring := info[4] + s
				if len(tmptmpstring) < numSpaces {
					tmptmpstring += strings.Repeat(" ", (numSpaces - len(tmptmpstring)))
				}
				tmpString += tmptmpstring
			}
			fmt.Println(tmpString)
			tmpString = ""
			for j := 0; j < numOfRacers; j++ {
				tmptmpstring := info[5] + updateList[j].lapTime
				if len(tmptmpstring) < numSpaces {
					tmptmpstring += strings.Repeat(" ", (numSpaces - len(tmptmpstring)))
				}
				tmpString += tmptmpstring
			}
			fmt.Println(tmpString)
			tmpString = ""
			for j := 0; j < numOfRacers; j++ {
				tmptmpstring := info[7] + updateList[j].lastUpdate
				if len(tmptmpstring) < numSpaces {
					tmptmpstring += strings.Repeat(" ", (numSpaces - len(tmptmpstring)))
				}
				tmpString += tmptmpstring
			}
			fmt.Println(tmpString)
			fmt.Println("")
			fmt.Println("Total Time: ", time.Now().Sub(start).String())

		} else {
			for j := 0; j < 8; j++ {
				tmptmpstring := info[0] + strconv.Itoa(updateList[j].id)
				if len(tmptmpstring) < numSpaces {
					tmptmpstring += strings.Repeat(" ", (numSpaces - len(tmptmpstring)))
				}
				tmpString += tmptmpstring
			}
			fmt.Println(tmpString)
			tmpString = ""
			for j := 0; j < 8; j++ {
				tmptmpstring := info[1] + strconv.Itoa(updateList[j].rail)
				if len(tmptmpstring) < numSpaces {
					tmptmpstring += strings.Repeat(" ", (numSpaces - len(tmptmpstring)))
				}
				tmpString += tmptmpstring
			}
			fmt.Println(tmpString)
			tmpString = ""
			for j := 0; j < 8; j++ {
				tmptmpstring := info[2] + strconv.Itoa(updateList[j].position)
				if len(tmptmpstring) < numSpaces {
					tmptmpstring += strings.Repeat(" ", (numSpaces - len(tmptmpstring)))
				}
				tmpString += tmptmpstring
			}
			fmt.Println(tmpString)
			tmpString = ""
			for j := 0; j < 8; j++ {
				tmptmpstring := ""
				if updateList[j].lap == totalLaps+1 {
					tmptmpstring = info[3] + "Finished!"
				} else {
					tmptmpstring = info[3] + strconv.Itoa(updateList[j].lap)
				}
				if len(tmptmpstring) < numSpaces {
					tmptmpstring += strings.Repeat(" ", (numSpaces - len(tmptmpstring)))
				}
				tmpString += tmptmpstring
			}
			fmt.Println(tmpString)
			tmpString = ""
			for j := 0; j < 8; j++ {
				s := fmt.Sprintf("%f", updateList[j].speed)
				tmptmpstring := info[4] + s
				if len(tmptmpstring) < numSpaces {
					tmptmpstring += strings.Repeat(" ", (numSpaces - len(tmptmpstring)))
				}
				tmpString += tmptmpstring
			}
			fmt.Println(tmpString)
			tmpString = ""
			for j := 0; j < 8; j++ {
				tmptmpstring := info[5] + updateList[j].lapTime
				if len(tmptmpstring) < numSpaces {
					tmptmpstring += strings.Repeat(" ", (numSpaces - len(tmptmpstring)))
				}
				tmpString += tmptmpstring
			}
			fmt.Println(tmpString)
			tmpString = ""
			for j := 0; j < 8; j++ {
				tmptmpstring := info[7] + updateList[j].lastUpdate
				if len(tmptmpstring) < numSpaces {
					tmptmpstring += strings.Repeat(" ", (numSpaces - len(tmptmpstring)))
				}
				tmpString += tmptmpstring
			}
			fmt.Println(tmpString)
			fmt.Println("")
			tmpString = ""
			for j := 8; j < numOfRacers; j++ {
				tmptmpstring := info[0] + strconv.Itoa(updateList[j].id)
				if len(tmptmpstring) < numSpaces {
					tmptmpstring += strings.Repeat(" ", (numSpaces - len(tmptmpstring)))
				}
				tmpString += tmptmpstring
			}
			fmt.Println(tmpString)
			tmpString = ""
			for j := 8; j < numOfRacers; j++ {
				tmptmpstring := info[1] + strconv.Itoa(updateList[j].rail)
				if len(tmptmpstring) < numSpaces {
					tmptmpstring += strings.Repeat(" ", (numSpaces - len(tmptmpstring)))
				}
				tmpString += tmptmpstring
			}
			fmt.Println(tmpString)
			tmpString = ""
			for j := 8; j < numOfRacers; j++ {
				tmptmpstring := info[2] + strconv.Itoa(updateList[j].position)
				if len(tmptmpstring) < numSpaces {
					tmptmpstring += strings.Repeat(" ", (numSpaces - len(tmptmpstring)))
				}
				tmpString += tmptmpstring
			}
			fmt.Println(tmpString)
			tmpString = ""
			for j := 8; j < numOfRacers; j++ {
				tmptmpstring := ""
				if updateList[j].lap == totalLaps+1 {
					tmptmpstring = info[3] + "Finished!"
				} else {
					tmptmpstring = info[3] + strconv.Itoa(updateList[j].lap)
				}
				if len(tmptmpstring) < numSpaces {
					tmptmpstring += strings.Repeat(" ", (numSpaces - len(tmptmpstring)))
				}
				tmpString += tmptmpstring
			}
			fmt.Println(tmpString)
			tmpString = ""
			for j := 8; j < numOfRacers; j++ {
				s := fmt.Sprintf("%f", updateList[j].speed)
				tmptmpstring := info[4] + s
				if len(tmptmpstring) < numSpaces {
					tmptmpstring += strings.Repeat(" ", (numSpaces - len(tmptmpstring)))
				}
				tmpString += tmptmpstring
			}
			fmt.Println(tmpString)
			tmpString = ""
			for j := 8; j < numOfRacers; j++ {
				tmptmpstring := info[5] + updateList[j].lapTime
				if len(tmptmpstring) < numSpaces {
					tmptmpstring += strings.Repeat(" ", (numSpaces - len(tmptmpstring)))
				}
				tmpString += tmptmpstring
			}
			fmt.Println(tmpString)
			tmpString = ""
			for j := 8; j < numOfRacers; j++ {
				tmptmpstring := info[7] + updateList[j].lastUpdate
				if len(tmptmpstring) < numSpaces {
					tmptmpstring += strings.Repeat(" ", (numSpaces - len(tmptmpstring)))
				}
				tmpString += tmptmpstring
			}
			fmt.Println(tmpString)
			t := time.Now().Sub(start).String()
			fmt.Println("")
			fmt.Println("Total Time: ", t)
			fmt.Println("------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------")

		}
		printTrack()
	}

}
