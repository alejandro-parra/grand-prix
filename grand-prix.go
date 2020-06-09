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

//type racer chan<- Location

var (
	track    [][]string
	funNames = [17]string{"?", "Mario", "Luigi", "Peach", "D.K", "Bowser", "Toad", "Rosalina", "Guido", "King Boo", "Daisy",
		"Raio Mqueen", "Toreto", "El rey", "El oliver", "El tachas", "el sonic"}
	funChars      = [17]string{"?", "M", "L", "P", "DK", "B", "T", "R", "G", "KB", "D", "~", "%", "$", "@", ">", "*"}
	competitors   = make(map[int]chan bool) //the reference to each communication with the racers
	requests      = make(chan Location)     //a channel that all racers use to ask main to move
	destroy       = make(chan Location, 60)
	updateChan    = make(chan Update, 60) //channel to provide the printing system each racer's stats
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
	competitors := make(map[int]chan bool)
	//array used to specify starting positions of racers at start of race
	initialPositions := [16]int{0, 0, 0, 0, 0, 0, 0, 0, 3, 3, 3, 3, 3, 3, 3, 3}

	nr := flag.Int("racers", 8, "number of racers!")
	nl := flag.Int("laps", 3, "number of laps!")
	flag.Parse()
	numOfRacers = *nr
	totalLaps = *nl
	if numOfRacers > 16 || numOfRacers < 1 || totalLaps < 1 {
		fmt.Println("wrong number of racers or laps, must be between 1 and 16")
		return
	}
	//initialize empty track array
	for i := range track {
		track[i] = make([]string, totalDistance)
	}
	for i := range track {
		for j := range track[i] {
			track[i][j] = " "
		}
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
				track[recievedRequest.rail][recievedRequest.position] = funChars[recievedRequest.id]
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
								fmt.Println(i+1, ") ", funNames[winners[i]])
							}
							return
						}
					} else {
						if len(winners) == 3 {
							killPrint <- struct{}{}
							fmt.Println("Race is over!")
							fmt.Println("THE WINNERS ARE:")
							fmt.Println("1ST: ", funNames[winners[0]])
							fmt.Println("2ND: ", funNames[winners[1]])
							fmt.Println("3RD: ", funNames[winners[2]])
							fmt.Println("CONGRATULATIONS !!!!!!!!!!!")
							return
						}
					}
				}
			} else {
				competitors[recievedRequest.id] <- false
			}
		//a call to destroy an object from the track
		case recievedRequest := <-destroy:
			if track[recievedRequest.rail][recievedRequest.position] == funChars[recievedRequest.id] {
				track[recievedRequest.rail][recievedRequest.position] = " "
			} else {
				fmt.Println(funNames[recievedRequest.id], ": LET'S GO!")
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

	startLap := time.Now()
	elapsed := time.Now().Sub(startLap)
	id := initLocation.id
	currentLocation := initLocation
	currentVelocity := acceleration
	desaccelerationRacer := -400.0
	r2 := rand.New(rand.NewSource(time.Now().UnixNano()))
	desaccelerationCurve := float64(-(r2.Intn(50-10) + 10))
	sleep := 800.0

	nextLocation := Location{0, 0, 0, 0}
	nextAcceleration := 0.0
	lastUpdateCar := ""
	lap := initLocation.currentLap
	for lap < totalLaps+1 { //mientras el coche no haya terminado la carrera
		//if currentVelocity < maxSpeed { //si el carro no ha llegado a su limite de velocidad
		time.Sleep(time.Duration(sleep-currentVelocity) * time.Millisecond)
		//} else { //si el carro ya llego a su limite de velocidad
		//	time.Sleep(time.Duration(sleep-maxSpeed) * time.Millisecond) //minimo 100
		//}
		for {
			firstThreat := false
			//se checan las siguientes 5 posiciones en busqueda de carros que estorben
			for i := (currentLocation.position + 1) % totalDistance; i != (currentLocation.position+5)%totalDistance; i = (i + 1) % totalDistance {
				if track[currentLocation.rail][i] != " " { //si en esta posición hay un carro
					firstThreat = true
					//ver si se puede mover a los lados y rebasar el otro carro
					if currentLocation.rail == 0 {
						if track[currentLocation.rail+1][(currentLocation.position+1)%totalDistance] == " " {
							nextLocation = Location{id, currentLocation.rail + 1, currentLocation.position + 1, lap}
							nextAcceleration = acceleration
							lastUpdateCar = fmt.Sprintf("i went right")
						} else {
							nextLocation = Location{id, currentLocation.rail, currentLocation.position + 1, lap}
							nextAcceleration = desaccelerationRacer
							lastUpdateCar = fmt.Sprintf("deaccel")
						}
					} else if currentLocation.rail == 7 {
						if track[currentLocation.rail-1][(currentLocation.position+1)%totalDistance] == " " {
							nextLocation = Location{id, currentLocation.rail - 1, currentLocation.position + 1, lap}
							nextAcceleration = acceleration
							lastUpdateCar = fmt.Sprintf("i went left")
						} else {
							nextLocation = Location{id, currentLocation.rail, currentLocation.position + 1, lap}
							nextAcceleration = desaccelerationRacer
							lastUpdateCar = fmt.Sprintf("deaccel")
						}
					} else {
						if track[currentLocation.rail+1][(currentLocation.position+1)%totalDistance] == " " {
							nextLocation = Location{id, currentLocation.rail + 1, currentLocation.position + 1, lap}
							nextAcceleration = acceleration
							lastUpdateCar = fmt.Sprintf("i went right")
						} else if track[currentLocation.rail-1][(currentLocation.position+1)%totalDistance] == " " {
							nextLocation = Location{id, currentLocation.rail - 1, currentLocation.position + 1, lap}
							nextAcceleration = acceleration
							lastUpdateCar = fmt.Sprintf("i went left")
						} else {
							nextLocation = Location{id, currentLocation.rail, currentLocation.position + 1, lap}
							nextAcceleration = desaccelerationRacer
							lastUpdateCar = fmt.Sprintf("deaccel")
						}
					}
				}
				//si ya encontro algo adelante, no sigas buscando
				if firstThreat {
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
			}
			chanRequest <- nextLocation

			if <-response == true {
				destroy <- currentLocation
				break
			}
		}
		if nextLocation.position == 0 {
			lastUpdateCar = "passed lap!"
			elapsed = time.Now().Sub(startLap)
			startLap = time.Now()
			lap++
		}
		currentLocation = nextLocation
		if nextAcceleration > 0 {
			if currentVelocity < maxSpeed {
				currentVelocity += nextAcceleration
			} else {
				currentVelocity = maxSpeed
			}
		} else {
			if currentVelocity+nextAcceleration < 0 {
				currentVelocity = 0
			} else {
				currentVelocity += nextAcceleration
			}
		}
		updateChan <- Update{id, currentLocation.rail, currentLocation.position, lap, currentVelocity, elapsed.String(), lastUpdateCar}
	}
}

// formats the print to be the same size as the others
func formatPrint(thePrint string, numSpaces int) string {
	if len(thePrint) < numSpaces {
		thePrint += strings.Repeat(" ", (numSpaces - len(thePrint)))
	}
	return thePrint
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

		for i := 0; i < space; i++ {
			select {
			case x := <-updateChan:
				updateList[x.id-1] = x
			case <-killT:
				return
			}
		}
		callClear()

		params := make([]string, 7)
		for j := 0; j < numOfRacers; j++ {
			if j == 8 {
				for i, v := range params {
					fmt.Println(v)
					params[i] = ""
				}
				fmt.Println("")
			}
			tmpString := info[0] + funNames[updateList[j].id] + " - " + funChars[updateList[j].id]
			params[0] += formatPrint(tmpString, numSpaces)                                    //id
			params[1] += formatPrint(info[1]+strconv.Itoa(updateList[j].rail), numSpaces)     //rail
			params[2] += formatPrint(info[2]+strconv.Itoa(updateList[j].position), numSpaces) //position
			if updateList[j].lap == totalLaps+1 {
				tmpString = info[3] + "Finished!"
			} else {
				tmpString = info[3] + strconv.Itoa(updateList[j].lap)
			}
			params[3] += formatPrint(tmpString, numSpaces) //location
			s := fmt.Sprintf("%f", updateList[j].speed)
			params[4] += formatPrint(info[4]+s, numSpaces)                        //speed
			params[5] += formatPrint(info[5]+updateList[j].lapTime, numSpaces)    //lap time
			params[6] += formatPrint(info[7]+updateList[j].lastUpdate, numSpaces) //last update
		}
		for _, v := range params {
			fmt.Println(v)
		}
		fmt.Println("")
		fmt.Println("Total Time: ", time.Now().Sub(start).String())
		printTrack()
	}

}
