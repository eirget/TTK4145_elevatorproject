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
	"sync"
	"time"
)

// where should this go???
func monitorElevatorActivity(e *elev_import.Elevator, runHra chan bool) {
	ticker := time.NewTicker(1 * time.Second) // Check every second
	defer ticker.Stop()
	fmt.Printf("Last active: %v \n: ", e.LastActive)
	// need to double check with some sort of "heartbeat" if it actually doesnt work, update lastActive if nothing is wrong
	for range ticker.C {
		if time.Since(e.LastActive) > 5*time.Second { // Elevator inactive for 5+ seconds
			fmt.Println("I have been inactive")
			if e.HasPendingOrders() {
				fmt.Println("And I have pending orders, calling hall request assigner")
				runHra <- true // Trigger hall request reassignment
				//return
			}
		}
	}
}

func periodicallySetLights(e *elev_import.Elevator) {
	ticker := time.NewTicker(100 * time.Millisecond) // Check every second
	defer ticker.Stop()
	for range ticker.C {
		e.SetLights()
	}
}

var elevatorMapLock sync.Mutex //NEW

func main() {

	hraExecutable := "hall_request_assigner"

	//make this automatic later
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	var port string
	flag.StringVar(&port, "port", "", "port of this peer")
	flag.Parse()

	//cant these just be made inside hra instead?

	//create map to store elevator states for all elevators on system, to backup orders
	//why string? maybe just decide that all cases of ID should just be string
	elevatorMap := make(map[string]*elev_import.Elevator)

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
	elevStateTx := make(chan elev_import.Elevator)
	elevStateRx := make(chan elev_import.Elevator)

	runHra := make(chan bool, 10)
	receiveRunHra := make(chan bool, 10)

	network.NetworkInit(id, peerUpdateCh, peerTxEnable, elevStateTx, elevStateRx, runHra, receiveRunHra)

	eAtFloor := elevio.WaitForValidFloor(elevio.MD_Up, drv_floors)
	fmt.Println("Elevator initalized at floor: ", eAtFloor)

	id_num, _ := strconv.Atoi(id)
	elevator := elev_import.ElevatorInit(eAtFloor, id_num) //figure out if the impostet file should change name (like this) or of the elevator.go file should change name, or maybe nothing should just be named elevator!

	elevStateTx <- *elevator

	elevio.SetFloorIndicator(elevator.Floor_nr)

	go fsm(elevator, elevStateTx, drv_buttons, drv_floors, drv_obstr, drv_stop, config.NumFloors)

	go monitorElevatorActivity(elevator, runHra)

	go periodicallySetLights(elevator)

	fmt.Printf("Started elevator system \n")

	go func() {

		for elevRx := range elevStateRx {
			//update elevator to have newest state of other elevators
			idStr := strconv.Itoa(elevRx.ID)

			elevatorMapLock.Lock()

			//elevatorMap[idStr] = &elevRx //may have to directly allocate new Elevator pointer
			//alternative to the above line:
			copy := elevRx
			elevatorMap[idStr] = &copy

			if idStr == id {
				elevator.Orders = elevRx.Orders
				elevator.LastActive = time.Now()
				//new_order_flag = true
			}

			//fmt.Printf("Elevators: %v\n", elevatorMap)
			//elevatorVar := elevatorMap[idStr]
			//fmt.Printf("Elevator behavior: %v id: %d", elevatorVar.Behavior, elevatorVar.ID)
			elevatorMapLock.Unlock()

			//fmt.Printf("Floor_nr: %v\n", a.Floor_nr)
			//fmt.Printf("Direction %v\n", a.Direction)
			//fmt.Println("timestamp(hall up): \n", a.Orders[a.Floor_nr][BT_HallUp].Timestamp)

			//NEW, idea for fixing when hall requests should actually be updated
			//func updateHallRequests(myElevator *Elevator, receivedElev Elevator) {
			//if idStr != id {

			updated := false
			for floor := 0; floor < config.NumFloors; floor++ {
				for btn := 0; btn < config.NumButtons-1; btn++ { // Only HallUp and HallDown
					// Compare timestamps to ensure only newer updates are accepted
					if elevRx.Orders[floor][btn].Timestamp.After(elevator.Orders[floor][btn].Timestamp) {
						elevator.Orders[floor][btn] = elevRx.Orders[floor][btn]
						updated = true
					}
				}
			}
			//}
			if updated {
				updated = false
				fmt.Printf("Recieved: \n")
				fmt.Printf("Message from ID: %v\n", elevRx.Orders[1][2].ElevatorID)
				for floor := 0; floor < config.NumFloors; floor++ {
					fmt.Printf("\n Floornr: %+v ", floor)
					for btn := elevio.ButtonType(0); btn < config.NumButtons; btn++ {
						fmt.Printf("%+v ", elevRx.Orders[floor][btn].State)
						fmt.Printf("%+v, ", elevRx.Orders[floor][btn].ElevatorID)
					}
				}
				fmt.Printf("\n")
				fmt.Printf("Timestamp: %v \n", time.Now())

			}

			elevator.SetLights()

			//if updated {
			//	elevStateTx <- *elevator
			//}

			//maybe just usa an UPDATED flag in here. and rebroadcast aand run hra based on this instead
			if new_order_flag {
				runHra <- true
				new_order_flag = false
			}
			//fmt.Printf("%+v\n", elevatorMap)
		}
	}()

	var latestLost []string
	var latesetLostMutex sync.Mutex

	for {
		select {
		case peerUpdate := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peerUpdate.Peers)
			fmt.Printf("  New:      %q\n", peerUpdate.New)
			fmt.Printf("  Lost:     %q\n", peerUpdate.Lost)

			latesetLostMutex.Lock()
			latestLost = peerUpdate.Lost
			latesetLostMutex.Unlock()

			//fixed elevators knowing of eachother even without an event happening
			if len(peerUpdate.New) != 0 {
				elevStateTx <- *elevator
				//when an elevator loses power, but then gets power again AND connects to the internet again it gets its old cab orders back through its whole Orders
				//the hall call part of Orders in the elevators map should have been updated while it was disconnected also
				//QUESTION: do we need to have functionality for the case where an elevator loses power, gets power again but does NOT connect back on the Internet? should it still get its cab calls somehow
				//hra should maybe be run when new and lost change
				if lastState, exists := elevatorMap[peerUpdate.New]; exists {
					fmt.Printf("Elevator %s reconnected! Restoring previous orders.\n", peerUpdate.New)
					elevatorMap[peerUpdate.New].Orders = lastState.Orders // Restore orders
					fmt.Printf("Orders at ID %s: %v", peerUpdate.New, elevatorMap[peerUpdate.New].Orders)
					elevStateTx <- *elevatorMap[peerUpdate.New] // Broadcast restored state
				} else {
					fmt.Printf("Elevator %s is new, initializing.\n", peerUpdate.New)
				}
			}

			if len(peerUpdate.Lost) != 0 {
				runHra <- true
			}

			//the new part of peers will only include one id, it is not a list, is that ok?

			/*
				for newID := range peerUpdate.New {
					if lastState, exists := elevatorMap[newID]; exists {
						fmt.Printf("Elevator %s reconnected! Restoring previous orders.\n", newID)
						elevatorMap[newID].Orders = lastState.Orders // Restore orders
						elevStateTx <- *elevatorMap[newID]           // Broadcast restored state
					} else {
						fmt.Printf("Elevator %s is new, initializing.\n", newID)
					}
				}
			*/

		//is current "heartbeat" functionality enough?
		//maybe both variable and channel name should include that these are states, maybe change names
		//case elevRx := <-elevStateRx: //can the buffer cause packet loss?

		case <-receiveRunHra:

			//sleep at the beginning of hra? or timer? just something to delay a little bit so that all elevators have latest orders

			//actually create logic that will be correct for all cases
			fmt.Println("Received runHra signal")

			activeElevators := make(map[string]*elev_import.Elevator) //have just in pure main, does this need to be pointers
			//just empty active elevators here

			fmt.Printf("Elevator map: %v\n", elevatorMap)

			elevatorMapLock.Lock()
			for id, elev := range elevatorMap {
				fmt.Printf("Elevator behaviour of id %s before hra: %v\n", id, elev.Behavior)
				fmt.Printf("Last active of id %v: ", id)
				fmt.Println(elev.LastActive)
				fmt.Printf("Time now: %v", time.Now())
				if contains(latestLost, id) {
					continue
				}

				if time.Since(elev.LastActive) < 5*time.Second { //|| !elev.HasPendingOrders() {
					copy := *elev
					activeElevators[id] = &copy
				}

			}

			fmt.Printf("Active elevators: %v\n", activeElevators)
			elevatorMapLock.Unlock()

			go hallRequestAssigner(elevator, elevatorMap, activeElevators, id, hraExecutable, elevStateRx)

			//might not be neccessary at all
			//case <-time.After(500 * time.Millisecond):
			//	elevStateTx <- *elevator
		}
	}
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
