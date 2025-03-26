package elevator

import (
	"fmt"
	"time"
)

// where should this go???
func MonitorActivity(e *Elevator, runHra chan<- bool) {
	ticker := time.NewTicker(1 * time.Second) // Check every second
	defer ticker.Stop()
	fmt.Printf("Last active: %v \n: ", e.LastActive)
	// need to double check with some sort of "heartbeat" if it actually doesnt work, update lastActive if nothing is wrong
	for range ticker.C {
		if time.Since(e.LastActive) > 5*time.Second { // Elevator inactive for 5+ seconds
			fmt.Println("I have been inactive")
			if e.HasPendingOrders() {
				fmt.Println("And I have pending orders, calling hall request assigner")
				runHra <- true // Trigger hall request reassignment
				//return
			}
		}
	}
}
