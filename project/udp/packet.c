#include "contiki.h"
#include "packet.h"

uint16_t new_ack_packet(uint8_t* packet_buffer, uint8_t sequence_number) {
    Header header = new_header(PacketTypeAck, sequence_number);
    packet_buffer[0] = header;
    return sizeof(header);
}

uint16_t new_data_packet(uint8_t* packet_buffer, uint8_t sequence_number) {
    Header header = new_header(PacketTypeData, sequence_number);
	packet_buffer[0] = header;
    return sizeof(header);
}

Header new_header(enum PacketType packet_type, uint8_t sequence_number) {
    Header header = 0;
    encode_packet_type(&header, packet_type);
    encode_sequence_number(&header, sequence_number);
    return header;
}

void encode_packet_type(Header* header, enum PacketType packet_type) {
#define PACKET_TYPE_SHIFT 7
	if (packet_type == PacketTypeData) {
		*header |= (uint8_t) PacketTypeData << PACKET_TYPE_SHIFT;
	} else if (packet_type == PacketTypeAck) {
		*header |= (uint8_t) PacketTypeAck << PACKET_TYPE_SHIFT;
	}
}

enum PacketType decode_packet_type(Header header) {
#define PACKET_TYPE_MASK 0x80
    int is_ack = (header & PACKET_TYPE_MASK) != 0;
	if (is_ack) {
		return PacketTypeAck;
	}
	return PacketTypeData;
}

int encode_sequence_number(Header* header, uint8_t sequence_number) {
	if (sequence_number > SEQUENCE_NUMBER_MAX) {
        return -1;
	}
	*header |= sequence_number;
    return 0;
}

uint8_t decode_sequence_number(Header header) {
#define SEQUENCE_NUMBER_MASK 0x7F
    return header & SEQUENCE_NUMBER_MASK;
}

// RemoveHeaderFromPacket remove the header that contains a PacketType and a SequenceNumber
// The header is a byte that has the following structure:
// 0b1 0000111
//   ^ ^^^^^^^
//   | |
//   | +------------ Sequence Number (last 7 bits)
//   +-- PacketType (first bit)
void remove_header_from_packet(const uint8_t* packet, Header* header) {
    *header = packet[0];
}