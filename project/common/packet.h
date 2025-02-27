#ifndef PACKET_H_
#define PACKET_H_

// Packet contains helper functions to handle packets as described in my master thesis.
// Packets contains a header with the type and the sequence number of the packet.

#include "contiki.h"

#define SEQUENCE_NUMBER_MAX 0b00111111

typedef uint8_t Header;

enum PacketType {
    PacketTypeData,
    PacketTypeDataNoACK,
    PacketTypeAck
};


uint16_t new_ack_packet(uint8_t* packet_buffer, uint8_t sequence_number);

uint16_t new_data_packet(uint8_t* packet_buffer, uint8_t sequence_number);

Header new_header_with_sequence_number(enum PacketType packet_type, uint8_t sequence_number);

Header new_header(enum PacketType packet_type);

void encode_packet_type(Header* header, enum PacketType packet_type);

enum PacketType decode_packet_type(Header header);

int encode_sequence_number(Header* header, uint8_t sequence_number);

uint8_t decode_sequence_number(Header header);

// RemoveHeaderFromPacket remove the header that contains a PacketType and a SequenceNumber
// The header is a byte that has the following structure:
// 0b01 0000111
//   ^  ^^^^^^^
//   |  |
//   |  +------------ Sequence Number (last 6 bits)
//   +-- PacketType (two first bits)
void remove_header_from_packet(const uint8_t* packet, Header* header);

#endif /* PACKET_H_ */
