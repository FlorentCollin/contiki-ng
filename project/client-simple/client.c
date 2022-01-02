#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "coap-blocking-api.h"
#include "coap-engine.h"
#include "coap-log.h"
#include "contiki-net.h"
#include "contiki.h"
#include "dev/button-hal.h"
#include "net/ipv6/uip.h"
#include "net/mac/tsch/tsch.h"
#include "net/routing/rpl-lite/rpl-dag-root.h"


// Cooja node id
#include "sys/node-id.h"

#include "schedule_updater.h"

/* Log configuration */
#define LOG_MODULE "Client"
#define LOG_LEVEL LOG_LEVEL_INFO

#define SERVER_EP "coap://[fd00::1]"

PROCESS(client_process, "Coap Example Client");
AUTOSTART_PROCESSES(&client_process);

#define APP_SLOTFRAME_HANDLE 1
#define APP_UNICAST_TIMESLOT 1
static struct tsch_slotframe *sf_common;

static void initialize_tsch_schedule() {
    LOG_INFO("LinkADDR SIZE: %d\n", LINKADDR_SIZE);
    sf_common = tsch_schedule_add_slotframe(APP_SLOTFRAME_HANDLE, APP_SLOTFRAME_SIZE);
            tsch_schedule_add_link(
                sf_common, (LINK_OPTION_RX | LINK_OPTION_TX | LINK_OPTION_SHARED | LINK_OPTION_TIME_KEEPING),
                LINK_TYPE_ADVERTISING, &tsch_broadcast_address, 0, 0, 1);
}

#define NUMBER_OF_URLS 1
/* leading and ending slashes only for demo purposes, get cropped automatically when setting the
 * Uri-Path */
char *service_urls[NUMBER_OF_URLS] = {"/schedule"};

/* This function is will be passed to COAP_BLOCKING_REQUEST() to handle responses. */
void client_chunk_handler(coap_message_t *response) {
    const uint8_t *payload;

    if (response == NULL) {
        LOG_INFO("Request timed out");
        return;
    }

    coap_get_payload(response, &payload);
    static struct schedule_updater_pkt update_pkt;
    schedule_updater_pkt_decode(&update_pkt, payload);
    schedule_updater_pkt_log(&update_pkt);
    schedule_updater_pkt_add_cells(&update_pkt, sf_common);
}

PROCESS_THREAD(client_process, ev, data) {
    static coap_endpoint_t server_ep;
    PROCESS_BEGIN();

    initialize_tsch_schedule();

    static coap_message_t request[1]; /* This way the packet can be treated as pointer as usual. */
    coap_endpoint_parse(SERVER_EP, strlen(SERVER_EP), &server_ep);

    static uip_ipaddr_t ipaddr;
    while (1) {
        PROCESS_YIELD();
        if (ev == button_hal_release_event) {
            NETSTACK_ROUTING.get_root_ipaddr(&ipaddr);
            LOG_INFO_6ADDR(&ipaddr);
            printf("\n");
            coap_init_message(request, COAP_TYPE_CON, COAP_GET, 0);
            coap_set_header_uri_path(request, service_urls[0]);

            LOG_INFO("--Requesting %s--\n", service_urls[0]);

            COAP_BLOCKING_REQUEST(&server_ep, request, client_chunk_handler);

            LOG_INFO("--Done--\n");
        }
    }

    PROCESS_END();
}
