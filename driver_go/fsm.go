package main

import (
	"Driver_go/elevio"
	"fmt"
	"time"
)

func fsm(elevator *Elevator,
	req_chan chan elevio.ButtonEvent,
	new_floor_chan chan int,
	obstr_chan chan bool,
	stop_chan chan bool,
	number_of_floors int) {

	doorTimer := time.NewTimer(0)
	<-doorTimer.C

	for {
		select{
		case a := <-req_chan:
			fmt.Printf("%+v\n", a)
			elevator.Orders[a.Button][a.Floor].State = true
			elevator.Orders[a.Button][a.Floor].ElevatorID = 1 //må bli gitt et annet sted senere

			fmt.Printf("%+v\n", elevator.Orders)


		case a := <- new_floor_chan:
			elevator.Floor_nr = a
			elevio.SetFloorIndicator(a)


		case a := <- obstr_chan:
			fmt.Printf("%+v\n", a)
			elevator.Obstruction = a

		case a := <- stop_chan:
			fmt.Printf("%+v\n", a)
			elevio.SetStopLamp(true)
			elevator.Orders = [3][4]OrderType{}
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