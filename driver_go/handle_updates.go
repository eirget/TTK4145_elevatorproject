package main

import (
	"Driver_go/config"
	"Driver_go/elevator"
	"Driver_go/elevio"
	"Driver_go/network/peers"
	"Driver_go/utils"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// update elevator to have newest state of other elevators
func handleElevatorUpdates(localElevator *elevator.Elevator, elevStateRx <-chan elevator.Elevator, elevatorMap map[string]*elevator.Elevator, runHraCh chan struct{}, newOrderCh chan struct{}) {
	for elevRx := range elevStateRx {

		elevatorMapLock.Lock()
		copy := elevRx
		idStr := strconv.Itoa(elevRx.ID)
		elevatorMap[idStr] = &copy

		if elevRx.ID == localElevator.ID {
			localElevator.Orders = elevRx.Orders
			//localElevator.LastActive = time.Now()  //WAS THERE AN IMPORTANT REASON FOR THIS!!! DONT ADD BACK!
			//new_order_flag = true
		}

		elevatorMapLock.Unlock()

		updated := false
		for floor := 0; floor < config.NumFloors; floor++ {
			for btn := 0; btn < config.NumButtons-1; btn++ { // Only HallUp and HallDown
				// Compare timestamps to ensure only newer updates are accepted
				if elevRx.Orders[floor][btn].Timestamp.After(localElevator.Orders[floor][btn].Timestamp) {
					localElevator.Orders[floor][btn] = elevRx.Orders[floor][btn]
					updated = true
				}
			}
		}

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

		localElevator.SetLights()

		//if newOrderFlag {
		//	runHraCh <- true
		//newOrderFlag = false
		//}
		for range newOrderCh {
			select {
			case runHraCh <- struct{}{}: // Trigger hall request reassignment
			default:

			}
		}
	}
}

func handlePeerUpdates(
	peerUpdateCh <-chan peers.PeerUpdate,
	latestLost *[]string,
	latestLostMutex *sync.Mutex,
	disconnectedFromNetwork *bool,
	disconnectedMutex *sync.RWMutex,
	elevStateTx chan<- elevator.Elevator,
	localElevator *elevator.Elevator,
	elevatorMapLock *sync.Mutex,
	runHraCh chan struct{},
	elevatorMap map[string]*elevator.Elevator) {
	for peerUpdate := range peerUpdateCh {
		fmt.Printf("Peer update:\n")
		fmt.Printf("  Peers:    %q\n", peerUpdate.Peers)
		fmt.Printf("  New:      %q\n", peerUpdate.New)
		fmt.Printf("  Lost:     %q\n", peerUpdate.Lost)

		latestLostMutex.Lock()
		*latestLost = peerUpdate.Lost
		latestLostMutex.Unlock()

		disconnectedMutex.Lock()
		*disconnectedFromNetwork = len(peerUpdate.Peers) == 0
		disconnectedMutex.Unlock()

		elevatorMapLock.Lock()
		if len(peerUpdate.New) != 0 {
			elevStateTx <- *localElevator
			if lastState, exists := elevatorMap[peerUpdate.New]; exists {
				fmt.Printf("Elevator %s reconnected! Restoring previous orders.\n", peerUpdate.New)
				elevatorMap[peerUpdate.New].Orders = lastState.Orders // Restore orders
				fmt.Printf("Orders at ID %s: %v", peerUpdate.New, elevatorMap[peerUpdate.New].Orders)
				elevStateTx <- *elevatorMap[peerUpdate.New] // Broadcast restored state
			} else {
				fmt.Printf("Elevator %s is new, initializing.\n", peerUpdate.New)
			}
		}
		elevatorMapLock.Unlock()

		if len(peerUpdate.Lost) != 0 {
			select {
			case runHraCh <- struct{}{}:
			default:

			}
		}
	}
}

func handleRunHraRequest(
	receiveRunHraCh <-chan struct{},
	localElevator *elevator.Elevator,
	elevatorMap map[string]*elevator.Elevator,
	elevatorMapLock *sync.Mutex,
	elevStateTx chan<- elevator.Elevator,
	elevStateRx chan elevator.Elevator,
	latestLost *[]string,
	latestLostMutex *sync.Mutex,
	disconnectedFromNetwork *bool,
	disconnectedMutex *sync.RWMutex,
	hraExecutable string) {
	for range receiveRunHraCh {
		//sleep at the beginning of hra? or timer? just something to delay a little bit so that all elevators have latest orders

		//actually create logic that will be correct for all cases
		fmt.Println("Received runHra signal")

		//NEW
		disconnectedMutex.RLock()
		isDisconnected := *disconnectedFromNetwork
		disconnectedMutex.RUnlock()

		if isDisconnected {
			fmt.Println("Alone mode: assigning all hall requests to self")
			elevator.AssignAllHallCallsToSelf(localElevator)
			elevStateTx <- *localElevator //why?
			continue
		}

		latestLostMutex.Lock()
		lostCopy := append([]string(nil), (*latestLost)...)
		latestLostMutex.Unlock()

		activeElevators := make(map[string]*elevator.Elevator) //have just in pure main, does this need to be pointers
		//just empty active elevators here

		fmt.Printf("Elevator map: %v\n", elevatorMap)

		elevatorMapLock.Lock()
		for id, elev := range elevatorMap {
			if utils.Contains(lostCopy, id) {
				continue
			}

			if time.Since(elev.LastActive) < 5*time.Second { //|| !elev.HasPendingOrders() {
				copy := *elev
				activeElevators[id] = &copy
			}

		}

		fmt.Printf("Active elevators: %v\n", activeElevators)
		elevatorMapLock.Unlock()

		id := strconv.Itoa(localElevator.ID)

		if len(activeElevators) == 0 {
			fmt.Printf("No active elevators")
		}

		go hallRequestAssigner(localElevator, activeElevators, id, hraExecutable, elevStateRx)
	}

}
