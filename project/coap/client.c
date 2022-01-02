#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "coap-observe-client.h"
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
#define LOG_MODULE "App"
#define LOG_LEVEL LOG_LEVEL_INFO

#define SERVER_EP "coap://[fd00::202:2:2:2]"

PROCESS(client_process, "Coap Example Client");
AUTOSTART_PROCESSES(&client_process);

#define APP_SLOTFRAME_HANDLE 1
#define APP_UNICAST_TIMESLOT 1
static struct tsch_slotframe *sf_common;

static void initialize_tsch_schedule() {
    sf_common = tsch_schedule_add_slotframe(APP_SLOTFRAME_HANDLE, APP_SLOTFRAME_SIZE);
            tsch_schedule_add_link(
                sf_common, (LINK_OPTION_RX | LINK_OPTION_TX | LINK_OPTION_SHARED | LINK_OPTION_TIME_KEEPING),
                LINK_TYPE_ADVERTISING, &tsch_broadcast_address, 0, 0, 1);
}

/* CoAP config */
static coap_endpoint_t server_ep;
static coap_observee_t *obs;
#define NUMBER_OF_URLS 1
char *service_urls[NUMBER_OF_URLS] = {"/schedule"};

/* This function is will be passed to COAP_BLOCKING_REQUEST() to handle responses. */
void lient_chunk_handler(void *response) {
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

/*----------------------------------------------------------------------------*/
/*
 * Handle the response to the observe request and the following notifications
 */
static void notification_callback(coap_observee_t *obs, void *notification, coap_notification_flag_t flag) {
  int len = 0;
  const uint8_t *payload = NULL;

  LOG_INFO("Notification handler\n");
  LOG_INFO("Observee URI: %s\n", obs->url);
  if(notification) {
    len = coap_get_payload(notification, &payload);
  }
  switch(flag) {
  case NOTIFICATION_OK:
    LOG_INFO("NOTIFICATION OK: %*s\n", len, (char *)payload);
    client_chunk_handler(notification)
    break;
  case OBSERVE_OK: /* server accepeted observation request */
    LOG_INFO("OBSERVE_OK: %*s\n", len, (char *)payload);
    break;
  case OBSERVE_NOT_SUPPORTED:
    LOG_INFO("OBSERVE_NOT_SUPPORTED: %*s\n", len, (char *)payload);
    obs = NULL;
    break;
  case ERROR_RESPONSE_CODE:
    LOG_INFO("ERROR_RESPONSE_CODE: %*s\n", len, (char *)payload);
    obs = NULL;
    break;
  case NO_REPLY_FROM_SERVER:
    LOG_INFO("NO_REPLY_FROM_SERVER: "
           "removing observe registration with token %x%x\n",
           obs->token[0], obs->token[1]);
    obs = NULL;
    break;
  }
}

/*----------------------------------------------------------------------------*/
/*
 * Toggle the observation of the remote resource
 */
void toggle_observation(void) {
  if(obs) {
    LOG_INFO("Stopping observation\n");
    coap_obs_remove_observee(obs);
    obs = NULL;
  } else {
    LOG_INFO("Starting observation\n");
    obs = coap_obs_request_registration(&server_ep, service_urls[0], notification_callback, NULL);
  }
}

/*----------------------------------------------------------*/
PROCESS_THREAD(client_process, ev, data) {
    PROCESS_BEGIN();

    initialize_tsch_schedule();

    static coap_message_t request[1]; /* This way the packet can be treated as pointer as usual. */
    coap_endpoint_parse(SERVER_EP, strlen(SERVER_EP), &server_ep);

    static uip_ipaddr_t ipaddr;
    while (1) {
        PROCESS_YIELD();
        if (ev == button_hal_release_event) {
            NETSTACK_ROUTING.get_root_ipaddr(&ipaddr);
            toggle_observation()
        }
    }

    PROCESS_END();
}
