package main

import (
	"Driver_go/elevio"
	"Driver_go/network/bcast"
	"Driver_go/network/peers"
	"flag"
	"fmt"
	"strconv"
	"time"
)

type HelloMsg struct {
	Message string
	Iter    int
}

type HRAElevState struct {
	Behavior    string `json:"behavior"` //behaviour?
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

const (
	NumFloors  = 4
	NumButtons = 3
)

func main() {

	//NETWORK

	hraExecutable := "hall_request_assigner"

	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	id = strconv.Itoa(55)

	hallRequests := make([][2]bool, NumFloors)
	cabRequests := make([]bool, NumFloors)

	//create map to store elevator states for all elevators on system
	elevators := make(map[string]Elevator)

	fmt.Printf("%+v\n", elevators)

	//make a channel for receiving updates on the id's of the peers that are alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	//we can enable/disable the transmitter after it has been started, coulb be used to signal that we are unavailable
	peerTxEnable := make(chan bool)
	//we make channels for sending and receiving our custom data types
	elevStateTx := make(chan Elevator)
	elevStateRx := make(chan Elevator)

	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	go bcast.Transmitter(20022, elevStateTx)
	go bcast.Receiver(20022, elevStateRx)

	elevio.Init("localhost:20244", NumFloors) //gjør til et flag

	var d elevio.MotorDirection = elevio.MD_Up
	elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	for f := 0; f < NumFloors; f++ {
		for b := elevio.ButtonType(0); b < 3; b++ {
			elevio.SetButtonLamp(b, f, false)
		}
	}
	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)

	a := <-drv_floors
	for a == -1 {
		elevio.SetMotorDirection(d)
	}
	elevio.SetMotorDirection(elevio.MD_Stop)

	id_num, _ := strconv.Atoi(id)
	elevator := ElevatorInit(a, id_num) //kanskje det blir penere å bare bruke string

	elevStateTx <- *elevator

	elevio.SetFloorIndicator(elevator.Floor_nr)

	go fsm(elevator, elevators, id, hallRequests, cabRequests, hraExecutable, drv_buttons, drv_floors, drv_obstr, drv_stop, NumFloors)

	//the example message, we just send one of these every second
	/*
		go func() {
			for {
				ordersTx <- elevator.Orders
				time.Sleep(1 * time.Second)
			}
		}()
	*/

	//HALL REQUESTS

	//non-blocking timer
	hraTimer := time.NewTimer(0)
	<-hraTimer.C
	hraTimer.Reset(1 * time.Second)

	fmt.Printf("Started elevator system \n")

	//TODO: NEED TO BROADCAST ELEVATOR STATES OFTEN

	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			//add to "elevators" if new

		//heartbeat check functionality in below case as well, time each ID, maybe peers is good enough already
		case a := <-elevStateRx: //kan hende denne vil miste orders om det blir fullt i buffer
			//update elevator to have newest state of other elevators
			idStr := strconv.Itoa(a.Orders[0][2].ElevatorID)
			elevators[idStr] = a

			fmt.Printf("Recieved: \n")
			fmt.Printf("Message from ID: %v\n", a.Orders[1][2].ElevatorID)
			fmt.Printf("Floor_nr: %v\n", a.Floor_nr)
			fmt.Printf("Direction %v\n", a.Direction)

			//fmt.Printf("%+v\n", elevators)

		}
	}
}
