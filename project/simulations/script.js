var FileWriter = Java.type("java.io.FileWriter");

TIMEOUT(18000000, log.testOK());
var speedLimit = 1.0;
//var speedLimit = 50.0;
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
for (var i = 0; i < motesCount; i++) {
    tx[i] = 0;
    rx[i] = 0;
    timeouts[i] = 0;
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
    }
    YIELD();
}
var stats = {}
stats.firstScheduleMsgTime = firstScheduleMsgTime;
stats.tx = tx;
stats.rx = rx;
stats.timeouts = timeouts;
stats.scheduleInstallationTime = ((time - firstScheduleMsgTime) / 1000000 / speedLimit);
stats.speedLimit = speedLimit;

log.log(JSON.stringify(stats, null, 4));
statsFile = new FileWriter("simulation-stats" + timestamp + ".json");
statsFile.write(JSON.stringify(stats, null, 4));
statsFile.close()
outputLog.close();

if (speedLimit == 1) {
    GENERATE_MSG(5000, "sleep"); //Wait for 5 sec
} else {
    GENERATE_MSG(2000 * speedLimit, "sleep");
}

YIELD_THEN_WAIT_UNTIL(msg.equals("sleep"));
log.testOK();
