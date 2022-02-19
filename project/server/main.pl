% Program
ingraph(N, M) :- graph(N, M).
ingraph(M, N) :- graph(N, M).
{ cell(N, M, S, C) } B :- totalBandwith(B), node(N), node(M), N != M, ingraph(N, M), S = 1..nSlots, C = 1..nChannels.


% Remove models without at least a cell from the root each node.
:- graph(M, N), not cell(N, M, _, _).

% Remove models that contains a cell on the same slot and channel as another node which would make a conflicts.
:- node(N), node(M), N != M, topology(N, M), cell(N, NN, S, C), NN != M, cell(M, MM, S, C), MM != N.

% Remove models with TX cell but not a RX associated cell.
:- cell(N, M, S, C), not cell(M, N, S, C).

% Remove models that doesn't satisfy the bandwith requirements.
:- node(N), bandwith(N, B), not B { cell(N, _, _, _) }.

#show cell/4.
