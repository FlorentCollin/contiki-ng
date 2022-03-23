#const nNodes = 1.
#const nSlots = 21.
#const nChannels = 16.

node(0..nNodes).
graph(1, 0). 

bandwidth(1, 2). bandwidth(0, 2). 
totalBandwidth(4).
