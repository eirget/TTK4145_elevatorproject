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
		case a := <-req_chan:  //se fsm_onRequestButtonPress i fsm.c 
			fmt.Printf("%+v\n", a)
			elevator.Orders[a.Floor][a.Button].State = true
			elevator.Orders[a.Floor][a.Button].ElevatorID = 1 //må bli gitt et annet sted senere
			if a.Button == BT_Cab {
				elevio.SetButtonLamp(a.Button, a.Floor, true)
			}
			

			fmt.Printf("%+v\n", elevator.Orders)

		case <- time.After(100 * time.Millisecond):

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
			

		case a := <- new_floor_chan:
			elevator.Floor_nr = a
			elevio.SetFloorIndicator(a)

			if elevator.shouldStop() {
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevator.clearAtCurrentFloor()

				elevator.Behavior = EB_DoorOpen
				elevio.SetDoorOpenLamp(true)

				doorTimer.Reset(3 * time.Second)
			}
			
		case <- doorTimer.C:

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

		case a := <- obstr_chan:
			fmt.Printf("%+v\n", a)
			elevator.Obstruction = a

		case a := <- stop_chan:
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