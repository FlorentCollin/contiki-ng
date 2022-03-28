#ifndef UDPACK_SERVER_H_
#define UDPACK_SERVER_H_

#include "contiki.h"

// init_udpack_server initialize the udp with one ack server.
// This function must be called before using `send_server` or
// `send_server_ack`.
void init_udpack_server();

// send_server_ack sends a packet to the server and wait for an ACK. 
//The packet is created using the encoder function which must encode 
// the packet to send into // the buffer and return the number of bytes 
// inserted into it.
void send_server_ack(uint16_t encoder(uint8_t *buffer));

// send_server sends a packet to the server without waiting for an ACK.
// The packet is created // using the encoder function which must encode 
// the packet to send into the buffer and return the number of bytes 
// inserted into it.
void send_server(uint16_t encoder(uint8_t *buffer));

#endif /* UDPACK_SERVER_H_ */