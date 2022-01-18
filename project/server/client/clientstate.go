package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

type ClientState struct {
	Addr      *net.UDPAddr
	conn      *net.UDPConn
	pktNumber uint16
	pktChan   <-chan []byte
}

func NewClientState(conn *net.UDPConn, addr *net.UDPAddr, pktChan <-chan []byte) ClientState {
	return ClientState{
		Addr:      addr,
		conn:      conn,
		pktNumber: 1,
		pktChan:   pktChan,
	}
}

func (client *ClientState) EncodePktNumber(buffer *bytes.Buffer) error {
	pktNumberBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(pktNumberBytes, client.pktNumber)
	_, err := buffer.Write(pktNumberBytes)
	client.pktNumber++
	return err
}

func (client *ClientState) Send(pktToSend []byte) error {
	_, err := client.conn.WriteTo(pktToSend, client.Addr)
	if err != nil {
		return err
	}
	for {
		select {
		case pkt := <-client.pktChan:
			var pktNumber uint16 = (uint16(pkt[0]) << 8) + uint16(pkt[1])
			expectedPktNumber := client.pktNumber - 1
			log.Printf("Expecting ACK: %d, got %d\n", expectedPktNumber, pktNumber)
			if pktNumber == expectedPktNumber {
				log.Println("ACK correctly received")
				fmt.Println()
				return nil // Client correctly received the pktToSend
			} else if pktNumber < expectedPktNumber {
				log.Println("Already received this ack -> doing nothing")
			} else {
				log.Println("Message received that was not the expected message, resending pktToSend")
				log.Println("Message received: ", string(pkt))
				_, err := client.conn.WriteTo(pktToSend, client.Addr)
				if err != nil {
					return err
				}
			}
		case <-time.After(6 * time.Second):
			log.Printf("Timeout on addr %s, resending pkt\n", client.Addr.IP)
			_, err := client.conn.WriteTo(pktToSend, client.Addr)
			if err != nil {
				return err
			}
		}
	}
}

func (client *ClientState) PeriodicSend() {
	buffer := make([]byte, 2)
	binary.BigEndian.PutUint16(buffer, client.pktNumber)
	client.Send(buffer)
}
