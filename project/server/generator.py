from dataclasses import dataclass
from random import random
import sys

@dataclass
class GeneratorConfig():
    nNodes: int
    nSlots: int
    nChannels: int
    bandwith: int

def generator(filename: str, config: GeneratorConfig):
    with open(filename, 'w') as f:
        f.write(f"#const nNodes = {config.nNodes}.\n")
        f.write(f"#const nSlots = {config.nSlots}.\n")
        f.write(f"#const nChannels = {config.nChannels}.\n")
        f.write("\n")
        f.write("node(0..nNodes).\n")
        for (n, m) in zip(range(1, config.nNodes + 1), range(0, config.nNodes + 1)):
            f.write(f"graph({n}, {m}). ")
        f.write("\n")
        for (n, m) in zip(range(1, config.nNodes + 1), range(0, config.nNodes + 1)):
            f.write(f"topology({n}, {m}). topology({m}, {n}). ")
        f.write("\n")
        f.write(f"bandwith(N, {config.bandwith}) :- node(N).\n")
        f.write(f"totalBandwith({config.bandwith * config.nNodes * 2}).\n")

if __name__ == "__main__":
    config = GeneratorConfig(nNodes = int(sys.argv[1]), nSlots = int(sys.argv[2]), nChannels = int(sys.argv[3]), bandwith = int(sys.argv[4]))
    generator("facts.pl", config)
