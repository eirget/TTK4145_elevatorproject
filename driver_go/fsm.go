package main

import (
	"Driver_go/config"
	"Driver_go/elevator"
	"Driver_go/elevio"
	"fmt"
	"sync"
	"time"
)

var hallRequestLock sync.Mutex
var new_order_flag bool

func fsm(elev *elevator.Elevator,
	elevStateTx chan elevator.Elevator,
	req_chan chan elevio.ButtonEvent,
	new_floor_chan chan int,
	obstr_chan chan bool,
	stop_chan chan bool,
	number_of_floors int) {

	doorTimer := time.NewTimer(0)
	<-doorTimer.C

	for {
		select {
		case a := <-req_chan: //se fsm_onRequestButtonPress i fsm.c

			fmt.Printf("%+v\n", a)
			//lock
			elev.Orders[a.Floor][a.Button].State = true
			//elevators[id] = elevator
			elev.Orders[a.Floor][a.Button].Timestamp = time.Now()
			//unlock
			fmt.Println("orders after button press: ", elev.Orders) //this print hapened at a somewhat random time? when is it supposed to happen?

			if a.Button == elevio.BT_Cab {
				elevio.SetButtonLamp(a.Button, a.Floor, true)
			}
			new_order_flag = true
			elevStateTx <- *elev
			//før vi kjører hall_request_assigner så må alle i elevators ha samme hall_call states

			//when we get a hall_call, broadcast message that makes all elevators run hall_request assigner
			//hra_chan <- true
			//fmt.Printf("%+v\n", elevator.Orders)

		case <-time.After(100 * time.Millisecond):

			//fmt.Println("Elevator behavior: ", elevator.Behavior)

			if elev.Behavior == elevator.EB_Idle {
				dirn, newBehavior := elev.ChooseDirection()
				//fmt.Println("chooseDirection said: ", newBehavior)

				switch newBehavior {
				case elevator.EB_Moving:
					elev.Direction = dirn
					elev.Behavior = elevator.EB_Moving
					elevio.SetMotorDirection(elev.Direction)
					fmt.Println("Elevator started moving:", elev.Direction)
				case elevator.EB_DoorOpen:
					elev.Behavior = elevator.EB_DoorOpen
					elevio.SetDoorOpenLamp(true)
					elev.ClearAtCurrentFloor()
					doorTimer.Reset(5 * time.Second)

				}
			}

		case a := <-new_floor_chan:
			elev.Floor_nr = a
			elevio.SetFloorIndicator(a)

			if elev.ShouldStop() {
				elevio.SetMotorDirection(elevio.MD_Stop)
				elev.ClearAtCurrentFloor() //timestamps are updated here

				elev.Behavior = elevator.EB_DoorOpen
				elevio.SetDoorOpenLamp(true)

				doorTimer.Reset(3 * time.Second)
			}

			//new_order_flag = true    //correct to have this inside if?

			elevStateTx <- *elev //this SHOULD sed over data with new time stamps before door_timer

		case <-doorTimer.C:

			if elev.Obstruction {
				fmt.Println("Waiting for obstruction to clear...")
				doorTimer.Reset(500 * time.Millisecond)

				//unsure whether the following code should be here or in the obstr_chan case
				// the following go routine might made just because we want a non-blocking timer, but code for this is already made in main.go line 101. Meaning the go routine might be unnessesary
				go func() { //start a timer in a separate go routine
					time.Sleep(2 * time.Second) //if obstruction is on for more than 2 seconds: reassign orders

					if elev.Obstruction {
						elevator.ReassignOrders(elev)
					} else {
						fmt.Println("Obstruction cleared")
					}
				}()
				
			} else {
				elevio.SetDoorOpenLamp(false)
				elev.Behavior = elevator.EB_Idle
				// Remove first floor from Orders
				// Turn off floor button light
				elev.Direction, elev.Behavior = elev.ChooseDirection()
				elevio.SetMotorDirection(elev.Direction)

				fmt.Println("Resuming movement...")
				fmt.Println("Resuming with Orders:")
				for f := 0; f < config.NumFloors; f++ {
					fmt.Printf("\n Floornr: %+v ", f)
					for b := elevio.ButtonType(0); b < 3; b++ {
						fmt.Printf("%+v ", elev.Orders[f][b].State)
						fmt.Printf("%+v ", elev.Orders[f][b].ElevatorID)
					}

				}
				fmt.Printf("\n")
			}

		case a := <-obstr_chan:
			fmt.Printf("%+v\n", a)
			elev.Obstruction = a

		case a := <-stop_chan:
			fmt.Printf("%+v\n", a)
			elevio.SetStopLamp(true)
			elevio.SetMotorDirection(elevio.MD_Stop)
			elev.Behavior = elevator.EB_Idle
			elev.Direction = elevio.MD_Stop
			elev.Orders = [4][3]elevator.OrderType{}
			fmt.Printf("%+v\n", elev.Orders)
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
