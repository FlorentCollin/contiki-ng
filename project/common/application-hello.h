#ifndef HELLO_APPLICATION_H
#define HELLO_APPLICATION_H

#include "application-type.h"
#include "contiki.h"
#include "udpack-server.h"

PROCESS(hello_application, "Hello Application");

void hello_application_start() { 
    process_start(&hello_application, NULL); 
}

static uint16_t encode_hello(uint8_t* packet_buffer);

static struct etimer timer;
PROCESS_THREAD(hello_application, ev, data) {
    PROCESS_BEGIN();
    etimer_set(&timer, 30 * CLOCK_SECOND);
    while (1) {
        PROCESS_WAIT_EVENT_UNTIL(etimer_expired(&timer));
        send_server(AppTypeHello, encode_hello);
        etimer_reset(&timer);
    }
    PROCESS_END();
}


static uint16_t encode_hello(uint8_t* packet_buffer) {
    static char charset[27] = "abcdefghijklmnopqrstuvwxyz";
    uint16_t index = 0;
    for (int i = 0; i < 70; i++) {
        packet_buffer[index++] = charset[i%26];
    }
    return index;
}

#endif /* HELLO_APPLICATION_H */