package main

import (
	"Driver-go/elevio"
	//"debug/elf" //virker ikke som den trengs
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
			elevio.SetButtonLamp(a.Button, a.Floor, true)
			elevator.AddToQueue(a.Floor)

		case a := <-new_floor_chan:
			fmt.Printf("%+v\n", a)
			elevator.Floor_nr = a

			// If the elevator reaches its first destination in queue
			if len(elevator.Queue) > 0 && elevator.Queue[0] == a {
				elevio.SetFloorIndicator(elevator.Floor_nr) //Kobler floor sensor med floor indicator light
				elevio.SetMotorDirection(elevio.MD_Stop) // Stop elevator
				elevio.SetDoorOpenLamp(true)
				fmt.Println("Stopping for 5 seconds...")
				time.Sleep(5 * time.Second) // Wait for passengers
				elevio.SetDoorOpenLamp(false)
				// Remove first floor from queue
				elevator.Queue = elevator.Queue[1:]
				fmt.Println("Empty queue")

				// Turn off floor button light
				if elevator.Queue[0] > a {
					elevio.SetButtonLamp(elevio.BT_HallUp, a, false)
					elevio.SetButtonLamp(elevio.BT_Cab, a, false)
				} else if elevator.Queue[0] < a {
					elevio.SetButtonLamp(elevio.BT_HallDown, a, false)
					elevio.SetButtonLamp(elevio.BT_Cab, a, false)
				} else {
					fmt.Println("Duplicate in queue, not valid")
				} // elsen er antagelig overflødig
				
				

				fmt.Println("Resuming movement...")
			}

		case a := <-obstr_chan:
			fmt.Printf("%+v\n", a)
			//obstruction kan skje mellom etasjer (settes mellom etasjer), men vil kun åpne dørene på en etasje
			if a {
				if elevator.Floor_nr != -1 { // dersom vi befinner oss på en etasje 
					elevio.SetMotorDirection(elevio.MD_Stop) //sørger for at den ikke kjører videre
					elevio.SetDoorOpenLamp(true) // dersom døren ikke allerede er åpen, gjør ingenting om lyset allerede er på
					// antar at lampen lyser så lenge obstruction er på
				} else { //dersom vi ikke befinner oss på en etasje, trenger vi ikke gjøre noe
					fmt.Println("Obstruction activated, not on floor.") // Kan fjernes når kode er good
				}
			} else { 
				elevio.SetMotorDirection(elevio.MotorDirection(elevator.Direction)) // dersom obstruction er false fortsetter vi som før 
			}

		//Vi tolker at stop skru på stop-lys, tømme køen, lyse i 5 sek (kun for demo) og skru av lys
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
