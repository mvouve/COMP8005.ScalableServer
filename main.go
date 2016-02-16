/*------------------------------------------------------------------------------
-- DATE:	       February, 2016
--
-- Source File:	 child_proc.go
--
-- REVISIONS: 	(Date and Description)
--
-- DESIGNER:	   Marc Vouve
--
-- PROGRAMMER:	 Marc Vouve
--
--
-- INTERFACE:
--	func newConnection(srvInfo serverInfo)
--  func worker(srvInfo serverInfo)
--  func connectionInstance(conn net.Conn) connectionInfo
--  func handleData(conn net.Conn, connInfo *connectionInfo) error
--  func observerLoop(srvInfo serverInfo, osSignals chan os.Signal)
--  func newServerInfo() serverInfo
--
-- NOTES: This file is for functions that are part of child go routines which
--        handle data for the EPoll version of the scalable server.
------------------------------------------------------------------------------*/
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
	HostName           string // the remote host name
	AmmountOfData      int    // the ammount of data transfered to/from the host
	NumberOfRequests   int    // the total requests sent to the server from this client
	ConnectionsAtClose int    // the total number of connections being sustained when the connection was closed.
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
const startingClients = 15
const freeServerMinimum = 10

func main() {
	if len(os.Args) < 2 { // validate args
		log.Fatalln("Missing args:", os.Args[0], " [PORT]")
	}

	srvInfo := newServerInfo()

	// create servers
	for i := 0; i < startingClients; i++ {
		*srvInfo.availableServers++
		go worker(srvInfo)
	}

	// when the server is killed it should print statistics need to catch the signal
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, os.Kill)

	observerLoop(srvInfo, osSignals)
}

/*-----------------------------------------------------------------------------
-- FUNCTION:    newConnection
--
-- DATE:        February 6, 2016
--
-- REVISIONS:
--
-- DESIGNER:		Marc Vouve
--
-- PROGRAMMER:	Marc Vouve
--
-- INTERFACE:   newConnection(srvInfo serverInfo)
--	 srvInfo:		information about the overall server
--
-- RETURNS:     void
--
-- NOTES:			Called when a new client connects to the server.
------------------------------------------------------------------------------*/
func newConnection(srvInfo serverInfo) {
	*srvInfo.totalConnections++
	if *srvInfo.availableServers < freeServerMinimum {
		go worker(srvInfo)
	} else {
		*srvInfo.availableServers--
	}
}

/*-----------------------------------------------------------------------------
-- FUNCTION:    worker
--
-- DATE:        February 6, 2016
--
-- REVISIONS:
--
-- DESIGNER:		Marc Vouve
--
-- PROGRAMMER:	Marc Vouve
--
-- INTERFACE:   serverInstance(srvInfo serverInfo)
--	 srvInfo:		information about the overall server
--
-- RETURNS:     void
--
-- NOTES:			This function is a worker thread, it accepts connections from
--						outside and handles data from them.
------------------------------------------------------------------------------*/
func worker(srvInfo serverInfo) {

	for {
		conn, err := srvInfo.listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		srvInfo.serverConnection <- newConnectionConst
		srvInfo.connectInfo <- connectionInstance(conn)
		conn.Close()
	}

}

/*-----------------------------------------------------------------------------
-- FUNCTION:    connectionInstance
--
-- DATE:        February 6, 2016
--
-- REVISIONS:
--
-- DESIGNER:		Marc Vouve
--
-- PROGRAMMER:	Marc Vouve
--
-- INTERFACE:   func connectionInstance(conn net.Conn) connectionInfo
--      conn:		a connection to a client.
--
-- RETURNS:   connectionInfo information about the connection when it's complete
--
-- NOTES:			This is the main data handling function
------------------------------------------------------------------------------*/
func connectionInstance(conn net.Conn) connectionInfo {
	connInfo := connectionInfo{HostName: conn.RemoteAddr().String()}
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

	return connInfo
}

/*-----------------------------------------------------------------------------
-- FUNCTION:    connectionInstance
--
-- DATE:        February 6, 2016
--
-- REVISIONS:
--
-- DESIGNER:		Marc Vouve
--
-- PROGRAMMER:	Marc Vouve
--
-- INTERFACE:   func connectionInstance(conn net.Conn) connectionInfo
--      conn:		a connection to a client.
--
-- RETURNS:   connectionInfo information about the connection when it's complete
--
-- NOTES:			This is the main data handling function
------------------------------------------------------------------------------*/
func handleData(conn net.Conn, connInfo *connectionInfo) error {
	reader := bufio.NewReader(conn)
	data, err := reader.ReadBytes('\n')
	if err != nil {
		return err
	}
	connInfo.AmmountOfData += len(data)
	connInfo.NumberOfRequests++
	conn.Write(data)

	return nil
}

/*-----------------------------------------------------------------------------
-- FUNCTION:    observerLoop
--
-- DATE:        February 6, 2016
--
-- REVISIONS:
--
-- DESIGNER:		Marc Vouve
--
-- PROGRAMMER:	Marc Vouve
--
-- INTERFACE:   func observerLoop(srvInfo serverInfo, osSignals chan os.Signal)
--   srvInfo:		Information about the server.
-- osSignals:		reads signals from the OS and stops the program.
--
-- RETURNS:   connectionInfo information about the connection when it's complete
--
-- NOTES:			This is the main data handling function
------------------------------------------------------------------------------*/
func observerLoop(srvInfo serverInfo, osSignals chan os.Signal) {
	currentConnections := 0
	connectionsMade := list.New()

	for {
		select {
		case <-srvInfo.serverConnection:
			currentConnections++
			newConnection(srvInfo)
		case serverHost := <-srvInfo.connectInfo:
			serverHost.ConnectionsAtClose = currentConnections
			connectionsMade.PushBack(serverHost)
			currentConnections--
		case <-osSignals:
			generateReport(time.Now().String(), connectionsMade)
			fmt.Println("Total connections made:", connectionsMade.Len())
			os.Exit(1)
		}
	}
}

/*-----------------------------------------------------------------------------
-- FUNCTION:    newServerInfo
--
-- DATE:        February 6, 2016
--
-- REVISIONS:
--
-- DESIGNER:		Marc Vouve
--
-- PROGRAMMER:	Marc Vouve
--
-- INTERFACE:   func newServerInfo() serverInfo
--
-- RETURNS:   serverInfo information about the server
--
-- NOTES:			This function builds the basic info about the server.
------------------------------------------------------------------------------*/
func newServerInfo() serverInfo {
	var err error
	srvInfo := serverInfo{totalConnections: new(int), availableServers: new(int),
		serverConnection: make(chan int, 10), connectInfo: make(chan connectionInfo)}
	if srvInfo.listener, err = net.Listen("tcp", os.Args[1]); err != nil {
		log.Fatalln(err)
	}

	return srvInfo
}
