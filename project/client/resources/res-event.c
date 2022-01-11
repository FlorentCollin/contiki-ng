#include <stdio.h>
#include <string.h>
#include "coap-engine.h"
#include "coap.h"
#include "schedule_updater.h"

/* Log configuration */
#include "sys/log.h"
#define LOG_MODULE "App"
#define LOG_LEVEL LOG_LEVEL_APP

static void res_get_handler(coap_message_t *request, coap_message_t *response, uint8_t *buffer, uint16_t preferred_size, int32_t *offset);
static void res_event_handler(void);

/*
 * Example for an event resource.
 * Additionally takes a period parameter that defines the interval to call [name]_periodic_handler().
 * A default post_handler takes care of subscriptions and manages a list of subscribers to notify.
 */
EVENT_RESOURCE(res_event,
               "title=\"Event demo\";obs",
               res_get_handler,
               NULL,
               NULL,
               NULL,
               res_event_handler);

/*
 * Use local resource state that is accessed by res_get_handler() and altered by res_event_handler() or PUT or POST.
 */
static int32_t event_counter = 0;

static struct schedule_updater_pkt pkt = {
    .neighbor_addr = {{}},
    .cell_count = 1,
    .cells = {{
        .link_options = LINK_OPTION_TX,
        .timeslot = 5,
        .channel = 3,
        }
    }};

static void
res_get_handler(coap_message_t *request, coap_message_t *response, uint8_t *buffer, uint16_t preferred_size, int32_t *offset) {
    coap_set_header_content_format(response, TEXT_PLAIN);
    coap_set_payload(response, buffer, snprintf((char *)buffer, preferred_size, "EVENT %lu", (unsigned long) event_counter));
    schedule_updater_pkt_encode(buffer, &pkt);
    coap_set_payload(response, buffer, schedule_updater_pkt_size_needed(&pkt));

    /* A post_handler that handles subscriptions/observing will be called for periodic resources by the framework. */
}

/*
 * Additionally, res_event_handler must be implemented for each EVENT_RESOURCE.
 * It is called through <res_name>.trigger(), usually from the server process.
 */
static void
res_event_handler(void) {
    LOG_INFO("res_event_handler called\n");
    /* Do the update triggered by the event here, e.g., sampling a sensor. */
    ++event_counter;

    /* Usually a condition is defined under with subscribers are notified, e.g., event was above a threshold. */
    if(1) {
        LOG_DBG("TICK %u for /%s\n", (unsigned)event_counter, res_event.url);

        /* Notify the registered observers which will trigger the res_get_handler to create the response. */
        coap_notify_observers(&res_event);
    }
}

