import os
from matplotlib import pyplot as plt
import numpy as np
import pandas as pd
def process_content(content):
    result = {}
    skews = []
    concurrencys = []
    lines = content.split("\n")
    concurrency = 0
    final_result = {}
    current_mode = ""
    current_skew = 0
    skews = []
    for line in lines:
        if line == "":
            continue
        splits = line.split("\t")
        while True:
            try:
                splits.remove("")
            except:
                break
        # print(splits)
        if len(splits) == 1:
            key_value = splits[0].split("=")
            if key_value[0] == "Concurrency":
                concurrency = key_value[1]
            if key_value[0] == "mode":
                current_mode = key_value[1]
                if current_mode not in final_result:
                    final_result[current_mode] = {}
            if key_value[0] == "skew":
                current_skew = float(key_value[1])
                if current_skew not in skews:
                    skews.append(current_skew)
        else:
            final_result[current_mode][current_skew] = {}
            for split in splits:
                split_key_value = split.split("=")

                final_result[current_mode][current_skew][split_key_value[0]] = split_key_value[1]
    return {
        "data": final_result,
        "concurrency": concurrency,
        "skews": skews
    }
def drawFigureAndWriteExcel(inputs, concurrency, skews, width=0.2):
    abort_data = {}
    tps_data = {}
    for data in inputs:
        abort_data[data] = []
        tps_data[data] = []
        for skew in skews:
            abort_data[data].append(float(inputs[data][skew]["Abort Rate"]))
            time_data = inputs[data][skew]["Time"]
            if time_data.endswith("ms"):
                time_data = float(time_data[:len(time_data) - 2])
            else:
                time_data = float(time_data[:-1]) * 1000
            tps_data[data].append(int(inputs[data][skew]["Tx Number"]) * (1 - float(inputs[data][skew]["Abort Rate"]) / 100) / time_data * 1000)
    fig, ax = plt.subplots()
    x = np.arange(len(skews))
    index = 0
    for mode in abort_data:
        print(abort_data[mode])
        if index == 0:
            ax.bar(x - width, abort_data[mode], width=width, label=mode)
        if index == 1:
            ax.bar(x, abort_data[mode], width=width, label=mode)
        if index == 2:
            ax.bar(x + width, abort_data[mode], width=width, label=mode)
        index += 1
#     ax.set_ylim(0, max(max(abort_data[mode]) for mode in abort_data))
    ax.set_xticks(x, labels=skews)
    ax.set_xlabel("Skewness")
    ax.set_ylabel("Abort Rate")
    plt.legend()
    plt.savefig("abort_{}.png".format(concurrency))
    plt.close()
    print(tps_data)
    fig, ax = plt.subplots()
    x = np.arange(len(skews))
    index = 0
    for mode in tps_data:
        print(tps_data[mode])
        if index == 0:
            ax.bar(x - width, tps_data[mode], width=width, label=mode)
        if index == 1:
            ax.bar(x, tps_data[mode], width=width, label=mode)
        if index == 2:
            ax.bar(x + width, tps_data[mode], width=width, label=mode)
        index += 1
#     ax.set_ylim(0, max(max(abort_data[mode]) for mode in abort_data))
    ax.set_xticks(x, labels=skews)
    ax.set_xlabel("Skewness")
    ax.set_ylabel("Effective Tps")
    plt.legend()
    plt.savefig("tps{}.png".format(concurrency))
    plt.close()
    with pd.ExcelWriter("E4_{}.xlsx".format(concurrency)) as writer:
        # 将 abort_data 写入 "abort" sheet
        df_abort = pd.DataFrame(abort_data, index=skews)
        df_abort.to_excel(writer, sheet_name="abort")

        # 将 tps_data 写入 "tps" sheet
        df_tps = pd.DataFrame(tps_data, index=skews)
        df_tps.to_excel(writer, sheet_name="tps")
def process(directory):
    for root, dirs, files in os.walk(directory):
        for file in sorted(files):
            if file.endswith('.txt'):
                with open(os.path.join(root, file), 'r') as f:
                    content = f.read()
                    result = process_content(content)
                    concurrency, data, skews = result["concurrency"], result["data"], result["skews"]
                    drawFigureAndWriteExcel(data, concurrency, skews)
                    f.close()

process("./")