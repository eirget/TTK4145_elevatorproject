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
	elevators map[string]*elevator.Elevator,
	id string,
	hallRequests [][2]bool,
	cabRequests []bool,
	hraExecutable string,
	elevStateTx chan elevator.Elevator) {

	fmt.Printf("HRA started1\n")

	hallRequestLock.Lock()
	elevators[id] = elev

	// Remove stuck elevators from consideration
	activeElevators := make(map[string]*elevator.Elevator)
	for peerID, e := range elevators {
		if time.Since(e.LastActive) < 5*time.Second { // ✅ Only include active elevators
			activeElevators[peerID] = e
		}
	}

	// Run the normal hall request assignment process
	input := HRAInput{
		HallRequests: hallRequests,
		States:       make(map[string]HRAElevState),
	}

	for peerID, e := range activeElevators {
		for f := 0; f < config.NumFloors; f++ {
			cabRequests[f] = e.Orders[f][2].State
		}
		input.States[peerID] = HRAElevState{
			Behavior:    elevator.BehaviorMap[e.Behavior],
			Floor:       e.Floor_nr,
			Direction:   elevator.DirectionMap[e.Direction],
			CabRequests: cabRequests,
		}
	}

	
	hallRequestLock.Unlock()

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		return
	}

	ret, err := exec.Command("../"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		return
	}

	// Process the output and update orders
	hallRequestLock.Lock()

	//output er tom, så når den oppdateres blir den kun fylt med false states vi må altså somehow sørge for at det som står i output er det samme som står i elevator

	output := new(map[string][][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error ", err)
		return
	}


	
	for peerID, newRequests := range *output {
		assignedID, _ := strconv.Atoi(peerID)
		for i_id := range activeElevators {
			for f := 0; f < config.NumFloors; f++ {
				// feilen er at denne alrdri blir true fordi newRequests ikke oppdateres riktig 
				if newRequests[f][0] {
					fmt.Printf("if 1 happened")
					activeElevators[i_id].Orders[f][0].ElevatorID = assignedID
					activeElevators[i_id].Orders[f][0].Timestamp = time.Now()
				}
				if newRequests[f][1] {
					fmt.Printf("if 2 happened")
					activeElevators[i_id].Orders[f][1].ElevatorID = assignedID
					activeElevators[i_id].Orders[f][1].Timestamp = time.Now()
				}
			}
		}
		elev.Orders = activeElevators[id].Orders
	}
	
	// Notify FSM
	elevStateTx <- *elev
	hallRequestLock.Unlock()
}
