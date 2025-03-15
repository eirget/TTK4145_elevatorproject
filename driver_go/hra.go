package main

import (
	"Driver_go/config"
	"Driver_go/elevator"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

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

	hallRequestLock.Lock()
	//elevStateTx <- *elevator  //kanksje alt vi trenger å gjøre for å broadcaste vår state
	elevators[id] = elev

	for i := 0; i < config.NumFloors; i++ {
		hallRequests[i][0] = elev.Orders[i][0].State
		hallRequests[i][1] = elev.Orders[i][1].State
	}
	input := HRAInput{
		HallRequests: hallRequests,
		States:       make(map[string]HRAElevState),
	}

	for peerID, elev := range elevators {
		for f := 0; f < config.NumFloors; f++ {
			cabRequests[f] = elev.Orders[f][2].State
		}
		input.States[peerID] = HRAElevState{
			Behavior:    elevator.BehaviorMap[elev.Behavior],
			Floor:       elev.Floor_nr,
			Direction:   elevator.DirectionMap[elev.Direction],
			CabRequests: cabRequests,
		}
	}

	hallRequestLock.Unlock()

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		return
	}

	fmt.Println("Length of hallreqs: ", len(hallRequests))
	fmt.Println("Length of cabreqs: ", len(cabRequests))

	//maybe need whole path
	ret, err := exec.Command("../"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		return
	}

	hallRequestLock.Lock()

	output := new(map[string][][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error ", err)
		return
	}

	fmt.Printf("output: \n")
	for k, v := range *output {
		fmt.Printf("%6v : %+v\n", k, v)
	}

	for peerID, newRequests := range *output {
		assignedID, _ := strconv.Atoi(peerID)
		for i_id := range elevators {
			for f := 0; f < config.NumFloors; f++ {
				//elevators[i_id].Orders[f][0].State = newRequests[f][0]
				//elevators[i_id].Orders[f][0].Timestamp = time.Now()
				//elevators[i_id].Orders[f][1].State = newRequests[f][1]
				//elevators[i_id].Orders[f][1].Timestamp = time.Now()

				if newRequests[f][0] {
					elevators[i_id].Orders[f][0].ElevatorID = assignedID
					elevators[i_id].Orders[f][0].Timestamp = time.Now()
					//is it enough to only change the timestaps if the if's happens
					//fmt.Println("the actual ID of elev now:", elevators[i_id].Orders[f][0].ElevatorID)
				}
				if newRequests[f][1] {
					elevators[i_id].Orders[f][1].ElevatorID = assignedID
					elevators[i_id].Orders[f][1].Timestamp = time.Now()
				}

				//update timestamp in the end?
			}
			// Important: Reassign modified struct back into the map

		}
		//er cab calls det de skal være nå?
		elev.Orders = elevators[id].Orders
	}
	//fmt.Printf("Orders after hra: %+v\n", elevator.Orders) //!!her burde det være endringer

	elevStateTx <- *elev

	hallRequestLock.Unlock()
}
