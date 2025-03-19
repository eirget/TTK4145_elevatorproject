package main

import (
	"Driver_go/config"
	elev_import "Driver_go/elevator"
	"Driver_go/elevio"
	"Driver_go/network"
	"Driver_go/network/peers"
	"flag"
	"fmt"
	"strconv"
	"time"
)

//tested to obstruct the closest elevator to a call, the available elevator did not take over the order

func monitorElevatorActivity(e *elev_import.Elevator, runHra chan bool) {
	ticker := time.NewTicker(1 * time.Second) // Check every second
	defer ticker.Stop()
	// need to double check with some sort of "heartbeat" if it actually doesnt work, update lastActive if nothing is wrong
	for range ticker.C {
		if time.Since(e.LastActive) > 5*time.Second { // Elevator inactive for 5+ seconds
			if e.HasPendingOrders() {
				runHra <- true // Trigger hall request reassignment
				return
			}
		}
	}
}

func main() {

	hraExecutable := "hall_request_assigner"

	//make this automatic later
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	var port string
	flag.StringVar(&port, "port", "", "port of this peer")
	flag.Parse()

	hallRequests := make([][2]bool, config.NumFloors)
	cabRequests := make([]bool, config.NumFloors)

	//create map to store elevator states for all elevators on system, !!! point to discuss: *Elevator or not?
	elevatorMap := make(map[string]*elev_import.Elevator)

	addr := "localhost:" + port
	elevio.Init(addr, config.NumFloors) //gjør til et flag

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
	elevStateTx := make(chan elev_import.Elevator)
	elevStateRx := make(chan elev_import.Elevator)

	runHra := make(chan bool, 10)
	receiveRunHra := make(chan bool, 10)

	network.NetworkInit(id, peerUpdateCh, peerTxEnable, elevStateTx, elevStateRx, runHra, receiveRunHra)

	eAtFloor := elevio.WaitForValidFloor(elevio.MD_Up, drv_floors)
	fmt.Println("Elevator initialized at floor:", eAtFloor)
	fmt.Println("Elevator initalized at floor: ", eAtFloor)

	id_num, _ := strconv.Atoi(id)
	elevator := elev_import.ElevatorInit(eAtFloor, id_num) //kanskje det blir penere å bare bruke string

	elevStateTx <- *elevator

	elevio.SetFloorIndicator(elevator.Floor_nr)

	go fsm(elevator, elevStateTx, drv_buttons, drv_floors, drv_obstr, drv_stop, config.NumFloors, runHra)
	//go hraSignalListener(elevator, elevators, id, hallRequests, cabRequests, hraExecutable, elevStateTx, run_hra)

	go monitorElevatorActivity(elevator, runHra)

	fmt.Printf("Started elevator system \n")

	for {
		select {
		//ikke fornøyd med navnet peers, fordi det er new og lost også
		case peers := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peers.Peers)
			fmt.Printf("  New:      %q\n", peers.New)
			fmt.Printf("  Lost:     %q\n", peers.Lost)

			//fixed elevators knowing of eachother even without an event happening
			if len(peers.New) != 0 {
				elevStateTx <- *elevator
			}

		//heartbeat check functionality in below case as well, time each ID, maybe peers is good enough already
		case elevRx := <-elevStateRx: //kan hende denne vil miste orders om det blir fullt i buffer
			//update elevator to have newest state of other elevators
			idStr := strconv.Itoa(elevRx.ID)

			elevatorMap[idStr] = &elevRx //may have to directly allocate new Elevator pointer
			fmt.Printf("Elevators: %v", elevatorMap)

			fmt.Printf("Recieved: \n")
			fmt.Printf("Message from ID: %v\n", elevRx.Orders[1][2].ElevatorID)
			//fmt.Printf("Floor_nr: %v\n", a.Floor_nr)
			//fmt.Printf("Direction %v\n", a.Direction)
			//fmt.Println("timestamp(hall up): \n", a.Orders[a.Floor_nr][BT_HallUp].Timestamp)

			//NEW, idea for fixing when hall requests should actually be updated
			//func updateHallRequests(myElevator *Elevator, receivedElev Elevator) {
			//if idStr != id {
			for floor := 0; floor < config.NumFloors; floor++ {
				for btn := 0; btn < config.NumButtons-1; btn++ { // Only HallUp and HallDown
					// Compare timestamps to ensure only newer updates are accepted
					if elevRx.Orders[floor][btn].Timestamp.After(elevator.Orders[floor][btn].Timestamp) {
						elevator.Orders[floor][btn] = elevRx.Orders[floor][btn]
					}
				}
			}
			//}

			for floor := 0; floor < config.NumFloors; floor++ {
				fmt.Printf("\n Floornr: %+v ", floor)
				for btn := elevio.ButtonType(0); btn < config.NumButtons; btn++ {
					fmt.Printf("%+v ", elevRx.Orders[floor][btn].State)
					fmt.Printf("%+v, ", elevRx.Orders[floor][btn].ElevatorID)
				}

			}
			fmt.Printf("\n")
			fmt.Printf("Timestamp: %v \n", time.Now())

			elevator.SetLights()

			if new_order_flag {
				runHra <- true
				new_order_flag = false
			}

			//}

			//fmt.Printf("%+v\n", elevatorMap)

		case <-receiveRunHra:

			//actually create logic that will be correct for all cases
			fmt.Println("Received runHra signal")

			activeElevators := make(map[string]*elev_import.Elevator)

			for id, elev := range elevatorMap {
				if time.Since(elev.LastActive) < 5*time.Second || !elev.HasPendingOrders() {
					activeElevators[id] = elev
				}
			}
			go hallRequestAssigner(elevator, activeElevators, id, hallRequests, cabRequests, hraExecutable, elevStateRx)

			//might not be neccessary at all
			//case <-time.After(500 * time.Millisecond):
			//	elevStateTx <- *elevator
		}
	}
}
