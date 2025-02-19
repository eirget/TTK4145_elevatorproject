package main

import (
	"Driver-go/elevio"
	"fmt"
	"time"
)

func SimpleFsm(elevator *Elevator,
	reqchan chan elevio.ButtonEvent,
	new_floor_chan chan int,
	obstr_chan chan bool,
	stop_chan chan bool,
	number_of_floors int) {

	doorTimer := time.NewTimer(0)
	<-doorTimer.C

	obstructionTimer := time.NewTimer(0)
	<-obstructionTimer.C

	for {
		select {
		case a := <-reqchan:
			fmt.Printf("%+v\n", a)
			elevator.dealWithNewReq(a.Floor, a.Button)
			elevio.SetButtonLamp(a.Button, a.Floor, true)

		case a := <-new_floor_chan:
			fmt.Printf("%+v\n", a)
			elevator.Floor_nr = a
			elevio.SetFloorIndicator(a)

			// heller lage en funksjon inni en annen modul som gjør alt dette sikkert

			// If the elevator reaches its first destination in queue
			if len(elevator.Queue) > 0 && elevator.Queue[0].Floor == a {
				elevio.SetMotorDirection(elevio.MD_Stop) // Stop elevator
				elevator.Direction = elevio.MD_Stop
				fmt.Println("Stopping for 5 seconds...")
				elevio.SetDoorOpenLamp(true)

				doorTimer.Reset(5 * time.Second)
			}


		case <-doorTimer.C:

			if elevator.Obstruction {
				fmt.Println("Waiting for obstruction to clear...")
				obstructionTimer.Reset(500 * time.Millisecond)
			} else {
				elevio.SetDoorOpenLamp(false)
				// Remove first floor from queue
				// Turn off floor button light
				elevio.SetButtonLamp(elevator.Queue[0].Button_direction, elevator.Queue[0].Floor, false)
				elevio.SetButtonLamp(elevio.BT_Cab, elevator.Queue[0].Floor, false)

				fmt.Printf("%+v\n", elevator.Queue)

				elevator.Queue = elevator.Queue[1:]
				

				fmt.Println("Resuming movement...")
				fmt.Println("Resuming with queue:")
				fmt.Printf("%+v\n", elevator.Queue)
			}

		case <-obstructionTimer.C:
			if elevator.Obstruction {
				obstructionTimer.Reset(500 * time.Millisecond)
			} else {
				fmt.Println("Obstruction cleared")
				

				elevio.SetDoorOpenLamp(false)

				elevio.SetButtonLamp(elevator.Queue[0].Button_direction, elevator.Queue[0].Floor, false)
				elevio.SetButtonLamp(elevio.BT_Cab, elevator.Queue[0].Floor, false)

				fmt.Printf("%+v\n", elevator.Queue)

				elevator.Queue = elevator.Queue[1:]
				
				

				fmt.Println("Resuming movement...")
				fmt.Println("Resuming with queue:")
				fmt.Printf("%+v\n", elevator.Queue)
			}
			


		case a := <-obstr_chan:
			fmt.Printf("%+v\n", a)
			elevator.Obstruction = a
			/*
				if a && elevio.GetFloor() != -1 {
					elevio.SetMotorDirection(elevio.MD_Stop) //sørger for at den ikke kjører videre
					elevio.SetDoorOpenLamp(true)        // dersom døren ikke allerede er åpen, gjør ingenting om lyset allerede er på
					// antar at lampen lyser så lenge obstruction er på
				} else { //dersom vi ikke befinner oss på en etasje, trenger vi ikke gjøre noe
					fmt.Println("Obstruction activated, not on floor.") // Kan fjernes når kode er good
				}
			*/
		case a := <-stop_chan:
			fmt.Printf("%+v\n", a)
			elevio.SetStopLamp(true)
			elevator.Queue = []Order{}
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

