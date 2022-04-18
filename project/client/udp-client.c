#include "contiki.h"
#include "etimer.h"
#include "net/mac/tsch/tsch.h"
#include "net/mac/tsch/tsch-schedule.h"
#include "schedule_updater.h"
#include "sys/log.h"

#include "topology-application.h"
#include "bandwidth-application.h"

#define LOG_MODULE "App"
#define LOG_LEVEL LOG_LEVEL_INFO

#define SEND_INTERVAL (20 * CLOCK_SECOND)


static struct tsch_slotframe *sf_common;

static void initialize_tsch_schedule() {
    tsch_schedule_create_minimal();
    sf_common = tsch_schedule_get_slotframe_by_handle(0);
    if (sf_common == NULL) {
        LOG_ERR("Couldn't create the initial slotframe\n");
    }
    tsch_schedule_add_link(sf_common, (LINK_OPTION_RX | LINK_OPTION_TX | LINK_OPTION_SHARED | LINK_OPTION_SHARED), LINK_TYPE_NORMAL,
    &tsch_broadcast_address, 1, 0, 1);
    tsch_schedule_add_link(sf_common, (LINK_OPTION_RX | LINK_OPTION_TX | LINK_OPTION_SHARED | LINK_OPTION_SHARED), LINK_TYPE_NORMAL,
    &tsch_broadcast_address, 2, 0, 1);
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
    bandwidth_application_start(5);
    PROCESS_END();
}
/*---------------------------------------------------------------------------*/
