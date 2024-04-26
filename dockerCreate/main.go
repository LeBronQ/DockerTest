package main

import (
	"context"
	"fmt"
	"log"
	"io"
	"os"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func main() {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}
	hostConfig := &container.HostConfig{
		Privileged: true, // 将容器设置为特权模式
	}

	for i := 0; i < 5; i++ {
		// 创建并启动容器
		containerName := "docker" + strconv.Itoa(i)
		resp, err := cli.ContainerCreate(context.Background(), &container.Config{
			Image: "my-image:latest",
			Tty:   true,
			Cmd:   []string{"tail","-f","/dev/null"},
		}, hostConfig, nil, nil, containerName)
		if err != nil {
			log.Fatal(err)
		}

		// 启动容器
		err = cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{})
		if err != nil {
			log.Fatal(err)
		}
		// 容器执行命令的配置

		commands := []string{"sudo ip link add veth1 type veth peer name veth1-peer", 
		"sudo ip link set veth1 up", 
		"sudo ip link set veth1-peer up",
		"sudo ip link add veth2 type veth peer name veth2-peer", 
		"sudo ip link set veth2 up", 
		"sudo ip link set veth2-peer up",
		"sudo ip link add veth3 type veth peer name veth3-peer", 
		"sudo ip link set veth3 up", 
		"sudo ip link set veth3-peer up",
		"sudo ip link add veth0 type veth peer name veth0-peer", 
		"sudo ip link set veth0 up", 
		"sudo ip link set veth0-peer up",
		"sudo tc qdisc add dev veth1 root tbf rate 1mbit burst 1600 latency 50ms",
		"sudo tc qdisc add dev veth2 root tbf rate 1mbit burst 1600 latency 50ms",
		"sudo tc qdisc add dev veth3 root tbf rate 1mbit burst 1600 latency 50ms",
		"sudo tc qdisc add dev veth0 root tbf rate 1mbit burst 1600 latency 50ms",
		} 
		// 将多个命令连接成一个字符串
		command := ""
		for _, cmd := range commands {
			command += cmd + " && "
		}
		// 删除末尾的 " && " 字符串
		command = command[:len(command)-4]
		execConfig := types.ExecConfig{
			Cmd:          []string{"sh", "-c", command}, 
			AttachStdout: true,                 // 指定是否将标准输出附加到当前进程的stdout
			AttachStderr: true,                 // 指定是否将标准错误附加到当前进程的stderr
			Tty:          false,                // 指定是否为执行分配一个tty
		}

		// 创建容器执行命令
		execID, err := cli.ContainerExecCreate(context.Background(), resp.ID, execConfig)
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

		// 打印容器 ID
		fmt.Printf("Container %d started: %s\n", i+1, resp.ID[:10])
	}
}

