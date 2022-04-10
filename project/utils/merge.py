import json
import numpy as np

def load_files(filenames):
    res = []
    for filename in filenames:
        print(f"Loading filename: {filename}...")
        with open(filename) as f:
            res.append(json.load(f))
    return res

def merge(filenames):
    stats = load_files(filenames)
    # Assume that each stats contains the same flat JSON structure.
    merged_stats = {}
    for key in stats[0]:
        if type(stats[0][key]) == list:
            merged_stats[key] = np.matrix([s[key] for s in stats])
        else:
            merged_stats[key] = np.array([s[key] for s in stats])
    return merged_stats


def timeToInstall(filenames):
    stats = load_files(filenames)
    nclients = [s["nclients"] for s in stats]
    maxClient = max(nclients)
    minClient = min(nclients)
    data = [[] for n in range(minClient, maxClient + 1)]
    startTimes = [pd.to_datetime(s["scheduleUpdateStart"]) for s in stats]
    endTimes = [pd.to_datetime(s["scheduleUpdateEnd"]) for s in stats]
    nclients = [s["nclients"] for s in stats]
    deltas = [(n, (e - s).total_seconds()) for (n, s, e) in zip(nclients, startTimes, endTimes)]
    for n, delta in deltas:
        data[n-minClient].append(delta)

    nsimulations = [len(x) for x in data]
    # print(data)
    print(nsimulations)
    ax = sns.boxplot(data=data)
    ax.set_xticklabels(list(range(minClient, maxClient+1)))
    ax.set_title("Time(s) to install a new schedule on a linear topology")
    ax.set_ylabel("Time(s)")
    ax.set_xlabel("Number of motes")
    plt.show()
    plt.close()

if __name__ == "__main__":
    from sys import argv
    import pandas as pd
    import seaborn as sns
    import matplotlib.pyplot as plt
    timeToInstall(argv[1:])
    # merged_stats = merge(argv[1:])
    # ax = sns.boxplot(data=merged_stats["tx"])
    # _, nmotes = merged_stats["tx"].shape
    # ax.set_xticklabels([str(i) for i in range(1, nmotes+1)])

    # plt.close()

    # sns.histplot(data=merged_stats["scheduleInstallationTime"], bins=6)
    # plt.show()
    # print(merged_stats["scheduleInstallationTime"])
    # print(merged_stats["scheduleInstallationTime"].mean())

