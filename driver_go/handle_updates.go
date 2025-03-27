package main

import (
	"Driver_go/config"
	"Driver_go/elevator"
	"Driver_go/network/peers"
	"Driver_go/utils"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// update elevator to have newest state of other elevators
func handleElevatorUpdates(
	elev *elevator.Elevator,
	elevStateRx <-chan elevator.Elevator,
	elevatorMap map[string]*elevator.Elevator,
	runHraCh chan struct{},
	newOrderCh chan struct{}) {

	for elevRx := range elevStateRx {

		elevatorMapLock.Lock()
		elevRxCopy := elevRx
		idStr := strconv.Itoa(elevRx.ID)
		elevatorMap[idStr] = &elevRxCopy

		if elevRx.ID == elev.ID {
			for floor := 0; floor < config.NumFloors; floor++ {
				for btn := 0; btn < config.NumButtons; btn++ {
					if elevRx.Orders[floor][btn].Timestamp.After(elev.Orders[floor][btn].Timestamp) {
						elev.Orders[floor][btn] = elevRx.Orders[floor][btn]
					}
				}
			}
		}

		elevatorMapLock.Unlock()

		for floor := 0; floor < config.NumFloors; floor++ {
			for btn := 0; btn < config.NumHallButtons; btn++ {
				// Compare timestamps to ensure only newer updates are accepted
				if elevRx.Orders[floor][btn].Timestamp.After(elev.Orders[floor][btn].Timestamp) {
					elev.Orders[floor][btn] = elevRx.Orders[floor][btn]
				}
			}
		}

		elev.SetLights()

		select {
		case <-newOrderCh:
			select {
			case runHraCh <- struct{}{}:
			default:
			}
		default:
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
				elevatorMap[peerUpdate.New].Orders = lastState.Orders
				elevStateTx <- *elevatorMap[peerUpdate.New]
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
		fmt.Println("Received runHra signal")

		disconnectedMutex.RLock()
		isDisconnected := *disconnectedFromNetwork
		disconnectedMutex.RUnlock()

		if isDisconnected {
			fmt.Println("Not connected: assigning all hall requests to self")
			elevator.AssignAllHallCallsToSelf(localElevator)
			elevStateTx <- *localElevator
			continue
		}

		latestLostMutex.Lock()
		lostCopy := append([]string(nil), (*latestLost)...)
		latestLostMutex.Unlock()

		activeElevators := make(map[string]*elevator.Elevator)

		elevatorMapLock.Lock()
		for id, elev := range elevatorMap {
			if utils.Contains(lostCopy, id) {
				continue
			}

			if time.Since(elev.LastActive) < 5*time.Second {
				elevCopy := *elev
				activeElevators[id] = &elevCopy
			}

		}

		elevatorMapLock.Unlock()

		id := strconv.Itoa(localElevator.ID)

		go hallRequestAssigner(localElevator, activeElevators, id, hraExecutable, elevStateRx)
	}
}
