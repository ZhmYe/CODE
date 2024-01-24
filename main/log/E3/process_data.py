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
    final_result = []
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
        # concurrency
        if len(splits) == 1:
            concurrency = splits[0].split("=")[1]
        else:
            result = {}
            result["id"] = int(splits[0].split(" ")[1])
            result["concurrency"] = int(splits[1].split("=")[1])
            time_duration = splits[2].split("=")[1]
            if time_duration.endswith("ms"):
                result["time"] = float(time_duration[:len(time_duration) - 2])
            else:
                result["time"] = 1000 * float(time_duration[:len(time_duration) - 1])
            result["tx_number"] = int(splits[3].split("=")[1])
            # result["abort_rate"] = float(splits[4].split("=")[1])
            final_result.append(result)
    return {
        "data": final_result,
        "concurrency": concurrency
    }
def drawFigure(inputs, width=0.2):
    tx_numbers = []
    times = []
    tps = []
    ids = []
    x = np.arange(len(inputs))
    for data in inputs:
        tx_numbers.append(data["tx_number"])
        times.append(data["time"])
        tps.append(data["tx_number"] / data["time"] * 1000)
        ids.append("Instance{}".format(data["id"]))
    plt.clf()
    plt.bar(ids, tx_numbers, width=width)
    # plt.set_xticks(x, ids)
    plt.savefig("./tx_number.png")
    plt.clf()
    plt.bar(ids, times, width=width)
    # plt.set_xticks(x, ids)
    plt.savefig("./times.png")
    plt.clf()
    plt.bar(ids, tps, width=width)
    # plt.set_xticks(x, ids)
    plt.savefig("./tps.png")
    df_data = pd.DataFrame({
            "ID": ids,
            "Tx Numbers": tx_numbers,
            "Times": times,
            "TPS": tps
        })

    # Save the DataFrame to the Excel file with different sheets
    with pd.ExcelWriter("E3.xlsx") as writer:
        df_data.to_excel(writer, sheet_name="data", index=False)
def process(directory):
    for root, dirs, files in os.walk(directory):
        for file in sorted(files):
            if file.endswith('.txt'):
                with open(os.path.join(root, file), 'r') as f:
                    content = f.read()
                    result = process_content(content)
                    concurrency, data = result["concurrency"], result["data"]
                    drawFigure(data)
                    f.close()

process("./")