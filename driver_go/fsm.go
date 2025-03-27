package main

import (
	"Driver_go/config"
	"Driver_go/elevator"
	"Driver_go/elevio"
	"fmt"
	"time"
)

func fsm(
	elev *elevator.Elevator,
	elevStateTx chan elevator.Elevator,
	reqCh chan elevio.ButtonEvent,
	newFloorCh chan int,
	obstrCh chan bool,
	stopCh chan bool,
	newOrderCh chan struct{}) {

	doorTimer := time.NewTimer(0)
	<-doorTimer.C

	for {
		select {
		case newReq := <-reqCh:
			fsmHandleNewRequest(newReq, elev, elevStateTx, newOrderCh)

		case <-time.After(100 * time.Millisecond):
			fsmHandleIdleState(elev, elevStateTx, doorTimer)

		case newFloor := <-newFloorCh:
			fsmHandleNewFloor(newFloor, elev, elevStateTx, doorTimer)

		case <-doorTimer.C:
			fsmHandleDoorTimeout(elev, doorTimer, elevStateTx)

		case isObstructed := <-obstrCh:
			fsmHandleObstruction(isObstructed, elev)

		case <-stopCh:
			fsmHandleEmergencyStop(elev)
		}
	}
}

func fsmHandleNewRequest(
	newReq elevio.ButtonEvent,
	elev *elevator.Elevator,
	elevStateTx chan elevator.Elevator,
	newOrderCh chan struct{}) {

	elev.Orders[newReq.Floor][newReq.Button].State = true
	elev.Orders[newReq.Floor][newReq.Button].Timestamp = time.Now()

	elevStateTx <- *elev

	if newReq.Button == elevio.BTCab {
		elevio.SetButtonLamp(newReq.Button, newReq.Floor, true)
	}

	select {
	case newOrderCh <- struct{}{}:
	default:
	}
}

func fsmHandleIdleState(
	elev *elevator.Elevator,
	elevStateTx chan elevator.Elevator,
	doorTimer *time.Timer) {

	if elev.Behavior == elevator.EBIdle {
		elev.HandleIdleState(doorTimer)
		if elev.Behavior == elevator.EBDoorOpen {
			//doorTimer.Reset(3 * time.Second)
		}
	}
	elevStateTx <- *elev
}

func fsmHandleNewFloor(
	newFloor int,
	elev *elevator.Elevator,
	elevStateTx chan elevator.Elevator,
	doorTimer *time.Timer) {

	elev.FloorNr = newFloor
	elevio.SetFloorIndicator(elev.FloorNr)
	elev.LastActive = time.Now()

	if elev.ShouldStop() {
		elev.StopAtFloor()
		doorTimer.Reset(3 * time.Second)
	}
	elevStateTx <- *elev
}

func fsmHandleDoorTimeout(
	elev *elevator.Elevator,
	doorTimer *time.Timer,
	elevStateTx chan elevator.Elevator) {

	if elev.ShouldReopenForOppositeHallCall() {
		fmt.Println("Should reopen for opposite hall call. Changing direction")

		if elev.LastDirection == elevio.MDUp {
			elev.LastDirection = elevio.MDDown
		} else if elev.LastDirection == elevio.MDDown {
			elev.LastDirection = elevio.MDUp
		}

		elev.Direction = elev.LastDirection
		elev.StopAtFloor()
		doorTimer.Reset(3 * time.Second)
		return
	}

	if elev.Obstruction {
		fmt.Println("Waiting for obstruction to clear...")
		doorTimer.Reset(500 * time.Millisecond)
		return
	}

	elev.CloseDoorAndResume(doorTimer)
	elevStateTx <- *elev
}

func fsmHandleObstruction(isObstructed bool, elev *elevator.Elevator) {

	elev.Obstruction = isObstructed

	if isObstructed {
		fmt.Println("Obstruction detected!")
	}
}

func fsmHandleEmergencyStop(elev *elevator.Elevator) {

	elevio.SetStopLamp(true)
	elevio.SetMotorDirection(elevio.MDStop)
	elev.Behavior = elevator.EBIdle
	elev.Direction = elevio.MDStop
	elev.Orders = [4][3]elevator.OrderType{}

	for floor := 0; floor < config.NumFloors; floor++ {
		for btn := elevio.ButtonType(0); btn < 3; btn++ {
			elevio.SetButtonLamp(btn, floor, false)
		}
	}

	time.Sleep(5 * time.Second)
	elevio.SetStopLamp(false)
}
