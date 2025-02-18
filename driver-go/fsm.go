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
	for {
		select {
		case a := <-reqchan:
			fmt.Printf("%+v\n", a)
			elevator.dealWithNewReq(a.Floor)
			elevio.SetButtonLamp(a.Button, a.Floor, true)

		case a := <-new_floor_chan:
			fmt.Printf("%+v\n", a)
			elevator.Floor_nr = a
			elevio.SetFloorIndicator(a)

			// heller lage en funksjon inni en annen modul som gjør alt dette sikkert

			// If the elevator reaches its first destination in queue
			if len(elevator.Queue) > 0 && elevator.Queue[0] == a {
				elevio.SetMotorDirection(elevio.MD_Stop) // Stop elevator
				fmt.Println("Stopping for 5 seconds...")
				elevio.SetDoorOpenLamp(true)
				time.Sleep(5 * time.Second) // Wait for passengers
				if elevator.Obstruction {
					go elevator.checkForObstruction()

					<-elevator.Resumed
				}
				elevio.SetDoorOpenLamp(false)
				// Remove first floor from queue
				elevator.Queue = elevator.Queue[1:]

				// Turn off floor button light
				elevio.SetButtonLamp(elevio.BT_HallUp, a, false)
				elevio.SetButtonLamp(elevio.BT_HallDown, a, false)
				elevio.SetButtonLamp(elevio.BT_Cab, a, false)

				fmt.Println("Resuming movement...")

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
			elevator.Queue = []int{}
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

/*
func fsm(e *Elevator) { //må ta inn mange channels for ulike "signaler"
	//mutex rundt alt for å beskytte states igjen kanskje
	for {
		select{
			case request := <- ...: //chan for incoming request
				fsm_for_new_request()

			case floor_sensor_update := <- ...:
				fsm_for_floor_sensor_update()

			case obstruction := <-

			case light_update := <-

			case timer := <- //ulike cases for ulike timere om vi ender opp med flere
				//om dør har vært åpen lenge nok --> kjør videre (f.eks)
		}
	}
}
*/
