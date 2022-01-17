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
#include "sys/etimer.h"

#include "schedule_updater.h"

/* Log configuration */
#define LOG_MODULE "App"
#define LOG_LEVEL LOG_LEVEL_INFO

#define SERVER_EP "coap://[fd00::1]"
// #define SERVER_EP "coap://[fd00::201:1:1:1]"

PROCESS(client_process, "Coap Example Client");
AUTOSTART_PROCESSES(&client_process);

#define APP_SLOTFRAME_HANDLE 1
#define APP_UNICAST_TIMESLOT 1
static struct tsch_slotframe *sf_common;
// static struct tsch_slotframe *sf_next;

static void initialize_tsch_schedule() {
    sf_common = tsch_schedule_add_slotframe(APP_SLOTFRAME_HANDLE, APP_SLOTFRAME_SIZE);
    tsch_schedule_add_link(
        sf_common, (LINK_OPTION_RX | LINK_OPTION_TX | LINK_OPTION_SHARED | LINK_OPTION_TIME_KEEPING),
        LINK_TYPE_ADVERTISING, &tsch_broadcast_address, 0, 0, 1);
    for (int i = 1; i < APP_SLOTFRAME_SIZE; i++) {
      for (int j = 0; j < 4; j++) {
        struct tsch_link *link = tsch_schedule_add_link(
            sf_common, (LINK_OPTION_RX | LINK_OPTION_TX | LINK_OPTION_SHARED),
            LINK_TYPE_NORMAL, &tsch_broadcast_address, i, j, 1);

        if (link == NULL) {
            LOG_ERR("Couldn't had the link with (timeslot, channel): (%d, %d)\n", i, j);
        }
      }
    }
}

/* CoAP config */
static coap_endpoint_t server_ep;
static coap_observee_t *obs;


/* This function is will be passed to COAP_BLOCKING_REQUEST() to handle responses. */
void client_chunk_handler(const uint8_t *pkt) {
    // update_pkt_log(pkt);
    switch (update_pkt_type(pkt)) {
      case schedule_updater_pkt_type_update:
            LOG_INFO("Update pkt\n");
            break;
      case schedule_updater_pkt_type_update_complete:
            LOG_INFO("Complete pkt\n");
            break;

    }
    #if 0
    int err;
    switch (update_pkt_type(pkt)) {
        case schedule_updater_pkt_type_update:
            passed = true;
            update_pkt_add_cells(pkt, sf_common);
            break;
        case schedule_updater_pkt_type_update_complete:
            if (passed) {
              LOG_INFO("Update Complete\n");
            }
            err = tsch_schedule_remove_slotframe(sf_common);
            if (err == 0) {
                LOG_ERR("Couldn't remove the tsch slot_frame");
                return;
            }
            sf_common = sf_next;
            sf_next = tsch_schedule_add_slotframe(SCHEDULE_UPDATER_SLOTFRAME_HANDLE, SCHEDULE_UPDATER_SLOTFRAME_SIZE);
            if (sf_next == NULL) {
                LOG_ERR("Couldn't add a new slotframe\n");
                return;
            }
            tsch_schedule_add_link(
                sf_next, (LINK_OPTION_RX | LINK_OPTION_TX | LINK_OPTION_SHARED | LINK_OPTION_TIME_KEEPING),
                LINK_TYPE_ADVERTISING, &tsch_broadcast_address, 0, 0, 1);
            break;
        default:
            LOG_ERR("SHOULD NEVER HAPPEN, unhandled switch case statement\n");
            return;
    }
    // LOG_INFO("Payload decoded %d\n", c++);
    #endif
}


/*----------------------------------------------------------------------------*/
/*
 * Handle the response to the observe request and the following notifications
 */
static void notification_callback(coap_observee_t *obs, void *notification, coap_notification_flag_t flag) {
  int len = 0;
  const uint8_t *payload = NULL;

  LOG_INFO("Notification handler\n");
  if(notification) {
    len = coap_get_payload(notification, &payload);
  }

  switch(flag) {
  case NOTIFICATION_OK:
    LOG_INFO("NOTIFICATION OK, payload size: %d\n", len);
    // LOG_INFO("Payload: %*s\n", len - 3, (char*) (payload + 3));
    // LOG_INFO("Payload: ");
    // for (int i = 0; i < len; i++) {
      // printf("%x:", payload[i]);
    // }
    // printf("\n");
    client_chunk_handler(payload + 3);
    //update_pkt_log(payload + 3);
    break;
  case OBSERVE_OK: /* server accepeted observation request */
    LOG_INFO("OBSERVE_OK, payload size: : %d\n", len);
    break;
  case OBSERVE_NOT_SUPPORTED:
    LOG_INFO("OBSERVE_NOT_SUPPORTED, payload size: %d\n", len);
    obs = NULL;
    break;
  case ERROR_RESPONSE_CODE:
    LOG_INFO("ERROR_RESPONSE_CODE, payload size: %d\n", len);
    /*obs = NULL;*/
    /*coap_obs_remove_observee(obs);*/
    break;
  case NO_REPLY_FROM_SERVER:
    LOG_INFO("NO_REPLY_FROM_SERVER: "
           "removing observe registration with token %x%x\n",
           obs->token[0], obs->token[1]);
    coap_obs_remove_observee(obs);
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
    obs = coap_obs_request_registration(&server_ep, "/schedule", notification_callback, NULL);
  }
}

/*----------------------------------------------------------*/
PROCESS_THREAD(client_process, ev, data) {
    static struct etimer etimer;
    PROCESS_BEGIN();
    etimer_set(&etimer, 6 * 60 * CLOCK_SECOND);

    initialize_tsch_schedule();

    coap_endpoint_parse(SERVER_EP, strlen(SERVER_EP), &server_ep);

    // toggle observation after 6 min.
    PROCESS_WAIT_EVENT_UNTIL(etimer_expired(&etimer));
    toggle_observation();
    LOG_INFO("First toggle obsersvation done\n");
    while (1) {
        PROCESS_YIELD();
        if (ev == button_hal_release_event) {
            toggle_observation();
        }
    }

    PROCESS_END();
}
