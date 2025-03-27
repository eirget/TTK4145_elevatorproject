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

	var port string
	flag.StringVar(&port, "port", "", "port of this peer")
	flag.Parse()

	//automatically assign ID
	portNum, err := strconv.Atoi(port)
	if err != nil {
		fmt.Printf("Invalid port number %v \n", err)
	}
	id := (portNum - 20000)
	idStr := strconv.Itoa(id)

	var latestLost []string
	var latestLostMutex sync.Mutex

	var disconnectedFromNetwork = true
	var disconnectedMutex sync.RWMutex

	//map to store elevator states for all elevators on system, backup for orders
	elevatorMap := make(map[string]*elevator.Elevator)

	addr := "localhost:" + port
	elevio.Init(addr, config.NumFloors)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	elevio.ElevioInit(drv_buttons, drv_floors, drv_obstr, drv_stop)

	//channels for receiving and transmitting updates on the id's of the peers that are alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	//channels for sending and receiving Elevator types
	elevStateTx := make(chan elevator.Elevator)
	elevStateRx := make(chan elevator.Elevator)

	//channels for sending and receiving runHra signal
	runHraCh := make(chan struct{}, 1)
	receiveRunHraCh := make(chan struct{}, 1)

	newOrderCh := make(chan struct{}, 1)

	network.NetworkInit(idStr, peerUpdateCh, peerTxEnable, elevStateTx, elevStateRx, runHraCh, receiveRunHraCh)

	eAtFloor := elevator.WaitForValidFloor(elevio.MD_Up, drv_floors)
	fmt.Println("Elevator initalized at floor: ", eAtFloor)

	localElevator := elevator.ElevatorInit(eAtFloor, id)

	elevStateTx <- *localElevator

	elevio.SetFloorIndicator(localElevator.FloorNr)

	fmt.Printf("Started elevator system \n")

	go fsm(localElevator, elevStateTx, drv_buttons, drv_floors, drv_obstr, drv_stop, config.NumFloors, newOrderCh)

	go elevator.MonitorActivity(localElevator, runHraCh) //SHOULD WE TRY TO MAKE THIS A METHOD?

	go localElevator.LightUpdater()

	go handleElevatorUpdates(localElevator, elevStateRx, elevatorMap, runHraCh, newOrderCh)

	go handlePeerUpdates(peerUpdateCh, &latestLost, &latestLostMutex, &disconnectedFromNetwork, &disconnectedMutex, elevStateTx, localElevator, &elevatorMapLock, runHraCh, elevatorMap)

	go handleRunHraRequest(receiveRunHraCh, localElevator, elevatorMap, &elevatorMapLock, elevStateTx, elevStateRx, &latestLost, &latestLostMutex, &disconnectedFromNetwork, &disconnectedMutex, hraExecutable)

	select {}
}
