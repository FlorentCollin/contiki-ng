#include <stdlib.h>
#include <string.h>
#include "coap-engine.h"

#include "sys/log.h"
#define LOG_MODULE "App"
#define LOG_LEVEL LOG_LEVEL_INFO

#include "schedule_updater.h"

static void res_get_handler(coap_message_t *request, coap_message_t *response, uint8_t *buffer, uint16_t preferred_size, int32_t *offset);

/*
 * A handler function named [resource name]_handler must be implemented for each RESOURCE.
 * A buffer for the response payload is provided through the buffer pointer. Simple resources can ignore
 * preferred_size and offset, but must respect the REST_MAX_CHUNK_SIZE limit for the buffer.
 * If a smaller block size is requested for CoAP, the REST framework automatically splits the data.
 */
RESOURCE(res_hello,
         "title=\"Hello world\";rt=\"Text\"",
         res_get_handler,
         NULL,
         NULL,
         NULL);

static void
res_get_handler(coap_message_t *request, coap_message_t *response, uint8_t *buffer, uint16_t preferred_size, int32_t *offset) {
    static struct schedule_updater_pkt update_pkt = {
        .cell_ids_len = 1,
        .cell_ids = {{
            .link_options = LINK_OPTION_RX,
            .timeslot = 3,
            .channel = 2,
        }}
    };

    uint16_t length = schedule_updater_pkt_size_needed(&update_pkt);
    if (length > REST_MAX_CHUNK_SIZE) {
        LOG_ERR("Packet size is too big");
        return;
    }
    schedule_updater_pkt_encode(buffer, &update_pkt);

    coap_set_header_content_format(response, TEXT_PLAIN); /* text/plain is the default, hence this option could be omitted. */
    coap_set_header_etag(response, (uint8_t *)&length, 1);
    char hello[] = "Hello";
    coap_set_payload(response, &hello, 5);
}
