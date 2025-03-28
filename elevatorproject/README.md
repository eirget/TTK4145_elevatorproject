# Elevator Project

A software system designed to control n elevators operating concurrently across m floors.


## About
This project implements a distributed control system for multiple elevators operating across several floors. The system is designed to be fault-tolerant, responsive, and capable of recovering from common failure scenarios such as disconnections or crashes. 


## DesignOverview 

### Peer-to-Peer Network Architecture 
The system follows a pure peer-to-peer design. Each elevator runs independently and communicates directly with the other elevators using UDP broadcast. There is no central coordinator or server dependency. 

### Multi-Channel Communication 
To organize responsibilities, different types of data are transmitted over distinct UDP ports. These include elevator states, peer connectivity updates, and order coordination signals. 

### Failure Resilience 
The system handles elevator restarts, power loss, and network disconnections. Orders are not lost during temporary failures, and elevators resume from the last known state when reconnected. 

### Dynamic Hall Call Distribution 
A separate Hall Request Assigner (HRA) module is used to dynamically assign hall calls based on elevator availability and state. This module is triggered when an elevator connects/disconnects or becomes idle. 

### Cab Call Persistence 
Cab calls are stored locally by the elevator they belong to. This means they are serviced even if the elevator is disconnected from the network, allowing passengers to still reach their destination. 

### Modular Hardware Interface 
The elevator hardware is abstracted through a driver layer, responsible for polling sensors and setting button lamps, floor indicators, and door lights. 


## Usage
Each elevator is started with a specified port number, which determines its ID. The system supports multiple elevators on the same local network and is configurable for different numbers of floors or elevators. No manual intervention is needed for recovering from disconnections or restarts. 
To run this program write the following command in terminal: "make PORT=<portnumber>". 


## ImplementationHighlights

### Order Synchronization via Timestamps 
Each order has a corresponding timestamp. This ensures that elevators only apply newer updates from peers, preventing outdated data from overwriting a valid local state. 

### Activity Monitoring and Reassignment 
Elevators monitor their own activity status and can autonomously trigger a reassignment of hall requests if they detect a lack of progress or failure.  

### State Recovery and Merging  
When elevators rejoin the network their last known state is restored. Then the current sets of orders are merged to preserve valid active calls. 


## Authors 
This project is a collaboration between the members of group 64 of the NTNU course TTK4145 - Real-Time Programming:
- **[Eirin Berget](https://github.com/eirget)** 
- **[Adele Ferger](https://github.com/aferger)** 
- **[Stine Kvitastein](https://github.com/Stikvi)** 


## Acknowledgments 
We would like to acknowledge and give thanks to the creators of the following resources that were used in our project: 

- **[TTK4145 Organization](https://github.com/TTK4145/Project-resources/tree/master/cost_fns/hall_request_assigner)** - The files in our hall_request_assigner folder are from this resource. 
- **[TTK4145 Organization](https://github.com/TTK4145/Network-go/tree/master/network)** - The files bcast.go, bcast_conn.go, localip.go and peers.go in our network folder are from this resource. 
- **[TTK4145 Organization](https://github.com/TTK4145/driver-go/blob/master/elevio/elevator_io.go)** - The file elevator_io.go in our elevio folder are from this project resource (except for the function ElevioInit())



