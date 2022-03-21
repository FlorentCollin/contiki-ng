package udpack

import (
	"coap-server/utils"
	"log"
	"net"
	"time"
)

// UDPAckConn handles UDP connections with ACK of packets. Only one UDP packet
// can be flying at the same time.
type UDPAckConn struct {
	conn                     *net.UDPConn
	receivedSequencesNumbers SequencesNumbersMap
	sentSequencesNumbers     SequencesNumbersMap
	config                   *UDPAckConnSendConfig
	ackChannels              map[IPString]chan []byte
}

func NewUDPAckServer(conn *net.UDPConn, config *UDPAckConnSendConfig) *UDPAckConn {
	if config == nil {
		config = newDefaultUDPAckConnSendConfig()
	}
	return &UDPAckConn{
		conn:                     conn,
		receivedSequencesNumbers: make(SequencesNumbersMap),
		sentSequencesNumbers:     make(SequencesNumbersMap),
		config:                   config,
		ackChannels:              make(map[IPString]chan []byte),
	}
}

// UDPAckServerHandler is the callback called when receinving a packet. The
// packet can be of two types `PacketTypeData` which corresponds to a data packet
// and `PacketTypeAck` which corresponds to an Ack packet. The Ack packet must be
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
		logger.InfoPrintln("Message received from ", remoteAddrString)
		packet := buffer[:rlen]

		err = udpAckConn.handlePacket(remote, packet, handler)
		if err != nil {
			return err
		}
	}
}

func (udpAckConn *UDPAckConn) WriteTo(packet []byte, addr *net.UDPAddr) error {
	logger.InfoPrintln("Size of the packet to send: ", len(packet))
	addrIP := AddrToIPString(addr)
	ackChan, in := udpAckConn.ackChannels[addrIP]
	if !in {
		ackChan = make(chan []byte)
		udpAckConn.ackChannels[addrIP] = ackChan
	}
	nextSequenceNumber := udpAckConn.sentSequencesNumbers.expected(addrIP)
	packetWithHeader, err := newDataPacket(nextSequenceNumber, packet)
	if err != nil {
		return err
	}
	logger.InfoPrintln("Size of packet: ", len(packetWithHeader))
	config := udpAckConn.config
	conn := udpAckConn.conn
	for i := 0; i < config.MaxRetries; i++ {
		_, err = conn.WriteTo(packetWithHeader, addr)
		if err != nil {
			logger.ErrorPrintln("Error while trying to send the packet to the udpack: ", err)
			logger.ErrorPrintln("Retrying in ", config.TimesBetweenRetries)
			time.Sleep(config.TimesBetweenRetries)
		} else {
			break
		}
	}
	if err != nil {
		log.Panic("Error maximum numbre of retries exhausted panicking with error : ", err)
	}
	// Waiting for the ack to be received
	for {
		select {
		case pkt := <-ackChan:
			header, _ := RemoveHeaderFromPacket(pkt)
			sequenceNumber := decodeSequenceNumber(header)
			expectedSequenceNumber := udpAckConn.sentSequencesNumbers.expected(addrIP)
			logger.InfoPrintln("Expecting ACK: ", expectedSequenceNumber, " got ", sequenceNumber, '\n')
			if sequenceNumber == expectedSequenceNumber {
				logger.InfoPrintln("ACK correctly received")
				udpAckConn.sentSequencesNumbers.increment(addrIP, expectedSequenceNumber)
				return nil // Client correctly received the pktToSend
			} else if sequenceNumber < expectedSequenceNumber {
				logger.InfoPrintln("Already received this ack -> doing nothing")
			} else {
				logger.InfoPrintln("Ack received that was not the expected Ack, resending the packet")
				logger.InfoPrintln("Message received: ", string(pkt))
				_, err := conn.WriteTo(packetWithHeader, addr)
				if err != nil {
					return err
				}
			}
		case <-time.After(config.Timeout):
			logger.WarningPrintln("Timeout on addr %s, resending pkt\n", addr.IP.String())
			_, err := conn.WriteTo(packetWithHeader, addr)
			if err != nil {
				return err
			}
		}
	}
}

func (udpAckConn *UDPAckConn) Close() error {
	return udpAckConn.conn.Close()
}

func (udpAckConn *UDPAckConn) handlePacket(addr *net.UDPAddr, packet []byte, handler UDPAckServerHandler) error {
	addrIP := AddrToIPString(addr)
	packetHeader, packetWithoutHeader := RemoveHeaderFromPacket(packet)
	packetType := DecodePacketType(packetHeader)
	logger.InfoPrintln("Packet Type:", [...]string{"Data", "ACK"}[packetType])

	sequenceNumber := decodeSequenceNumber(packetHeader)
	logger.InfoPrintln("Sequence number associated with the packet:", sequenceNumber)
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
			logger.WarningPrintln("The sequence number received is greather than the expected sequence number, therefore we do nothing and wait for the precedents packages to arrive")
			return nil
		} else {
			logger.WarningPrintln("Already received this sequence number, sending the Ack")
			// We already received this packet, therefore we just send out the Ack to notify
			// the Addr that we correctly received the packet
			return udpAckConn.sendAck(addr, sequenceNumber)
		}
	}
	logger.InfoPrintln("Received the expected sequence number, normal handler is called")
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
	logger.InfoPrintln("Sending the ACK to %s with sequence number %d\n", addr, sequenceNumber)
	ackPacket, err := newAckPacket(sequenceNumber)
	if err != nil {
		return err
	}
	_, err = udpAckConn.conn.WriteTo(ackPacket, addr)
	return err
}

func (udpAckConn *UDPAckConn) handleAck(addrIP IPString, packet []byte) {
	if ackChan, in := udpAckConn.ackChannels[addrIP]; in {
		select {
		case ackChan <- packet:
			break
		default:
			logger.WarningPrintln("Dropping the ACK because no goroutine is currently listening to ACK channel")
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
		MaxRetries:          4,
		TimesBetweenRetries: 4 * time.Second,
		Timeout:             4 * time.Second,
	}
}

var logger = utils.NewLogger(utils.LogLevelInfo, utils.WHITE)
