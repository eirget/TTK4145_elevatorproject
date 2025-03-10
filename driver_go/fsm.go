package main

import (
	"Driver_go/elevio"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

var hallRequestLock sync.Mutex
var new_order_flag bool

func fsm(elevator *Elevator,
	elevators map[string]*Elevator,
	id string,
	hallRequests [][2]bool,
	cabRequests []bool,
	hraExecutable string,
	elevStateTx chan Elevator,
	req_chan chan elevio.ButtonEvent,
	new_floor_chan chan int,
	obstr_chan chan bool,
	stop_chan chan bool,
	hra_chan chan bool,
	number_of_floors int) {

	doorTimer := time.NewTimer(0)
	<-doorTimer.C

	for {
		select {
		case a := <-req_chan: //se fsm_onRequestButtonPress i fsm.c

			fmt.Printf("%+v\n", a)
			//lock
			elevator.Orders[a.Floor][a.Button].State = true
			//elevators[id] = elevator
			elevator.Orders[a.Floor][a.Button].Timestamp = time.Now()
			//unlock
			fmt.Println("orders after button press: ", elevator.Orders)

			if a.Button == BT_Cab {
				elevio.SetButtonLamp(a.Button, a.Floor, true)
			}
			new_order_flag = true
			elevStateTx <- *elevator
			//før vi kjører hall_request_assigner så må alle i elevators ha samme hall_call states

			//when we get a hall_call, broadcast message that makes all elevators run hall_request assigner
			//hra_chan <- true
			//fmt.Printf("%+v\n", elevator.Orders)

		case <-time.After(100 * time.Millisecond):

			//fmt.Println("Elevator behavior: ", elevator.Behavior)

			if elevator.Behavior == EB_Idle {
				dirn, newBehavior := elevator.chooseDirection()
				//fmt.Println("chooseDirection said: ", newBehavior)

				switch newBehavior {
				case EB_Moving:
					elevator.Direction = dirn
					elevator.Behavior = EB_Moving
					elevio.SetMotorDirection(elevator.Direction)
					fmt.Println("Elevator started moving:", elevator.Direction)
				case EB_DoorOpen:
					elevator.Behavior = EB_DoorOpen
					elevio.SetDoorOpenLamp(true)
					elevator.clearAtCurrentFloor()
					doorTimer.Reset(5 * time.Second)

				}
			}

		case a := <-new_floor_chan:
			elevator.Floor_nr = a
			elevio.SetFloorIndicator(a)

			if elevator.shouldStop() {
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevator.clearAtCurrentFloor() //timestamps are updated here
				

				elevator.Behavior = EB_DoorOpen
				elevio.SetDoorOpenLamp(true)

				doorTimer.Reset(3 * time.Second)
			}

			//new_order_flag = true    //correct to have this inside if?
			fmt.Println("timestamp that will be broadcasted after clear(hall_down): ", elevator.Orders[a][BT_HallDown].Timestamp)
			fmt.Println("timestamp that will be broadcasted after clear(hall up): ", elevator.Orders[a][BT_HallUp].Timestamp)
			fmt.Println("timestamp that will be broadcasted after clear(cab): ", elevator.Orders[a][BT_Cab].Timestamp)
			elevStateTx <- *elevator  //this SHOULD sed over data with new time stamps before door_timer

		case <-doorTimer.C:

			if elevator.Obstruction {
				fmt.Println("Waiting for obstruction to clear...")
				doorTimer.Reset(500 * time.Millisecond)
			} else {
				elevio.SetDoorOpenLamp(false)
				elevator.Behavior = EB_Idle
				// Remove first floor from Orders
				// Turn off floor button light
				elevator.Direction, elevator.Behavior = elevator.chooseDirection()
				elevio.SetMotorDirection(elevator.Direction)

				fmt.Println("Resuming movement...")
				fmt.Println("Resuming with Orders:")
				fmt.Printf("%+v\n", elevator.Orders)
			}

		case a := <-obstr_chan:
			fmt.Printf("%+v\n", a)
			elevator.Obstruction = a

		case a := <-stop_chan:
			fmt.Printf("%+v\n", a)
			elevio.SetStopLamp(true)
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevator.Behavior = EB_Idle
			elevator.Direction = elevio.MD_Stop
			elevator.Orders = [4][3]OrderType{}
			fmt.Printf("%+v\n", elevator.Orders)
			for f := 0; f < number_of_floors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
			time.Sleep(5 * time.Second)
			elevio.SetStopLamp(false)
		}
	}

}

func fsm_hallRequestAssigner(elevator *Elevator,
	elevators map[string]*Elevator,
	id string,
	hallRequests [][2]bool,
	cabRequests []bool,
	hraExecutable string,
	elevStateTx chan Elevator) {

	hallRequestLock.Lock()
	//elevStateTx <- *elevator  //kanksje alt vi trenger å gjøre for å broadcaste vår state
	elevators[id] = elevator

	for i := 0; i < NumFloors; i++ {
		hallRequests[i][0] = elevator.Orders[i][0].State
		hallRequests[i][1] = elevator.Orders[i][1].State
	}
	input := HRAInput{
		HallRequests: hallRequests,
		States:       make(map[string]HRAElevState),
	}

	for peerID, elev := range elevators {
		for f := 0; f < NumFloors; f++ {
			cabRequests[f] = elev.Orders[f][2].State
		}
		input.States[peerID] = HRAElevState{
			Behavior:    behaviorMap[elev.Behavior],
			Floor:       elev.Floor_nr,
			Direction:   directionMap[elev.Direction],
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

	fmt.Println("Raw HRA output: ", string(ret))

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
		fmt.Println("peerID: ", peerID)
		assignedID, _ := strconv.Atoi(peerID)
		fmt.Println("assignedID: ", assignedID)
		for i_id := range elevators {
			for f := 0; f < NumFloors; f++ {
				//elevators[i_id].Orders[f][0].State = newRequests[f][0]
				//elevators[i_id].Orders[f][0].Timestamp = time.Now()
				//elevators[i_id].Orders[f][1].State = newRequests[f][1]
				//elevators[i_id].Orders[f][1].Timestamp = time.Now()
				
				if newRequests[f][0] {
					fmt.Println("if happened")
					fmt.Println("assignedID: ", assignedID)
					elevators[i_id].Orders[f][0].ElevatorID = assignedID
					elevators[i_id].Orders[f][0].Timestamp = time.Now()
					//is it enough to only change the timestaps if the if's happens
					//fmt.Println("the actual ID of elev now:", elevators[i_id].Orders[f][0].ElevatorID)
				}
				if newRequests[f][1] {
					fmt.Println("if 2 happened")
					fmt.Println("assignedID: ", assignedID)
					elevators[i_id].Orders[f][1].ElevatorID = assignedID
					elevators[i_id].Orders[f][1].Timestamp = time.Now()
				}

				//update timestamp in the end?
			}
			// Important: Reassign modified struct back into the map
			
		}
		//er cab calls det de skal være nå?
		elevator.Orders = elevators[id].Orders
	}
	fmt.Printf("Orders after hra: %+v\n", elevator.Orders) //!!her burde det være endringer

	elevStateTx <- *elevator

	hallRequestLock.Unlock()
}
