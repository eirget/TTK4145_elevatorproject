package main

import (
	"Driver_go/elevio"
	"Driver_go/network/bcast"
	"Driver_go/network/peers"
	"encoding/json"
	"flag"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

type HelloMsg struct {
	Message string
	Iter 	int
}

type HRAElevState struct {
	Behavior string `json:"behavior"` //behaviour?
	Floor	int `json:"floor"`
	Direction	string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool `json:"hallRequests"`
	States map[string]HRAElevState `json:"states"`
}

const(
	NumFloors = 4
	NumButtons = 3
)


func main() {
	
	//NETWORK

	hraExecutable := "hall_reqest_assigner"

	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()


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

	elevio.Init("localhost:15657", NumFloors) //gjør til et flag

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

	elevio.SetFloorIndicator(elevator.Floor_nr)

	go fsm(elevator, drv_buttons, drv_floors, drv_obstr, drv_stop, NumFloors)
	

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

	fmt.Printf("Started elevator system")



	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-elevStateRx:  //kan hende denne vil miste orders om det blir fullt i buffer
			fmt.Printf("Recieved: \n")
			fmt.Printf("Message from ID: %v\n", a.Orders[1][2].ElevatorID)
			fmt.Printf("Floor_nr: %v\n", a.Floor_nr)
			fmt.Printf("Direction %v\n", a.Direction)
		
		case <- hraTimer.C:
			elevStateTx <- *elevator  //kanksje alt vi trenger å gjøre for å broadcaste vår state

			hraTimer.Reset(1 * time.Second)
			var hallRequests [][2]bool
			var cabRequests []bool 
			for i := 0; i < NumFloors; i++ {
				hallRequests[i][0] = elevator.Orders[i][0].State
				hallRequests[i][1] = elevator.Orders[i][1].State
				cabRequests[i] = elevator.Orders[i][2].State
			}
			input := HRAInput {
				HallRequests: hallRequests,
				States: make(map[string]HRAElevState),
			}

			input.States[id] = HRAElevState{
				Behavior: behaviorMap[elevator.Behavior],
				Floor: elevator.Floor_nr,
				Direction: directionMap[elevator.Direction],
				CabRequests: cabRequests,
			}

			jsonBytes, err := json.Marshal(input)
			if err != nil {
				fmt.Println("json.Marshal error: ", err)
				return
			}

			ret, err := exec.Command("../hall_request_assigner/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
			if err != nil {
				fmt.Println("exec.Command error: ", err)
				fmt.Println(string(ret))
				return
			}

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
		}

	}


}
