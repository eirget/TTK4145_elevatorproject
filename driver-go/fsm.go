package main

import "Driver-go/elevio"

func simple_fsm(req chan elevio.ButtonEvent) {
	for {
		select {
		case a := <- drv_buttons:
			fmt.Printf("%+v\n", a)
			elevio.SetButtonLamp(a.Button, a.Floor, true)
			simple_req_fsm(a)
			
		
		/*
		case a := <- drv_floors:
			fmt.Printf("%+v\n", a)
			if a == numFloors-1 {
				d = elevio.MD_Down
			} else if a == 0 {
				d = elevio.MD_Up
			}
			//elevio.SetMotorDirection(d)


		case a := <- drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				//elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				//elevio.SetMotorDirection(d)
			}

		case a := <- drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			} */
		}
	}
}

func simple_req_fsm() {
	//deal with request
}

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