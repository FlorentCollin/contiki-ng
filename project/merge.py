import json
import numpy as np

def load_files(filenames):
    res = []
    for filename in filenames:
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


if __name__ == "__main__":
    from sys import argv
    import seaborn as sns
    import matplotlib.pyplot as plt
    merged_stats = merge(argv[1:])
    ax = sns.boxplot(data=merged_stats["tx"])
    _, nmotes = merged_stats["tx"].shape
    ax.set_xticklabels([str(i) for i in range(1, nmotes+1)])

    plt.close()

    sns.histplot(data=merged_stats["scheduleInstallationTime"], bins=10, kde=True)
    plt.show()
    print(merged_stats["scheduleInstallationTime"])
    print(merged_stats["scheduleInstallationTime"].mean())

