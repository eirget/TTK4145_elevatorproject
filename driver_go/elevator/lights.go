package elevator

import "time"

func (e *Elevator) LightUpdater() {
	ticker := time.NewTicker(100 * time.Millisecond) // Check every second
	defer ticker.Stop()
	for range ticker.C {
		e.SetLights()
	}
}
