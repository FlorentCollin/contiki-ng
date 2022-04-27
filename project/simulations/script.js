var FileWriter = Java.type("java.io.FileWriter");

TIMEOUT(18000000, log.testOK());
var updateSpeedLimit = 1.0;
var speedLimit = 1000.0;
sim.setSpeedLimit(speedLimit);

var timestamp = Date.now()
var outputLog = new FileWriter("simulation" + timestamp +".log");

var completed = {};
var scheduleMsg = "schedule_updater_pkt";
var firstScheduleMsgTime = 0;
var completeMsg = "pkt->type = complete";
var motesCount = mote.getSimulation().getMotesCount();

var txMsg = "Sending";
var tx = Array(motesCount);
var rxMsg = "Received";
var rx = Array(motesCount);
var timeoutMsg = "Timer expired";
var timeouts = Array(motesCount);
var linkStatsMsg = "neighbor";
var linkStats = Array(motesCount);
for (var i = 0; i < motesCount; i++) {
    tx[i] = 0;
    rx[i] = 0;
    timeouts[i] = 0;
    linkStats[i] = {};
}


while(Object.keys(completed).length < motesCount) {
    outputLog.write(id + ":" + msg + "\n");
    if (msg.contains(completeMsg)) {
        completed[id] = true;
    } else if (firstScheduleMsgTime == 0 && msg.contains(scheduleMsg)) {
        firstScheduleMsgTime =  time;
    } else if (msg.contains(rxMsg)) {
        rx[id-1]++;
    } else if (msg.contains(txMsg)) {
        tx[id-1]++;
    } else if (msg.contains(timeoutMsg)) {
        timeouts[id-1]++;
    } else if (msg.contains("Switching parent")) {
        if (id === motesCount) {
            sim.setSpeedLimit(updateSpeedLimit);
            outputLog.write("SETTING SPEED LIMIT\n");
        }
    } else if (msg.contains(linkStatsMsg)) {
        splits = msg.split("{");
        try {

        stats = JSON.parse("{" + splits[splits.length - 1]);
        neighbor = stats.neighbor;
        delete stats.neighbor;
        linkStats[id-1][neighbor] = stats;
        } catch(e) {
            // do nothing
        }
    }
    YIELD();
}
var stats = {
    firstScheduleMsgTime: firstScheduleMsgTime,
    tx: tx,
    rx: rx,
    timeouts: timeouts,
    linkStats: linkStats,
    updateSpeedLimit: updateSpeedLimit,
    scheduleInstallationTime: ((time - firstScheduleMsgTime) / 1000000 / speedLimit),
}

log.log(JSON.stringify(stats, null, 4));
statsFile = new FileWriter("simulation-stats" + timestamp + ".json");
statsFile.write(JSON.stringify(stats, null, 4));
statsFile.close()
outputLog.close();

GENERATE_MSG(5000, "sleep"); //Wait for 5 secondes

YIELD_THEN_WAIT_UNTIL(msg.equals("sleep"));
log.testOK();
