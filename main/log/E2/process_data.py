import os
from matplotlib import pyplot as plt
import numpy as np
import pandas as pd
def process_content(content):
    result = {}
    skews = []
    concurrencys = []
    lines = content.split("\n")
    blockSize = 0
    current_block_size = 0
    current_skew = 0
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
        # BlockSize, skew
        if len(splits) == 1:
            key_value_eq = splits[0].split("=")
            if len(key_value_eq) == 1:
                continue
            key, value = key_value_eq[0], key_value_eq[1]
            if key == "BlockSize":
                current_block_size = int(value)
                result[current_block_size] = {}
                blockSize = current_block_size
            if key == "skew":
                current_skew = float(value)
                if float(current_skew) not in skews:
                    skews.append(float(current_skew))
                result[current_block_size][current_skew] = {"latency": [], "abort_rate": []}
        # concurrency, time, abort rate
        else:
            latency = splits[0].split("=")[-1]
            abort_rate = splits[1].split("=")[-1]
            if latency.endswith('ms'):
                result[current_block_size][current_skew]["latency"].append(float(latency[:len(latency) - 2]))
            else:
                result[current_block_size][current_skew]["latency"].append(float(latency[:len(latency) - 1]) * 1000)
            result[current_block_size][current_skew]["abort_rate"].append(float(abort_rate))
    return {
        "data": result,
        "skews": skews,
        "blockSize": blockSize,
    }
def drawFigure(inputs, name, index, title, width=0.2):
    fig, ax = plt.subplots()
    x = np.arange(len(inputs["complete"]))
    ax.bar(x - width * 0.5, inputs["complete"], width=width, label="Complete Concurrency")
    ax.bar(x + width * 0.5, inputs["optimal"], width=width, label="Optimal Concurrency")
    ax.set_ylabel(title)
    ax.set_xticks(x, labels=index)

    ax.set_xlabel("Block Size")
    plt.legend()
    plt.savefig(name)
    plt.close()
def writeExcel(inputs, name, path):
    with pd.ExcelWriter('{}/{}.xlsx'.format(path, name)) as writer:
        for key, values in inputs.items():
            df = pd.DataFrame({'100': values['complete'][0], '200': values['complete'][1], '500': values['complete'][2],
                               '1000': values['complete'][3], '2000': values['complete'][4], '5000': values['complete'][5]},
                              index=['complete'])

            df_optimal = pd.DataFrame({'100': values['optimal'][0], '200': values['optimal'][1], '500': values['optimal'][2],
                                       '1000': values['optimal'][3], '2000': values['optimal'][4], '5000': values['optimal'][5]},
                                      index=['optimal'])

            df_combined = pd.concat([df, df_optimal])
            df_combined.to_excel(writer, sheet_name=f'{key}_skew')
def process(directory):
    abort_datas, tps_data  = {}, {}
    block_size_index = []
    for root, dirs, files in os.walk(directory):
        for file in sorted(files):
            if file.endswith('.txt'):
                with open(os.path.join(root, file), 'r') as f:
                    content = f.read()
                    result = process_content(content)
                    data, skews, blockSize = result["data"], result["skews"], result["blockSize"]
                    f.close()
                datas = data[blockSize]
                block_size_index.append(blockSize)
    #             print()
                for skew in datas:
                    each_data = datas[skew]
                    if skew not in abort_datas:
                        abort_datas[skew] = {"complete": [], "optimal": []}
                    if skew not in tps_data:
                        tps_data[skew] = {"complete": [], "optimal": []}
                    abort_datas[skew]["complete"].append(each_data["abort_rate"][0])
                    abort_datas[skew]["optimal"].append(each_data["abort_rate"][1])
                    tps_data[skew]["complete"].append(blockSize * (1 - each_data["abort_rate"][0] / 100) / each_data["latency"][0] * 1000)
                    tps_data[skew]["optimal"].append(blockSize * (1 - each_data["abort_rate"][1] / 100) / each_data["latency"][1] * 1000)

    writeExcel(abort_datas, directory[2:], directory)
    for skew in abort_datas:
        drawFigure(abort_datas[skew], "{}/abort_skew_{}.png".format(directory, skew), block_size_index, "Abort Rate(%)")
    for skew in tps_data:
        drawFigure(tps_data[skew], "{}/tps_skew_{}.png".format(directory, skew), block_size_index, "Effective Throughput(tps)")
for directory in ["./concurrency_64_fabric", "./concurrency_64_nezha", "./concurrency_128_fabric", "./concurrency_128_nezha"]:
    process(directory)