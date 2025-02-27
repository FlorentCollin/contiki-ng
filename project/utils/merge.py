import json
import numpy as np

def load_files(filenames):
    res = []
    for filename in filenames:
        # print(f"Loading filename: {filename}...")
        with open(filename) as f:
            stats = json.load(f)
            stats["filename"] = filename
            res.append(stats)
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


def timeToInstall(stats):
    nclients = [s["nclients"] for s in stats]
    maxClient = max(nclients)
    minClient = min(nclients)
    data = [[] for _ in range(minClient, maxClient + 1)]
    startTimes = [pd.to_datetime(s["scheduleUpdateStart"]) for s in stats]
    endTimes = [pd.to_datetime(s["scheduleUpdateEnd"]) for s in stats]
    nclients = [s["nclients"] for s in stats]
    deltas = [(n, (e - s).total_seconds()) for (n, s, e) in zip(nclients, startTimes, endTimes)]
    for n, delta in deltas:
        data[n-minClient].append(delta)
    data = list(filter(lambda d: len(d) !=0, data))
    nsimulations = [len(x) for x in data]
    print(nsimulations)
    ax = sns.boxplot(data=data)
    # theoritical_best_times_xs = [i-minClient for i in range(minClient, maxClient + 1)]
    # theoritical_best_times_ys = [best_time(i) * 0.210 for i in range(minClient, maxClient + 1)] # linear
    # theoritical_best_times_ys = [4 * (i - 1) * 0.210 for i in range(minClient, maxClient + 1)] # star
    # print(theoritical_best_times_ys)
    # sns.scatterplot(ax=ax, x=theoritical_best_times_xs, y=theoritical_best_times_ys)
    print(data)
    clients = list(set(s["nclients"] for s in stats))
    clients.sort()
    ax.set_xticklabels(clients)
    ax.set_ylim(0)
    ax.set_title("Temps(s) pour installer un nouvel ordonnancement")
    ax.set_ylabel("Temps(s)")
    ax.set_xlabel("Nombre de nœuds du WSN")
    plt.savefig("installation-time.png")
    plt.close()

def timeToInstallAll(stats_linear, stats_grid, stats_star):
    df = []
    for i, stats in enumerate((stats_linear, stats_grid, stats_star)):
        stats = list(filter(lambda s: s["nclients"], stats))
        data = [[] for _ in range(0, 13)]
        startTimes = [pd.to_datetime(s["scheduleUpdateStart"]) for s in stats]
        endTimes = [pd.to_datetime(s["scheduleUpdateEnd"]) for s in stats]
        if i == 0:
            topology = "linear"
        if i == 1:
            topology = "grid"
        if i == 2:
            topology = "star"

        nclients = [s["nclients"] for s in stats]
        data = [(n, (e - s).total_seconds(), topology) for (n, s, e) in zip(nclients, startTimes, endTimes)]
        df = df + data
    
    df = pd.DataFrame(df, columns=["N", "Temps d'installation", "Topologie"])
    print(df)

    ax = sns.pointplot(data=df, x="N", y="Temps d'installation", hue="Topologie")
    # ax.set_xticklabels(list(range(2, 13)))
    ax.set_ylim(0)
    ax.set_title("Temps(s) pour installer un nouvel ordonnancement")
    ax.set_ylabel("Temps(s)")
    ax.set_xlabel("Nombre de nœuds du WSN")
    plt.savefig("installation-time.png")
    plt.close()


def best_time(n: int) -> float:
    res = sum((3*(i-1) for i in range (2, n)))
    return (res + 2*n) * 2


def timeouts1(stats, nclients=9):
    stats = list(filter(lambda s: s["nclients"] == nclients, stats))
    if not stats:
        return
    data = {}
    for s in stats:
        try:
            timeouts = s["timeouts"]["IPMap"]
            for k, v in timeouts.items():
                try:
                    data[k].append(v)
                except:
                    data[k] = [v]
        except:
            pass
    keys = sorted(data.keys())
    data = [data[k] for k in keys]
    ax = sns.boxplot(data=data)
    clients = list(set(s["nclients"] for s in stats))
    clients.sort()
    ax.set_xticklabels(clients)
    ax.set_title(f"Nombre de timeouts par nœud pour une topologie linéaire de {nclients} nœuds")
    ax.set_ylabel("Nombre de timeouts")
    ax.set_xlabel("Nœud")
    plt.savefig("timeouts1.png")
    plt.close()

def timeouts2(stats):
    #Timeouts vs installation time
    data=[]
    for s in stats:
        filename = s["filename"]
        start = pd.to_datetime(s["scheduleUpdateStart"])
        end = pd.to_datetime(s["scheduleUpdateEnd"])
        installation_time = (end - start).total_seconds()

        try:
            timeouts = s["timeouts"]["IPMap"]
            total_timeout = sum(v for v in timeouts.values())
        except:
            total_timeout = 0
        data.append([filename, installation_time, total_timeout, s["nclients"]])

    df = pd.DataFrame(data, columns=["filename", "installation_time", "total_timeout", "Nombre de nœuds"])
    print(df.sort_values(by=["total_timeout", "installation_time"]).to_string())

    ax = sns.scatterplot(data=df, x="total_timeout", y="installation_time", hue="Nombre de nœuds", palette="deep")
    ax.set_xlim(0)
    ax.set_ylim(0)
    ax.set_title("Corrélation entre le nombre de timeouts et le temps d'installation")
    ax.set_xlabel("Nombre de timeouts")
    ax.set_ylabel("Temps(s) d'installation de l'ordonnancement")
    plt.savefig("timeout-installation-correlation.png")
    plt.clf()
    plt.close()


if __name__ == "__main__":
    from sys import argv
    import pandas as pd
    import seaborn as sns
    import matplotlib.pyplot as plt
    import matplotlib as mpl
    mpl.rcParams['figure.dpi'] = 200
    stats = load_files(argv[1:])
    timeToInstall(stats)
    # stats_linear = load_files(argv[1].split(" "))
    # stats_grid = load_files(argv[2].split(" "))
    # stats_star = load_files(argv[3].split(" "))
    # timeToInstallAll(stats_linear, stats_grid, stats_star)
    # timeouts1(stats,  nclients=10)
    timeouts2(stats)
    # merged_stats = merge(argv[1:])
    # ax = sns.boxplot(data=merged_stats["tx"])
    # _, nmotes = merged_stats["tx"].shape
    # ax.set_xticklabels([str(i) for i in range(1, nmotes+1)])

    # plt.close()

    # sns.histplot(data=merged_stats["scheduleInstallationTime"], bins=6)
    # plt.show()
    # print(merged_stats["scheduleInstallationTime"])
    # print(merged_stats["scheduleInstallationTime"].mean())

