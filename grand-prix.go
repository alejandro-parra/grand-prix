package main

import (
	"flag"
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
	track       [][]string
	requests    = make(chan Location)
	destroy     = make(chan Location, 60)
	updateChan  = make(chan Update, 60) //no pasa naaa
	totalLaps   int
	numOfRacers int
	winners     = [3]int{0, 0, 0}
	clear       map[string]func() //create a map for storing clear funcs
)

const (
	totalDistance = 100
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
}

//Location : struct that contains essential elements for the broadcasters
type Location struct {
	id         int
	rail       int
	position   int
	currentLap int
}

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
	//winners = winners[:0]
	track = make([][]string, 8)
	competitors := make(map[int]chan bool)
	initialPositions := [16]int{0, 0, 0, 0, 0, 0, 0, 0, 3, 3, 3, 3, 3, 3, 3, 3}
	nr := flag.Int("racers", 4, "number of racers!")
	nl := flag.Int("laps", 3, "number of laps!")
	flag.Parse()
	numOfRacers = *nr
	totalLaps = *nl

	for i := range track {
		track[i] = make([]string, totalDistance)
	}

	for i := range track {
		for j := range track[i] {
			track[i][j] = " "
		}
	}

	// go run gran-prix.go -racers 16 - laps 20
	/*if len(args) == 4 {
		numOfRacers, _ = strconv.Atoi(os.Args[1])
		totalLaps, _ = strconv.Atoi(os.Args[3])
	}*/

	for i := 1; i < numOfRacers+1; i++ {
		tmpResponseChan := make(chan bool)
		competitors[i] = tmpResponseChan
		tmpMaxSpeed := float64(rand.Intn(900-400) + 400)
		tmpAcceleration := float64(rand.Intn(100-40) + 40)
		go racerDynamics(Location{i, i, initialPositions[i], 1}, tmpMaxSpeed, tmpAcceleration, requests, tmpResponseChan)
	}

	//start := time.Now()
	go prints()
	c := 0
	for {
		select {
		case recievedRequest := <-requests:
			if track[recievedRequest.rail][recievedRequest.position] == " " {
				track[recievedRequest.rail][recievedRequest.position] = strconv.Itoa(recievedRequest.id)
				//track[recievedRequest.rail][recievedRequest.position] = " "
				competitors[recievedRequest.id] <- true //change accepted
				//update track
				if recievedRequest.currentLap == totalLaps && recievedRequest.position == 0 { //declare winners
					fmt.Println("WINNER WINNER CHICKEN DINNER", recievedRequest.id)
					winners[c] = recievedRequest.id
					c++
					if winners[2] != 0 {
						println("race is over")
						fmt.Println("THE WINNERS ARE:", winners)
						return
					}
				}
			} else {
				competitors[recievedRequest.id] <- false
			}
		case recievedRequest := <-destroy:
			if track[recievedRequest.rail][recievedRequest.position] == strconv.Itoa(recievedRequest.id) {
				track[recievedRequest.rail][recievedRequest.position] = " "
			} else {
				fmt.Println("UYYYYYYYYYYYYY", track[recievedRequest.rail][recievedRequest.position])
			}
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

	start := time.Now() //tiempo al inicio de la carrera
	startLap := time.Now()
	elapsed := startLap.Sub(start)
	id := initLocation.id
	currentLocation := initLocation
	currentVelocity := 0.0
	currentVelocity += acceleration
	desaccelerationRacer := -400.0
	desaccelerationCurve := -100.0
	sleep := 1000.0

	nextLocation := Location{0, 0, 0, 0}
	nextAcceleration := 0.0
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
					fmt.Println("obstacle seen at rail", currentLocation.rail, afk, "spaces away")
					//ver si se puede mover a los lados y rebasar el otro carro
					if currentLocation.rail == 0 {
						if track[currentLocation.rail+1][(currentLocation.position+1)%totalDistance] == " " {
							nextLocation = Location{id, currentLocation.rail + 1, currentLocation.position + 1, lap}
							nextAcceleration = acceleration
							fmt.Println(id, "i am at far left")
						} else {
							nextLocation = Location{id, currentLocation.rail, currentLocation.position + 1, lap}
							nextAcceleration = desaccelerationRacer
							fmt.Println("deaccelerating3")
						}
					} else if currentLocation.rail == 7 {
						if track[currentLocation.rail-1][(currentLocation.position+1)%totalDistance] == " " {
							nextLocation = Location{id, currentLocation.rail - 1, currentLocation.position + 1, lap}
							nextAcceleration = acceleration
							fmt.Println(id, "i am at far right")
						} else {
							nextLocation = Location{id, currentLocation.rail, currentLocation.position + 1, lap}
							nextAcceleration = desaccelerationRacer
							fmt.Println("deaccelerating2")
						}
					} else {
						if track[currentLocation.rail+1][(currentLocation.position+1)%totalDistance] == " " {
							nextLocation = Location{id, currentLocation.rail + 1, currentLocation.position + 1, lap}
							nextAcceleration = acceleration
							fmt.Println(id, "i went to the right")
						} else if track[currentLocation.rail-1][(currentLocation.position+1)%totalDistance] == " " {
							nextLocation = Location{id, currentLocation.rail - 1, currentLocation.position + 1, lap}
							nextAcceleration = acceleration
							fmt.Println(id, "i went to the left")
						} else {
							nextLocation = Location{id, currentLocation.rail, currentLocation.position + 1, lap}
							nextAcceleration = desaccelerationRacer
							fmt.Println("deaccelerating")
						}
					}
				}
				if firstThreat || indexOutOfRange {
					break
				}
			}
			if firstThreat == false { //si no hay nada estorbando adelante, pone su siguiete ubicación recto.
				nextLocation = Location{id, currentLocation.rail, currentLocation.position + 1, lap}
				nextAcceleration = acceleration
			}
			//si el carro se encuentra en una zona de frenado (curvas)
			if (currentLocation.position >= 40 && currentLocation.position <= 50) || (currentLocation.position >= 80 && currentLocation.position <= 90) || (currentLocation.position >= 120 && currentLocation.position <= 130) || (currentLocation.position >= 160 && currentLocation.position <= 170) {
				nextAcceleration = desaccelerationCurve
			}
			if nextLocation.position >= totalDistance {
				nextLocation.position = 0
				fmt.Println(id, "my pos was total distance", nextLocation)
			}
			chanRequest <- nextLocation

			if <-response == true {
				destroy <- currentLocation
				break
			}
		}
		if nextLocation.position == 0 {
			fmt.Println(id, "a lap is done!", nextLocation)
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
				if currentVelocity-nextAcceleration < 0 {
					currentVelocity = 0
				} else {
					currentVelocity += nextAcceleration
				}

			}
		}
		updateChan <- Update{id, currentLocation.rail, currentLocation.position, lap, currentVelocity, elapsed.String(), time.Now().Sub(start).String()}
	}
}

/*type Update struct {
	id       int
	rail     int
	position int
	lap      int
	speed	 float64
	lapTime  string
	racingTime	string
}*/
func prints() {
	updateList := make([]Update, numOfRacers)
	funNames := [17]string{"?", "Mario", "Luigi", "Peach", "D.K", "Bowser", "Toad", "Rosalina", "Shyguy", "Boo", "Daisy",
		"Raio Mqueen", "Toreto", "El rey", "El oliver", "El tachas", "dicesiseis"}
	info := [7]string{"Player ", "Rail: ", "Position: ", "Lap: ", "Speed: ", "Lap Time: ", "GlobalTime: "}
	for {
		time.Sleep(500 * time.Millisecond)
		tmpUpdateList := make([]Update, 20)
		if len(winners) >= 1 {
			tmpUpdateList = make([]Update, 5)
		}

		for i := 0; i < len(tmpUpdateList); i++ {
			tmpUpdateList[i] = <-updateChan
		}

		for i := 0; i < len(tmpUpdateList); i++ {
			tmpid := tmpUpdateList[i].id
			updateList[tmpid-1] = tmpUpdateList[i]
		}

		tmpString := ""
		callClear()
		for j := 0; j < numOfRacers; j++ {
			tmptmpstring := info[0] + funNames[updateList[j].id]
			if len(tmptmpstring) < 15 {
				tmptmpstring += strings.Repeat(" ", (15 - len(tmptmpstring)))
			}
			tmpString += tmptmpstring
		}
		println(tmpString)
		tmpString = ""
		for j := 0; j < numOfRacers; j++ {
			tmptmpstring := info[1] + strconv.Itoa(updateList[j].rail)
			if len(tmptmpstring) < 15 {
				tmptmpstring += strings.Repeat(" ", (15 - len(tmptmpstring)))
			}
			tmpString += tmptmpstring
		}
		println(tmpString)
		tmpString = ""
		for j := 0; j < numOfRacers; j++ {
			tmptmpstring := info[2] + strconv.Itoa(updateList[j].position)
			if len(tmptmpstring) < 15 {
				tmptmpstring += strings.Repeat(" ", (15 - len(tmptmpstring)))
			}
			tmpString += tmptmpstring
		}
		println(tmpString)
		tmpString = ""
		for j := 0; j < numOfRacers; j++ {
			tmptmpstring := info[3] + strconv.Itoa(updateList[j].lap)
			if len(tmptmpstring) < 15 {
				tmptmpstring += strings.Repeat(" ", (15 - len(tmptmpstring)))
			}
			tmpString += tmptmpstring
		}
		println(tmpString)
		tmpString = ""
		for j := 0; j < numOfRacers; j++ {
			s := fmt.Sprintf("%f", updateList[j].speed)
			tmptmpstring := info[4] + s
			if len(tmptmpstring) < 15 {
				tmptmpstring += strings.Repeat(" ", (15 - len(tmptmpstring)))
			}
			tmpString += tmptmpstring
		}
		println(tmpString)
		tmpString = ""
		for j := 0; j < numOfRacers; j++ {
			tmptmpstring := info[5] + updateList[j].lapTime
			if len(tmptmpstring) < 15 {
				tmptmpstring += strings.Repeat(" ", (15 - len(tmptmpstring)))
			}
			tmpString += tmptmpstring
		}
		println(tmpString)
		tmpString = ""
		for j := 0; j < numOfRacers; j++ {
			tmptmpstring := info[6] + updateList[j].racingTime
			if len(tmptmpstring) < 15 {
				tmptmpstring += strings.Repeat(" ", (15 - len(tmptmpstring)))
			}
			tmpString += tmptmpstring
		}
		println(tmpString)
	}

}
