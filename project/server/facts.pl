#const nNodes = 2.
#const nSlots = 5.
#const nChannels = 5.

node(0..nNodes).
graph(1, 0). graph(2, 1). 
topology(1, 0). topology(0, 1). topology(2, 1). topology(1, 2). 
bandwith(N, 10) :- node(N).
totalBandwith(40).
