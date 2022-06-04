#ifndef APPLICATION_TYPE_H
#define APPLICATION_TYPE_H

// ApplicationType encode the application type. This type is used by the server
// to dispatch incoming packet to the correct application. This type is encoded in the
// packet sent by the nodes.
enum ApplicationType { 
    AppTypeGraph,
    AppTypeTopology,
    AppTypeBandwidth,
    AppTypeHello,
    AppTypeAll
};

#endif
