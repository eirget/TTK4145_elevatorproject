package main

import (
	"Driver_go/config"
	"Driver_go/elevator"
	"Driver_go/elevio"
	"Driver_go/network"
	"flag"
	"fmt"
	"strconv"
	"time"
)

//tested to obstruct the closest elevator to a call, the available elevator did not take over the order

func main() {

	//NETWORK

	hraExecutable := "hall_request_assigner"

	//make this automatic later
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	var port string
	flag.StringVar(&port, "port", "", "port of this peer")
	flag.Parse()

	hallRequests := make([][2]bool, config.NumFloors)
	cabRequests := make([]bool, config.NumFloors)

	peerUpdateCh, elevStateTx, elevStateRx, run_hra, receive_run_hra := network.NetworkInit(id)

	//create map to store elevator states for all elevators on system, !!! point to discuss: *Elevator or not?
	elevators := make(map[string]*elevator.Elevator)

	fmt.Printf("%+v\n", elevators)

	addr := "localhost:" + port
	elevio.Init(addr, config.NumFloors) //gjør til et flag

	var d elevio.MotorDirection = elevio.MD_Up
	//

	fmt.Printf("Before go routines \n")

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	for f := 0; f < config.NumFloors; f++ {
		for b := elevio.ButtonType(0); b < 3; b++ {
			elevio.SetButtonLamp(b, f, false)
		}
	}
	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)

	//this go routine is made to make the code work when the elvator is initalized between floors. For this we need a new chan (floorChan) because using drv_floors might cause race conditions
	floorChan := make(chan int)

	go func() {
		elevio.SetMotorDirection(d)

		for {
			select {
			case a := <-drv_floors:
				if a != -1 {
					fmt.Println("Started at floor: ", a)
					elevio.SetMotorDirection(elevio.MD_Stop)
					floorChan <- a
					return
				}
			case <-time.After(500 * time.Millisecond):
				fmt.Println("Waiting for valid floor signal...")
			}
		}
		
	}()
	a := <-floorChan
	fmt.Println("Elevator initalized at floor: ", a)


	id_num, _ := strconv.Atoi(id)
	elevator := elevator.ElevatorInit(a, id_num) //kanskje det blir penere å bare bruke string


	elevStateTx <- *elevator

	elevio.SetFloorIndicator(elevator.Floor_nr)

	go fsm(elevator, elevStateTx, drv_buttons, drv_floors, drv_obstr, drv_stop, config.NumFloors)
	//go hraSignalListener(elevator, elevators, id, hallRequests, cabRequests, hraExecutable, elevStateTx, run_hra)

	//non-blocking timer
	hraTimer := time.NewTimer(0)
	<-hraTimer.C
	hraTimer.Reset(1 * time.Second)

	fmt.Printf("Started elevator system \n")

	//TODO: NEED TO BROADCAST ELEVATOR STATES OFTEN, or just when changed

	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			//fixed elevators knowing of eachother even without an event happening
			if len(p.New) != 0 {
				elevStateTx <- *elevator
			}

		//heartbeat check functionality in below case as well, time each ID, maybe peers is good enough already
		case a := <-elevStateRx: //kan hende denne vil miste orders om det blir fullt i buffer
			//update elevator to have newest state of other elevators
			idStr := strconv.Itoa(a.ID)

			elevators[idStr] = &a //may have to directly allocate new Elevator pointer

			fmt.Printf("Recieved: \n")
			fmt.Printf("Message from ID: %v\n", a.Orders[1][2].ElevatorID)
			//fmt.Printf("Floor_nr: %v\n", a.Floor_nr)
			//fmt.Printf("Direction %v\n", a.Direction)
			//fmt.Println("timestamp(hall up): \n", a.Orders[a.Floor_nr][BT_HallUp].Timestamp)

			//NEW, idea for fixing when hall requests should actually be updated
			//func updateHallRequests(myElevator *Elevator, receivedElev Elevator) {
			//if idStr != id {
			for f := 0; f < config.NumFloors; f++ {
				for b := 0; b < config.NumButtons-1; b++ { // Only HallUp and HallDown
					// Compare timestamps to ensure only newer updates are accepted
					if a.Orders[f][b].Timestamp.After(elevator.Orders[f][b].Timestamp) {
						elevator.Orders[f][b] = a.Orders[f][b]
					}
				}
			}
			//}

			for f := 0; f < config.NumFloors; f++ {
				fmt.Printf("\n Floornr: %+v ", f)
				for b := elevio.ButtonType(0); b < 3; b++ {
					fmt.Printf("%+v ", a.Orders[f][b].State)
					fmt.Printf("%+v, ", a.Orders[f][b].ElevatorID)
				}

			}
			fmt.Printf("\n")
			fmt.Printf("Timestamp: %v \n", time.Now())

			if new_order_flag {
				run_hra <- true
				new_order_flag = false
			}

			//}

			//fmt.Printf("%+v\n", elevators)

		case <-receive_run_hra:
			go hallRequestAssigner(elevator, elevators, id, hallRequests, cabRequests, hraExecutable, elevStateRx)

			//might not be neccessary at all
			//case <-time.After(500 * time.Millisecond):
			//	elevStateTx <- *elevator
		}
	}
}
