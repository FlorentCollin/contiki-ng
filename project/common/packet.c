#include "contiki.h"
#include "packet.h"

uint16_t new_ack_packet(uint8_t* packet_buffer, uint8_t sequence_number) {
    Header header = new_header_with_sequence_number(PacketTypeAck, sequence_number);
    packet_buffer[0] = header;
    return sizeof(header);
}

uint16_t new_data_packet(uint8_t* packet_buffer, uint8_t sequence_number) {
    Header header = new_header_with_sequence_number(PacketTypeData, sequence_number);
	packet_buffer[0] = header;
    return sizeof(header);
}

uint16_t new_data_noack_packet(uint8_t* packet_buffer) {
    Header header = new_header(PacketTypeDataNoACK);
    packet_buffer[0] = header;
    return sizeof(header);
}

Header new_header_with_sequence_number(enum PacketType packet_type, uint8_t sequence_number) {
    Header header = 0;
    encode_packet_type(&header, packet_type);
    encode_sequence_number(&header, sequence_number);
    return header;
}

Header new_header(enum PacketType packet_type) {
    Header header = 0;
    encode_packet_type(&header, packet_type);
    return header;
}

void encode_packet_type(Header* header, enum PacketType packet_type) {
#define PACKET_TYPE_SHIFT 6
    *header |= (uint8_t) packet_type << PACKET_TYPE_SHIFT;
}

enum PacketType decode_packet_type(Header header) {
#define PACKET_TYPE_SHIFT 6
#define PACKET_TYPE_MASK 0b11000000
    return (header & PACKET_TYPE_MASK) >> PACKET_TYPE_SHIFT;
}

int encode_sequence_number(Header* header, uint8_t sequence_number) {
	if (sequence_number > SEQUENCE_NUMBER_MAX) {
        return -1;
	}
	*header |= sequence_number;
    return 0;
}

uint8_t decode_sequence_number(Header header) {
#define SEQUENCE_NUMBER_MASK 0b00111111
    return header & SEQUENCE_NUMBER_MASK;
}

// RemoveHeaderFromPacket remove the header that contains a PacketType and a SequenceNumber
// The header is a byte that has the following structure:
// 0b01 000111
//   ^  ^^^^^^
//   | |
//   | +------------ Sequence Number (last 6 bits)
//   +-- PacketType (first two bits)
void remove_header_from_packet(const uint8_t* packet, Header* header) {
    *header = packet[0];
}