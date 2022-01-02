CONTIKI_PROJECT = client
all: $(CONTIKI_PROJECT)

PROJECT_SOURCEFILES += schedule_updater.c

PLATFORMS_ONLY = cooja native

CONTIKI=../..

# force Security from command line
MAKE_WITH_SECURITY ?= 0

MAKE_MAC = MAKE_MAC_TSCH
MAKE_ROUTING = MAKE_ROUTING_RPL_LITE

include $(CONTIKI)/Makefile.dir-variables
# MODULES += $(CONTIKI_NG_SERVICES_DIR)/shell
MODULES += $(CONTIKI_NG_SERVICES_DIR)/rpl-border-router

# Include CoAP resources
MODULES_REL += ./resources

# Include the CoAP implementation
include $(CONTIKI)/Makefile.dir-variables
MODULES += $(CONTIKI_NG_APP_LAYER_DIR)/coap

ifeq ($(MAKE_WITH_SECURITY),1)
CFLAGS += -DWITH_SECURITY=1
endif

include $(CONTIKI)/Makefile.include
