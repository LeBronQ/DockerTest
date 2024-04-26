package main

import (
    "github.com/nats-io/nats.go"
    "log"
    "time"
    "encoding/json"
    "sync"
    "fmt"
    "math/rand"
)

type Message struct {
    DockerName      string
	InterfaceName   string
	PLR             float64
}

func main() {
    dockerId := []string{"7802311333a3","0525c620a41d","087cf2b90a89","a9004af52853","2a39afad0b9c"}
    var wg sync.WaitGroup
    nc, err := nats.Connect("nats://localhost:4222")
    if err != nil {
        log.Fatal(err)
    }
    defer nc.Close()

    for i := 0; i < 400; i++ {
        wg.Add(1)
        rand.Seed(time.Now().UnixNano())
		randomDockerNumber := rand.Intn(5)
        randomVethNumber := rand.Intn(4)
        DName := dockerId[randomDockerNumber]
        VethName := fmt.Sprintf("veth%d", randomVethNumber)
        msg := Message{
            DockerName: DName,
            InterfaceName: VethName,
            PLR:    0.001,
        }
        reqData, err := json.Marshal(msg)
	    if err != nil {
		    log.Fatal(err)
	    }
        _, err = nc.Request("foo", reqData, 10*time.Second)
        if err != nil {
            log.Fatal(err)
        }
        wg.Done()
    }
    wg.Wait()
}