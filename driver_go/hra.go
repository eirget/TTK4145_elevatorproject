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
	Behavior    string `json:"behavior"` //behaviour?
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

func hallRequestAssigner(elev *elevator.Elevator,
	elevatorMap map[string]*elevator.Elevator,
	activeElevators map[string]*elevator.Elevator,
	id string,
	hraExecutable string,
	elevStateTx chan elevator.Elevator) {

	fmt.Printf("HRA started\n")

	hallRequests := make([][2]bool, config.NumFloors)
	//cabRequests := make([]bool, config.NumFloors)

	hallRequestLock.Lock()
	//activeElevators[id] = elev //?
	//elevatorMap[id] = elev

	//tror probleme ligger under her

	for i := 0; i < config.NumFloors; i++ {
		hallRequests[i][0] = elev.Orders[i][0].State
		hallRequests[i][1] = elev.Orders[i][1].State
	}

	// Run the normal hall request assignment process
	input := HRAInput{
		HallRequests: hallRequests,
		States:       make(map[string]HRAElevState),
	}

	for peerID, e := range activeElevators {
		individualCabRequests := make([]bool, config.NumFloors)
		for f := 0; f < config.NumFloors; f++ {
			individualCabRequests[f] = e.Orders[f][2].State //SE NØYE PÅ DETTE
		}
		fmt.Printf("Elevator behaviour of id %s: %v\n", peerID, e.Behavior)
		input.States[peerID] = HRAElevState{
			Behavior:    elevator.BehaviorMap[e.Behavior],
			Floor:       e.Floor_nr,
			Direction:   elevator.DirectionMap[e.Direction],
			CabRequests: individualCabRequests,
		}
	}

	hallRequestLock.Unlock()

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		return
	}

	fmt.Println("HRA JSON Input:", string(jsonBytes)) //DEBUG

	ret, err := exec.Command("../"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		return
	}

	// Process the output and update orders

	/*
		output := new(map[string][][2]bool)
		err = json.Unmarshal(ret, &output)
		if err != nil {
			fmt.Println("json.Unmarshal error ", err)
			return
		}
	*/

	var output map[string][][2]bool //DEBUG
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error ", err)
		return
	}

	hallRequestLock.Lock()

	for peerID, newRequests := range output {
		assignedID, _ := strconv.Atoi(peerID)

		fmt.Println("New requests from HRA:", newRequests) //DEBUG

		/*
			for i_id := range activeElevators {
				for f := 0; f < config.NumFloors; f++ {

					if newRequests[f][0] {
						fmt.Printf("if 1 happened \n")
						activeElevators[i_id].Orders[f][0].ElevatorID = assignedID
						activeElevators[i_id].Orders[f][0].Timestamp = time.Now()
					}
					if newRequests[f][1] {
						fmt.Printf("if 2 happened \n")
						activeElevators[i_id].Orders[f][1].ElevatorID = assignedID
						activeElevators[i_id].Orders[f][1].Timestamp = time.Now()
					}
				}
			}
			elev.Orders = activeElevators[id].Orders
		*/ //dont do this
		e, exists := activeElevators[peerID]
		if !exists || e == nil {
			fmt.Printf("Warning: No active elevator with ID %s\n", peerID)
			continue
		}

		/*
			for f := 0; f < config.NumFloors; f++ {
				if newRequests[f][0] {
					e.Orders[f][0].ElevatorID = assignedID
					e.Orders[f][0].Timestamp = time.Now()
				}
				if newRequests[f][1] {
					e.Orders[f][1].ElevatorID = assignedID
					e.Orders[f][1].Timestamp = time.Now()
				}
			}
		*/
		updateHallOrders(e, newRequests, assignedID)

		if peerID == id {
			//reflect updated orders in this local elevator's view
			for f := 0; f < config.NumFloors; f++ {
				elev.Orders[f][0] = e.Orders[f][0]
				elev.Orders[f][1] = e.Orders[f][1]
			}
		}
	}
	// Notify FSM
	elevStateTx <- *elev
	hallRequestLock.Unlock()
}

func updateHallOrders(e *elevator.Elevator, newRequests [][2]bool, assignedID int) {

	now := time.Now()

	for floor := 0; floor < config.NumFloors; floor++ {
		for btn := 0; btn <= 1; btn++ {
			current := &e.Orders[floor][btn]

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

			/*
				if newRequests[floor][btn] {
					e.Orders[floor][btn].State = true
					e.Orders[floor][btn].ElevatorID = assignedID
					e.Orders[floor][btn].Timestamp = time.Now()
				} else if e.Orders[floor][btn].ElevatorID == assignedID {
					continue
				} else {
					//set to false only if it was previously assigned to another
					e.Orders[floor][btn].State = false
					e.Orders[floor][btn].ElevatorID = 100
					e.Orders[floor][btn].Timestamp = time.Now()
				}
			*/
		}
	}
}
