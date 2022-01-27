package client

import (
	"coap-server/utils"
	"fmt"
	"log"
	"net"
	"time"
)

type State struct {
	Addr                *net.UDPAddr
	MaxRetries          int
	TimesBetweenRetries time.Duration
	conn                *net.UDPConn
	pktNumber           uint16
	pktChan             <-chan []byte
}

func NewClientState(conn *net.UDPConn, addr *net.UDPAddr, pktChan <-chan []byte) State {
	return State{
		Addr:                addr,
		MaxRetries:          4,
		TimesBetweenRetries: 4 * time.Second,
		conn:                conn,
		pktNumber:           1,
		pktChan:             pktChan,
	}
}

type PktEncoder interface {
	Encode(buffer []byte) []byte
}

func (client *State) Send(pktEncoder PktEncoder) error {
	// The packet to send is always prefixed by a packet number to track if the packet is correctly received
	pktToSend := client.newPkt()
	pktToSend = pktEncoder.Encode(pktToSend)
	log.Println("Size of packet: ", len(pktToSend))
	var err error
	for i := 0; i < client.MaxRetries; i++ {
		_, err = client.conn.WriteTo(pktToSend, client.Addr)
		if err != nil {
			log.Println("Error while trying to send the packet to the client: ", err)
			log.Println("Retrying in ", client.TimesBetweenRetries)
			time.Sleep(client.TimesBetweenRetries)
		} else {
			break
		}
	}
	if err != nil {
		log.Panic("Error maximum numbre of retries exhausted panicking with error : ", err)
	}
	for {
		select {
		case pkt := <-client.pktChan:
			var pktNumber = (uint16(pkt[1]) << 8) + uint16(pkt[0])
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
		case <-time.After(4 * time.Second):
			log.Printf("Timeout on addr %s, resending pkt\n", client.Addr.IP)
			_, err := client.conn.WriteTo(pktToSend, client.Addr)
			if err != nil {
				return err
			}
		}
	}
}
func (client *State) newPkt() []byte {
	pkt := make([]byte, 0, 128)
	pkt = client.encodePktNumber(pkt)
	return pkt
}

func (client *State) encodePktNumber(pkt []byte) []byte {
	pkt = utils.AppendLittleEndianUint16(pkt, client.pktNumber)
	client.pktNumber++
	return pkt
}
