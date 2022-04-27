package udpack

import (
	"errors"
	"fmt"
	"log"
	"net"
	"scheduleupdater-server/addrtranslation"
	"scheduleupdater-server/stats"
	"scheduleupdater-server/utils"
	"sync"
	"time"
)

// UDPAckConn handles UDP connections with ACK of packets. Only one UDP packet
// can be flying at the same time.
type UDPAckConn struct {
	Config                   *UDPAckConnSendConfig
	conn                     *net.UDPConn
	receivedSequencesNumbers SequenceNumbersMap
	sentSequencesNumbers     SequenceNumbersMap
	ackChannels              map[addrtranslation.IPString]chan []byte
	lock                     sync.RWMutex
}

func NewUDPAckServer(conn *net.UDPConn, config *UDPAckConnSendConfig) *UDPAckConn {
	if config == nil {
		config = newDefaultUDPAckConnSendConfig()
	}
	fmt.Println("Timeout value: ", config.Timeout)
	return &UDPAckConn{
		Config:                   config,
		conn:                     conn,
		receivedSequencesNumbers: NewSequenceNumbersMap(),
		sentSequencesNumbers:     NewSequenceNumbersMap(),
		ackChannels:              make(map[addrtranslation.IPString]chan []byte),
		lock:                     sync.RWMutex{},
	}
}

// UDPAckServerHandler is the callback called when receiving a packet. The
// packet can be of three types `PacketTypeData` which corresponds to a data packet,
// `PacketTypeDataNoACK` which is a data packet wich doesn't need an ACK
// and `PacketTypeAck` which corresponds to an ACK packet. The Ack packet must be
// processed by the handler and resent the latest packet if the sequence number
// contained in the Ack is not the sequence number expected.
type UDPAckServerHandler = func(addr *net.UDPAddr, packet []byte)

// Serve listen to all incoming packets, verify the sequence number and dispatch
// the UDP packet to the handler
func (udpAckConn *UDPAckConn) Serve(handler UDPAckServerHandler) error {
	buffer := make([]byte, 2048)
	for {
		rlen, remote, err := udpAckConn.conn.ReadFromUDP(buffer[:])
		if err != nil {
			return err
		}
		remoteAddrString := remote.IP.String()
		utils.Log.InfoPrintln("Message received from ", remoteAddrString)
		packet := make([]byte, rlen)
		copy(packet, buffer)

		err = udpAckConn.handlePacket(remote, packet, handler)
		if err != nil {
			return err
		}
	}
}

func (udpAckConn *UDPAckConn) WriteTo(packet []byte, addr *net.UDPAddr) (error, []byte) {
	addrIP := addrtranslation.AddrToIPString(addr)
	udpAckConn.lock.Lock()
	ackChan, in := udpAckConn.ackChannels[addrIP]
	if !in {
		ackChan = make(chan []byte)
		udpAckConn.ackChannels[addrIP] = ackChan
	}
	udpAckConn.lock.Unlock()
	nextSequenceNumber := udpAckConn.sentSequencesNumbers.expected(addrIP)
	packetWithHeader, err := newDataPacket(nextSequenceNumber, packet)
	if err != nil {
		return err, nil
	}
	config := udpAckConn.Config
	conn := udpAckConn.conn
	for i := 0; i < config.MaxRetries; i++ {
		_, err = conn.WriteTo(packetWithHeader, addr)
		stats.SimulationStats.ProtocolSent.Increment(addrIP)
		stats.SimulationStats.Nsent.Increment(addrIP)
		if err != nil {
			utils.Log.ErrorPrintln("Error while trying to send the packet to the udpack: ", err)
			utils.Log.ErrorPrintln("Retrying in ", config.TimesBetweenRetries)
			time.Sleep(config.TimesBetweenRetries)
		} else {
			break
		}
	}
	if err != nil {
		log.Panic("Error maximum numbre of retries exhausted panicking with error : ", err)
	}
	// Waiting for the ack to be received
	for i := 0; i < config.MaxRetries; i++ {
		utils.Log.WarningPrintln("Looping...")
		select {
		case pkt := <-ackChan:
			stats.SimulationStats.ProtocolReceived.Increment(addrIP)
			header, packetWithoutHeader := RemoveHeaderFromPacket(pkt)
			sequenceNumber := decodeSequenceNumber(header)
			expectedSequenceNumber := udpAckConn.sentSequencesNumbers.expected(addrIP)
			if sequenceNumber == expectedSequenceNumber {
				utils.Log.InfoPrintln("ACK correctly received")
				udpAckConn.sentSequencesNumbers.increment(addrIP, expectedSequenceNumber)
				return nil, packetWithoutHeader // Client correctly received the pktToSend
			} else if sequenceNumber < expectedSequenceNumber {
				utils.Log.InfoPrintln("Already received this ack -> doing nothing")
			} else {
				utils.Log.InfoPrintln("Ack received that was not the expected Ack, resending the packet. Expected ACK: ", expectedSequenceNumber, ", got ", sequenceNumber)
				stats.SimulationStats.ProtocolSent.Increment(addrIP)
				stats.SimulationStats.Nsent.Increment(addrIP)
				_, err := conn.WriteTo(packetWithHeader, addr)
				if err != nil {
					return err, nil
				}
			}
		case <-time.After(config.Timeout):
			utils.Log.WarningPrintln("Timeout on addr: ", addrIP, " resending pkt\n")
			stats.SimulationStats.Timeouts.Increment(addrIP)
			_, err := conn.WriteTo(packetWithHeader, addr)
			stats.SimulationStats.ProtocolSent.Increment(addrIP)
			stats.SimulationStats.Nsent.Increment(addrIP)
			if err != nil {
				return err, nil
			}
		}
	}
	return errors.New("the ACK was not received"), nil
}

func (udpAckConn *UDPAckConn) Close() error {
	return udpAckConn.conn.Close()
}

func (udpAckConn *UDPAckConn) handlePacket(addr *net.UDPAddr, packet []byte, handler UDPAckServerHandler) error {
	addrIP := addrtranslation.AddrToIPString(addr)
	stats.SimulationStats.Nreceived.Increment(addrIP)
	packetHeader, packetWithoutHeader := RemoveHeaderFromPacket(packet)
	packetType := DecodePacketType(packetHeader)
	if packetType == PacketTypeDataNoACK {
		handler(addr, packetWithoutHeader)
		return nil
	}

	sequenceNumber := decodeSequenceNumber(packetHeader)
	if packetType == PacketTypeAck {
		udpAckConn.handleAck(addrIP, packet)
		return nil
	}
	expectedSequenceNumber := udpAckConn.receivedSequencesNumbers.expected(addrIP)
	if sequenceNumber != expectedSequenceNumber {
		if sequenceNumber > expectedSequenceNumber && expectedSequenceNumber != 0 {
			// Rejects the incoming packet because it doesn't contain the expected sequence
			// number and the sequence number corresponds to a future packet. Doesn't return
			// an error because we just need to wait for the correct package to arrive
			utils.Log.WarningPrintln("The sequence number received is greather than the expected sequence number, therefore we do nothing and wait for the precedents packages to arrive")
			return nil
		} else {
			utils.Log.WarningPrintln("Already received this sequence number, sending the Ack")
			// We already received this packet, therefore we just send out the Ack to notify
			// the Addr that we correctly received the packet
			return udpAckConn.sendAck(addr, sequenceNumber)
		}
	}
	utils.Log.InfoPrintln("Received the expected sequence number, normal handler is called. Packet Type: ", [...]string{"Data", "DataNoACK", "ACK"}[packetType])
	// We received the packet with the expected sequence number, therefore we
	// dispatch it to the handler and send out the Ack.
	err := udpAckConn.sendAck(addr, sequenceNumber)
	if err != nil {
		return err
	}
	udpAckConn.receivedSequencesNumbers.increment(addrIP, sequenceNumber)
	// Note: we could use a goroutine here to speed up the process of the packet and
	// not block the next packets wainting to be processesd
	handler(addr, packetWithoutHeader)
	return nil
}

func (udpAckConn *UDPAckConn) sendAck(addr *net.UDPAddr, sequenceNumber uint8) error {
	addrIP := addrtranslation.AddrToIPString(addr)
	stats.SimulationStats.Nsent.Increment(addrIP)
	utils.Log.InfoPrintln("Sending the ACK to ", addr, " with sequence number ", sequenceNumber, "\n")
	ackPacket, err := newAckPacket(sequenceNumber)
	if err != nil {
		return err
	}
	_, err = udpAckConn.conn.WriteTo(ackPacket, addr)
	return err
}

func (udpAckConn *UDPAckConn) handleAck(addrIP addrtranslation.IPString, packet []byte) {
	if ackChan, in := udpAckConn.ackChannels[addrIP]; in {
		select {
		case ackChan <- packet:
			break
		//case <-time.After(5 * time.Second):
		default:
			utils.Log.WarningPrintln("Dropping the ACK because no goroutine is currently listening to ACK channel")
		}
	}
}

type UDPAckConnSendConfig struct {
	MaxRetries          int
	TimesBetweenRetries time.Duration
	Timeout             time.Duration
}

func newDefaultUDPAckConnSendConfig() *UDPAckConnSendConfig {
	return &UDPAckConnSendConfig{
		MaxRetries:          100,
		TimesBetweenRetries: 5 * time.Second,
		Timeout:             25 * time.Second,
	}
}
