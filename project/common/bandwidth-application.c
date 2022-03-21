#include "bandwidth-application.h"
#include "udpack-server.h"

#include "contiki.h"
#include "etimer.h"
#include "net/ipv6/uip-sr.h"
#include "net/ipv6/uip.h"
#include "net/routing/rpl-lite/rpl.h"
#include "sys/log.h"


#define LOG_MODULE "BandwidthApplication"
#define LOG_LEVEL LOG_LEVEL_INFO

PROCESS(bandwidth_application, "Bandwidth Application");

static uint8_t application_bandwidth = 0;

void bandwidth_application_start(uint8_t bandwidth) { 
    application_bandwidth = bandwidth;
    process_start(&bandwidth_application, NULL); 
}

static uint16_t encode_bandwidth(uint8_t* packet_buffer);

static struct etimer timer;
PROCESS_THREAD(bandwidth_application, ev, data) {
    PROCESS_BEGIN();
    LOG_INFO("Topology application started\n");
    etimer_set(&timer, 30 * CLOCK_SECOND);
    while (1) {
        PROCESS_WAIT_EVENT_UNTIL(etimer_expired(&timer));
        send_server(encode_bandwidth);
        etimer_reset(&timer);
    }
    PROCESS_END();
}


static uint16_t encode_bandwidth(uint8_t* packet_buffer) {
    LOG_DBG("Encoding the Bandwidth\n");
#define BANDWIDTH_APPLICATION_TYPE 2
    uint16_t index = 0;
    packet_buffer[index++] = BANDWIDTH_APPLICATION_TYPE;
    packet_buffer[index++] = application_bandwidth;
    return index;
}