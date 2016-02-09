package main

import (
	"bufio"
	"container/list"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

type connectionInfo struct {
	timeStamp          time.Time // the time the connection ended
	hostName           string    // the remote host name
	ammountOfData      int       // the ammount of data transfered to/from the host
	numberOfRequests   int       // the total requests sent to the server from this client
	connectionsAtClose int       // the total number of connections being sustained when the connection was closed.
}

type serverInfo struct {
	totalConnections *int
	availableServers *int
	serverConnection chan int
	connectInfo      chan connectionInfo
	listener         net.Listener
}

const newConnectionConst = 1
const finishedConnectionConst = -1

/* Author Marc Vouve
 *
 * Designer Marc Vouve
 *
 * Date: February 6 2016
 *
 * Notes: This is a helper function for when a new connection is detected by the
 *        observer loop
 *
 */
func newConnection(srvInfo serverInfo) {
	*srvInfo.totalConnections++
	if *srvInfo.availableServers < 2 {
		go serverInstance(srvInfo)
	} else {
		*srvInfo.availableServers--
	}
}

/* Author: Marc Vouve
 *
 * Designer: Marc Vouve
 *
 * Date: February 6 2016
 *
 * Notes: This function is an "instance" of a server which allows connections in
 *        and echos strings back. After a connection has been closed it will wait
 *        for annother connection
 */
func serverInstance(srvInfo serverInfo) {

	for {
		conn, err := srvInfo.listener.Accept()
		defer conn.Close()
		if err != nil {
			log.Print(err)
			continue
		}
		srvInfo.serverConnection <- newConnectionConst
		srvInfo.connectInfo <- connectionInstance(conn)
	}

}

/* Author: Marc Vouve
 *
 * Designer: Marc Vouve
 *
 * Date: February 6 2016
 *
 * Returns: connectionInfo information about the connection once the client has
 *          terminated the client.
 *
 * Notes: This function handles actual connections made to the server and tracks
 *        data about the ammount of data and the number of times data is set to
 *        the server.
 */
func connectionInstance(conn net.Conn) connectionInfo {
	connInfo := connectionInfo{hostName: conn.RemoteAddr().String()}

	for {
		err := handleData(conn, &connInfo)
		if err == nil {
			continue
		} else if err == io.EOF {
			break
		}
		log.Println(err)
		break
	}
	connInfo.timeStamp = time.Now()

	return connInfo
}

func handleData(conn net.Conn, connInfo *connectionInfo) error {
	reader := bufio.NewReader(conn)
	data, err := reader.ReadBytes('\n')
	if err != nil {
		return err
	}
	connInfo.ammountOfData += len(data)
	connInfo.numberOfRequests++
	conn.Write(data)

	return nil
}

/* Author: Marc Vouve
 *
 * Designer: Marc Vouve
 *
 * Date: February 7 2016
 *
 * Returns: connectionInfo information about the connection once the client has
 *          terminated the client.
 *
 * Notes: This was factored out of the main function.
 */
func observerLoop(srvInfo serverInfo, osSignals chan os.Signal) {
	currentConnections := 0
	connectionsMade := list.New()

	for {
		select {
		case <-srvInfo.serverConnection:
			currentConnections++
			newConnection(srvInfo)
		case serverHost := <-srvInfo.connectInfo:
			serverHost.connectionsAtClose = currentConnections
			connectionsMade.PushBack(serverHost)
			currentConnections--
		case <-osSignals:
			generateReport(time.Now().String(), connectionsMade)
			fmt.Println("Total connections made:", connectionsMade.Len())
			os.Exit(1)
		}
	}
}

func newServerInfo() serverInfo {
	srvInfo := serverInfo{totalConnections: new(int), availableServers: new(int),
		serverConnection: make(chan int, 10), connectInfo: make(chan connectionInfo)}
	srvInfo.listener, _ = net.Listen("tcp", os.Args[1])

	return srvInfo
}

func main() {
	if len(os.Args) < 2 { // validate args
		fmt.Println("Missing args:", os.Args[0], " [PORT]")

		os.Exit(0)
	}

	srvInfo := newServerInfo()

	// create servers
	for i := 0; i < 10; i++ {
		*srvInfo.availableServers++
		go serverInstance(srvInfo)
	}

	// when the server is killed it should print statistics need to catch the signal
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, os.Kill)

	observerLoop(srvInfo, osSignals)
}
