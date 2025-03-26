package main

import (
	"Driver_go/config"
	"Driver_go/elevator"
	"Driver_go/elevio"
	"Driver_go/network"
	"Driver_go/network/peers"
	"flag"
	"fmt"
	"strconv"
	"sync"
)

var elevatorMapLock sync.Mutex

func main() {

	hraExecutable := "hall_request_assigner"

	//make this automatic later
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	var port string
	flag.StringVar(&port, "port", "", "port of this peer")
	flag.Parse()

	var latestLost []string
	var latestLostMutex sync.Mutex

	//NEW
	var disconnectedFromNetwork = true //assume until proven otherwise
	var disconnectedMutex sync.RWMutex

	//create map to store elevator states for all elevators on system, to backup orders
	//why string? maybe just decide that all cases of ID should just be string
	elevatorMap := make(map[string]*elevator.Elevator)

	addr := "localhost:" + port
	elevio.Init(addr, config.NumFloors)

	//fmt.Printf("Before go routines \n")

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	elevio.ElevioInit(drv_buttons, drv_floors, drv_obstr, drv_stop)

	//make a channel for receiving updates on the id's of the peers that are alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	//we can enable/disable the transmitter after it has been started, coulb be used to signal that we are unavailable
	peerTxEnable := make(chan bool)
	//we make channels for sending and receiving our custom data types
	elevStateTx := make(chan elevator.Elevator)
	elevStateRx := make(chan elevator.Elevator)

	runHraCh := make(chan bool, 10)
	receiveRunHraCh := make(chan bool, 10)

	network.NetworkInit(id, peerUpdateCh, peerTxEnable, elevStateTx, elevStateRx, runHraCh, receiveRunHraCh)

	eAtFloor := elevio.WaitForValidFloor(elevio.MD_Up, drv_floors)
	fmt.Println("Elevator initalized at floor: ", eAtFloor)

	id_num, _ := strconv.Atoi(id)

	localElevator := elevator.ElevatorInit(eAtFloor, id_num) //figure out if the impostet file should change name (like this) or of the elevator.go file should change name, or maybe nothing should just be named elevator!

	elevStateTx <- *localElevator

	elevio.SetFloorIndicator(localElevator.Floor_nr)

	go fsm(localElevator, elevStateTx, drv_buttons, drv_floors, drv_obstr, drv_stop, config.NumFloors)

	go elevator.MonitorActivity(localElevator, runHraCh)

	go localElevator.RunLightUpdater()

	fmt.Printf("Started elevator system \n")

	go handleElevatorUpdates(localElevator, elevStateRx, elevatorMap, runHraCh)

	go handlePeerUpdates(peerUpdateCh, &latestLost, &latestLostMutex, &disconnectedFromNetwork, &disconnectedMutex, elevStateTx, localElevator, &elevatorMapLock, runHraCh, elevatorMap)

	go handleRunHraRequest(receiveRunHraCh, localElevator, elevatorMap, &elevatorMapLock, elevStateTx, elevStateRx, &latestLost, &latestLostMutex, &disconnectedFromNetwork, &disconnectedMutex, hraExecutable)

}
