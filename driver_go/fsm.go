package main

import (
	"Driver_go/elevator"
	"Driver_go/elevio"
	"fmt"
	"time"
)

func fsm(e *elevator.Elevator,
	elevStateTx chan elevator.Elevator,
	req_chan chan elevio.ButtonEvent,
	new_floor_chan chan int,
	obstr_chan chan bool,
	stop_chan chan bool,
	number_of_floors int,
	newOrderCh chan struct{}) {

	doorTimer := time.NewTimer(0)
	<-doorTimer.C

	for {
		select {
		case newReq := <-req_chan:
			fsmHandleNewRequest(newReq, e, elevStateTx, newOrderCh)

		case <-time.After(100 * time.Millisecond):
			fsmHandleIdleState(e, elevStateTx, doorTimer)

		case newFloor := <-new_floor_chan:
			fsmHandleNewFloor(newFloor, e, elevStateTx, doorTimer)

		case <-doorTimer.C:
			fsmHandleDoorTimeout(e, doorTimer, elevStateTx)

		case isObstructed := <-obstr_chan:
			fmt.Println("Obstruction happened")
			fsmHandleObstruction(isObstructed, e)

		case isStopped := <-stop_chan:
			fsmHandleEmergencyStop(isStopped, e, number_of_floors)
		}
	}
}

func fsmHandleNewRequest(newReq elevio.ButtonEvent, e *elevator.Elevator, elevStateTx chan elevator.Elevator, newOrderCh chan struct{}) {
	fmt.Printf("%+v\n", newReq)
	e.Orders[newReq.Floor][newReq.Button].State = true
	e.Orders[newReq.Floor][newReq.Button].Timestamp = time.Now()

	elevStateTx <- *e

	if newReq.Button == elevio.BT_Cab {
		elevio.SetButtonLamp(newReq.Button, newReq.Floor, true)
	}

	select {
	case newOrderCh <- struct{}{}:
	default:
	}
}

func fsmHandleIdleState(e *elevator.Elevator, elevStateTx chan elevator.Elevator, doorTimer *time.Timer) {
	if e.Behavior == elevator.EB_Idle {
		e.HandleIdleState()
		if e.Behavior == elevator.EB_DoorOpen {
			doorTimer.Reset(3 * time.Second)
		}
	}
	elevStateTx <- *e
}

func fsmHandleNewFloor(a int, e *elevator.Elevator, elevStateTx chan elevator.Elevator, doorTimer *time.Timer) {
	e.FloorNr = a
	elevio.SetFloorIndicator(a)
	e.LastActive = time.Now()

	if e.ShouldStop() {
		e.StopAtFloor()
		doorTimer.Reset(3 * time.Second)
	}
	elevStateTx <- *e
}

func fsmHandleDoorTimeout(e *elevator.Elevator, doorTimer *time.Timer, elevStateTx chan elevator.Elevator) {
	if e.ShouldReopenForOppositeHallCall() {
		fmt.Println("Should reopen for opposite hall call")

		if e.LastDirection == elevio.MD_Up {
			e.LastDirection = elevio.MD_Down
		} else if e.LastDirection == elevio.MD_Down {
			e.LastDirection = elevio.MD_Up
		}

		e.Direction = e.LastDirection
		e.StopAtFloor()
		doorTimer.Reset(3 * time.Second)
		return
	}
	if e.Obstruction {
		fmt.Println("Waiting for obstruction to clear...")
		doorTimer.Reset(500 * time.Millisecond)
		return
	}
	e.CloseDoorAndResume()
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
