#ifndef PROJECT_CONF_H_
#define PROJECT_CONF_H_

/* #define UIP_CONF_BUFFER_SIZE 127 */

/* Enable 6tisch minimal cell */
#define TSCH_SCHEDULE_CONF_WITH_6TISCH_MINIMAL 1

/* Size of the application-specific schedule; a number relatively prime to the hopseq length */
/* Same as TSCH_SCHEDULE_CONF_DEFAULT_LENGTH */
#define APP_SLOTFRAME_SIZE 101

/* Disable security */
#define LLSEC802154_CONF_ENABLED 0

/* Logging */
#define LOG_CONF_LEVEL_RPL                         LOG_LEVEL_ERR
#define LOG_CONF_LEVEL_TCPIP                       LOG_LEVEL_ERR
#define LOG_CONF_LEVEL_IPV6                        LOG_LEVEL_ERR
#define LOG_CONF_LEVEL_6LOWPAN                     LOG_LEVEL_ERR
#define LOG_CONF_LEVEL_MAC                         LOG_LEVEL_ERR
#define LOG_CONF_LEVEL_FRAMER                      LOG_LEVEL_ERR
#define LOG_CONF_LEVEL_COAP                        LOG_LEVEL_ERR
#define LOG_LEVEL_APP                              LOG_LEVEL_DBG
#define TSCH_LOG_CONF_PER_SLOT                     0

#define LINKADDR_CONF_SIZE 8
/*  #define UIP_CONF_BUFFER_SIZE 127 */
/* Disable fragmentation? */
/* #define SICSLOWPAN_CONF_FRAGMENT_BUFFERS 1 */

#endif /* PROJECT_CONF_H_ */
