package main

import (
	"Driver_go/elevator"
	"Driver_go/elevio"
	"fmt"
	"time"
)

// burde kanksje lages i main
var newOrderFlag bool

func fsm(e *elevator.Elevator,
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
		case newReq := <-req_chan:
			fsmHandleRequestButtonPress(newReq, e, elevStateTx, &newOrderFlag)

		case <-time.After(100 * time.Millisecond):
			fsmHandleIdleState(e, elevStateTx, doorTimer)

		case newFloor := <-new_floor_chan:
			fsmHandleNewFloor(newFloor, e, elevStateTx, doorTimer)

		case <-doorTimer.C:
			fsmHandleDoorTimeout(e, doorTimer, elevStateTx)
		//maybe name obstructionState? I misunderstood this variable name
		case isObstructed := <-obstr_chan:
			fmt.Println("Obstruction happened")
			fsmHandleObstruction(isObstructed, e)

		case isStopped := <-stop_chan:
			fsmHandleEmergencyStop(isStopped, e, number_of_floors)

			//case <-time.After(1 * time.Second):
			//	elevStateTx <- *e
		}
	}
}

func fsmHandleRequestButtonPress(a elevio.ButtonEvent, e *elevator.Elevator, elevStateTx chan elevator.Elevator, new_order_flag *bool) {
	fmt.Printf("%+v\n", a)
	e.Orders[a.Floor][a.Button].State = true
	e.Orders[a.Floor][a.Button].Timestamp = time.Now()

	*new_order_flag = true
	elevStateTx <- *e

	fmt.Println("orders after button press: ", e.Orders)

	if a.Button == elevio.BT_Cab {
		elevio.SetButtonLamp(a.Button, a.Floor, true)
	}

}

func fsmHandleIdleState(e *elevator.Elevator, elevStateTx chan elevator.Elevator, doorTimer *time.Timer) {
	if e.Behavior == elevator.EB_Idle {
		//this will probably never be added again
		/*
				if e.JustStopped {
					e.JustStopped = false
					return
				}

			if e.ShouldStop() {
				fmt.Println("Reopening at same floor to serve additional order")
				e.StopAtFloor()
				doorTimer.Reset(3 * time.Second)
				return
			}
		*/
		e.HandleIdleState()
		/*
			if e.Direction != elevio.MD_Stop {
				e.LastDirection = e.Direction
			}
		*/
		if e.Behavior == elevator.EB_DoorOpen {
			doorTimer.Reset(3 * time.Second)
		}
	}
	elevStateTx <- *e
}

func fsmHandleNewFloor(a int, e *elevator.Elevator, elevStateTx chan elevator.Elevator, doorTimer *time.Timer) {
	e.Floor_nr = a
	elevio.SetFloorIndicator(a)

	//ShouldStop returns true if it has pending orders in its current direction, or if you just in general have no reason to continue in your current direction
	if e.ShouldStop() {
		e.StopAtFloor()
		//after StopAtFloor we have MD_Stop and EB_DoorOpen
		doorTimer.Reset(3 * time.Second)
	}

	elevStateTx <- *e
}

func fsmHandleDoorTimeout(e *elevator.Elevator, doorTimer *time.Timer, elevStateTx chan elevator.Elevator) {
	//i think here we have MD_Stop and EB_DoorOpen
	if e.ShouldReopenForOppositeHallCall() {
		fmt.Printf("Yes, should reopen for opposite hall call")

		/* this should not be neccessary anymore but idk
		if e.LastDirection == elevio.MD_Up {
			e.LastDirection = elevio.MD_Down
		} else if e.LastDirection == elevio.MD_Down {
			e.LastDirection = elevio.MD_Up
		}
		*/
		//e.Direction = e.LastDirection //maybe we should still have this
		e.StopAtFloor()
		doorTimer.Reset(3 * time.Second)
		return
	}
	if e.Obstruction {
		fmt.Println("Waiting for obstruction to clear...")
		doorTimer.Reset(500 * time.Millisecond)
		return
	}
	fmt.Println("Close door and resume called")
	e.CloseDoorAndResume()
	//after close door and resume we have whatever choose direction says, maybe it should just return idle so that handleIdleState fixes it?
	elevStateTx <- *e
}

func fsmHandleObstruction(isObstructed bool, e *elevator.Elevator) {
	e.Obstruction = isObstructed

	if isObstructed {
		fmt.Println("Obstruction detected! Waiting for it to clear...")
	}
}

func fsmHandleEmergencyStop(a bool, e *elevator.Elevator, number_of_floors int) {
	fmt.Printf("%+v\n", a)
	elevio.SetStopLamp(true)
	elevio.SetMotorDirection(elevio.MD_Stop)
	e.Behavior = elevator.EB_Idle
	e.Direction = elevio.MD_Stop
	e.Orders = [4][3]elevator.OrderType{}

	for floor := 0; floor < number_of_floors; floor++ {
		for btn := elevio.ButtonType(0); btn < 3; btn++ {
			elevio.SetButtonLamp(btn, floor, false)
		}
	}

	time.Sleep(5 * time.Second)
	elevio.SetStopLamp(false)
}
