# Build/Run automation

Grand Prix Running instructions

## Requirements

* Have Go installed in the computer.
* Computer OS must be Linux or Windows.
* The program will run in console, you may need to adjust the font size and window size for the outputs to look good.

## Run instructions

### Build grand-prix.go
In order to build the program you can either

* use "go build grand-prix.go" directly
or
* use make

### Run grand-prix.go
In order to run the program you can either

* use "go run grand-prix.go <optional parameter 1> <optional parameter 2>" directly (does not need build)
or
* use make run <optional parameter 1> <optional parameter 2>
or
* use ./grand-prix.go <optional parameter 1> <optional parameter 2>

### Parameters
All parameters are optional and have a default value.

* **-racers** Number of racing cars to put in the simulator. default value = 8
* **-laps** Number of laps needed to win. default value = 3
