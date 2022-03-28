#include "contiki.h"
#include "etimer.h"
#include "net/mac/tsch/tsch.h"
#include "schedule_updater.h"
#include "sys/log.h"

#include "topology-application.h"
#include "bandwidth-application.h"

#define LOG_MODULE "App"
#define LOG_LEVEL LOG_LEVEL_INFO

#define SEND_INTERVAL (20 * CLOCK_SECOND)



#define HELLO_WORLD_APPLICATION 3
// static uint16_t hello_world_encoder(uint8_t *buffer) {
//     static uint8_t i = 0;
//     LOG_INFO("Encoding hello world: %d\n", i);
//     buffer[0] = HELLO_WORLD_APPLICATION;
//     buffer[1] = i++;
//     return 2;
// }

#define APP_SLOTFRAME_HANDLE 0
#define APP_UNICAST_TIMESLOT 1
static struct tsch_slotframe *sf_common;
// static struct tsch_slotframe *sf_next;

static void initialize_tsch_schedule() {
    sf_common = tsch_schedule_add_slotframe(APP_SLOTFRAME_HANDLE, APP_SLOTFRAME_SIZE);
    tsch_schedule_add_link(
        sf_common,
        (LINK_OPTION_RX | LINK_OPTION_TX | LINK_OPTION_SHARED | LINK_OPTION_TIME_KEEPING),
        LINK_TYPE_ADVERTISING, &tsch_broadcast_address, 0, 0, 1);
    for (int i = 1; i < APP_SLOTFRAME_SIZE; i++) {
        for (int j = 0; j < 4; j++) {
            struct tsch_link *link = tsch_schedule_add_link(
                sf_common, (LINK_OPTION_RX | LINK_OPTION_TX | LINK_OPTION_SHARED), LINK_TYPE_NORMAL,
                &tsch_broadcast_address, i, j, 1);

            if (link == NULL) {
                LOG_ERR("Couldn't had the link with (timeslot, channel): (%d, %d)\n", i, j);
            }
        }
    }
}

/*---------------------------------------------------------------------------*/
PROCESS(udp_client_process, "UDP client");
AUTOSTART_PROCESSES(&udp_client_process);
/*---------------------------------------------------------------------------*/
/*---------------------------------------------------------------------------*/
PROCESS_THREAD(udp_client_process, ev, data) {
    PROCESS_BEGIN();
    initialize_tsch_schedule();
    topology_application_start();
    bandwidth_application_start(10);
    PROCESS_END();
}
/*---------------------------------------------------------------------------*/
