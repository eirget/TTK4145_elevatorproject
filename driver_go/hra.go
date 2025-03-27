package main

import (
	"Driver_go/config"
	"Driver_go/elevator"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

var hallRequestLock sync.Mutex

type HRAElevState struct {
	Behavior    string `json:"behavior"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

func hallRequestAssigner(
	elev *elevator.Elevator,
	activeElevators map[string]*elevator.Elevator,
	id string,
	hraExecutable string,
	elevStateTx chan elevator.Elevator) {

	fmt.Printf("HRA started\n")

	input := buildHRAInput(elev, activeElevators)

	output, err := runHraProcess(input, hraExecutable)

	if err != nil {
		fmt.Printf("Failed to run hra process: %v\n", err)
		return
	}

	hallRequestLock.Lock()
	for peerID, newRequests := range output {
		assignedID, _ := strconv.Atoi(peerID)

		fmt.Println("New requests from HRA:", newRequests) //MAYBE REMOVE IN THE END

		updateHallOrders(peerID, elev, newRequests, assignedID, activeElevators, id)
	}
	hallRequestLock.Unlock()
	elevStateTx <- *elev
}

func buildHRAInput(
	elev *elevator.Elevator,
	activeElevators map[string]*elevator.Elevator) HRAInput {

	hallRequestLock.Lock()
	defer hallRequestLock.Unlock()

	hallRequests := make([][2]bool, config.NumFloors)

	for i := 0; i < config.NumFloors; i++ {
		hallRequests[i][0] = elev.Orders[i][0].State
		hallRequests[i][1] = elev.Orders[i][1].State
	}

	input := HRAInput{
		HallRequests: hallRequests,
		States:       make(map[string]HRAElevState),
	}

	for peerID, activeElev := range activeElevators {
		individualCabRequests := make([]bool, config.NumFloors)
		for floor := 0; floor < config.NumFloors; floor++ {
			individualCabRequests[floor] = activeElev.Orders[floor][2].State
		}
		input.States[peerID] = HRAElevState{
			Behavior:    elevator.BehaviorMap[activeElev.Behavior],
			Floor:       activeElev.FloorNr,
			Direction:   elevator.DirectionMap[activeElev.Direction],
			CabRequests: individualCabRequests,
		}
	}
	return input
}

func runHraProcess(input HRAInput, hraExecutable string) (map[string][][2]bool, error) {

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		return nil, err
	}

	ret, err := exec.Command("hall_request_assigner/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		return nil, err
	}

	var output map[string][][2]bool
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error ", err)
		return nil, err
	}
	return output, nil
}

func updateHallOrders(
	peerID string, elev *elevator.Elevator,
	newRequests [][2]bool, assignedID int,
	activeElevators map[string]*elevator.Elevator,
	localID string) {

	activeElev, exists := activeElevators[peerID]
	if !exists || activeElev == nil {
		fmt.Printf("Warning: No active elevator with ID %s\n", peerID)
		return
	}
	now := time.Now()
	for floor := 0; floor < config.NumFloors; floor++ {
		for btn := 0; btn <= 1; btn++ {
			current := &activeElev.Orders[floor][btn]

			if newRequests[floor][btn] {
				//update only if this is a new assignment or state was false
				if !current.State || current.ElevatorID != assignedID {
					current.State = true
					current.ElevatorID = assignedID
					current.Timestamp = now
				}
			} else if current.ElevatorID == assignedID {
				//do not clear it, preserve it
				continue
			}
		}
	}
	if peerID == localID {
		//reflect updated orders in this local elevator's view
		for f := 0; f < config.NumFloors; f++ {
			elev.Orders[f][0] = activeElev.Orders[f][0]
			elev.Orders[f][1] = activeElev.Orders[f][1]
		}
	}
}
