package server

func RunThreadedServer(serverFD int, wg *sync.WaitGroup) {
	defer wg.Done()
	defer syscall.Close(serverFD)

	log.Info("starting an threaded TCP server on", config.Host, config.Port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ln, err := net.Listen("tcp", config.Hos + ":" + config.Port)
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	fmt.Println("server listening on port ", config.Port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("error accepting connection, moving on ...", err)
			continue
		}
		
		cmds, hasABORT, err := readCommands(comm)

		for _, cmd := range cmds{
			ipool.Get().reqch <- &Request{conn: conn, cmd, core.getKeyForOperation(cmd)}
		}
	}
}