package main

import (
	"context"
    "github.com/nats-io/nats.go"
    "log"
	"sync"
    "fmt"
	"io"
	"os"
    "encoding/json"
    "strconv"

    "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/panjf2000/ants/v2"
)

type Message struct {
	DockerName      string
	InterfaceName   string
	PLR             float64
}

type Task struct {
	msg     Message
	wg      *sync.WaitGroup  
}


func (t *Task) configureInterfaces() {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}
	containerName := t.msg.DockerName
	s := strconv.FormatFloat(t.msg.PLR*100, 'f', -1, 64)
	execConfig := types.ExecConfig{
		Cmd:          []string{"sudo", "tc", "qdisc", "change", "dev", t.msg.InterfaceName, "root", "netem", "loss", s}, 
		AttachStdout: true,                
		AttachStderr: true,                 
		Tty:          false,                
	}

	// 创建容器执行命令
	execID, err := cli.ContainerExecCreate(context.Background(), containerName, execConfig)
	if err != nil {
		fmt.Println("Failed to create exec command:", err)
		return
	}
	// 执行容器命令并获取输出
	res, err := cli.ContainerExecAttach(context.Background(), execID.ID, types.ExecStartCheck{})
	if err != nil {
		fmt.Println("Failed to attach to exec command:", err)
		return
	}
	defer res.Close()

	// 将容器命令的输出复制到当前进程的stdout
	_, err = io.Copy(os.Stdout, res.Reader)
	if err != nil {
		fmt.Println("Failed to copy output:", err)
		return
	}

    defer cli.Close()
	t.wg.Done()
}

func main() {
    nc, err := nats.Connect("nats://localhost:4222")
    if err != nil {
        log.Fatal(err)
    }
    defer nc.Close()
    defer ants.Release()
	var wg sync.WaitGroup
	taskFunc := func(data interface{}) {
		task := data.(*Task)
		task.configureInterfaces()
	}
    p, _ := ants.NewPoolWithFunc(4, taskFunc)
    defer p.Release()
    nc.Subscribe("foo", func(msg *nats.Msg) {
        log.Println("Request receive:", string(msg.Data))
        var resp Message
        err = json.Unmarshal(msg.Data, &resp)
	    if err != nil {
		    log.Fatal(err)
	    }
        wg.Add(1)
		task := &Task{
			msg:    resp,
			wg:      &wg,
		}
		p.Invoke(task)
        wg.Wait()
        msg.Respond([]byte("Finished"))
    })

    select {}
}