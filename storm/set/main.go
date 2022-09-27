package main

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dicedb/dice/core"
)

func getRandomKeyValue() (string, int64) {
	token := int64(rand.Uint64() % 5000000)
	return "k" + strconv.FormatInt(token, 10), token
}

func stormSet(wg *sync.WaitGroup) {
	defer wg.Done()
	conn, err := net.Dial("tcp", "localhost:7379")
	if err != nil {
		panic(err)
	}

	for {
		time.Sleep(500 * time.Millisecond)
		k, v := getRandomKeyValue()
		var buf [512]byte
		cmd := fmt.Sprintf("SET %s %d", k, v)
		fmt.Println(cmd)
		_, err = conn.Write(core.Encode(strings.Split(cmd, " "), false))
		if err != nil {
			panic(err)
		}
		_, err = conn.Read(buf[:])
		if err != nil {
			if err == io.EOF {
				return
			}
			panic(err)
		}
	}
	conn.Close()
}

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go stormSet(&wg)
	}
	wg.Wait()
}
