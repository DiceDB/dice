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

func stormSet(wg *sync.WaitGroup, conn net.Conn) {
	defer wg.Done()
	for {
		time.Sleep(500 * time.Millisecond)
		k, v := getRandomKeyValue()
		cmd := fmt.Sprintf("SET %s %d", k, v)
		fmt.Println(cmd)
		buf, err := core.Encode(strings.Split(cmd, " "), false)
		if err != nil {
			log.Println(err)
			return
		}
		_, err = conn.Write(buf)
		if err != nil {
			log.Println(err)
			return
		}
		resp := make([]byte, 512)
		_, err = conn.Read(resp)
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Println(err)
			return
		}
	}
}

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		conn, err := net.Dial("tcp", "localhost:7379")
		if err != nil {
			log.Fatal(err)
		}
		wg.Add(1)
		go stormSet(&wg, conn)
	}
	wg.Wait()
}
