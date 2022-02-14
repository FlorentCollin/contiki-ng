#ifndef UDPACK_SERVER_H_
#define UDPACK_SERVER_H_

#include "contiki.h"

// send_server sends a packet to the server. The packet is created
// using the encoder function which must encode the packet to send into
// the buffer and return the number of bytes inserted into it.
void send_server(uint16_t encoder(uint8_t *buffer));

#endif /* UDPACK_SERVER_H_ */