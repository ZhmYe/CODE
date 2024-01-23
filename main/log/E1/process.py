import os
from matplotlib import pyplot as plt
import numpy as np
directory = './'
def process_content(content):
    result = {}
    skews = []
    CPUNumber = 0
    concurrencys = []
    lines = content.split("\n")
    current_cpu = 0
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
        # CPU, skew
        if len(splits) == 1:
            key_value_eq = splits[0].split("=")
            key, value = key_value_eq[0], key_value_eq[1]
            if key == "CPU":
                current_cpu = int(value)
                CPUNumber = current_cpu
                result[current_cpu] = {}
            if key == "skew":
                current_skew = float(value)
                if float(current_skew) not in skews:
                    skews.append(float(current_skew))
                result[current_cpu][current_skew] = {"latency": [], "abort_rate": []}
        # concurrency, time, abort rate
        else:
            concurrency = splits[0].split("=")[-1]
            latency = splits[1].split("=")[-1]
            abort_rate = splits[2].split("=")[-1]
            if int(concurrency) not in concurrencys:
                concurrencys.append(int(concurrency))
            if latency.endswith('ms'):
                result[current_cpu][current_skew]["latency"].append(float(latency[:len(latency) - 2]))
            else:
                result[current_cpu][current_skew]["latency"].append(float(latency[:len(latency) - 1]) * 1000)
            result[current_cpu][current_skew]["abort_rate"].append(float(abort_rate))
    return {
        "data": result,
        "skews": skews,
        "cpu": CPUNumber,
        "concurrencys": concurrencys
    }
total_tps_data = []
for root, dirs, files in os.walk(directory):
    for file in files:
        if file.endswith('.txt'):
            with open(os.path.join(root, file), 'r') as f:
                content = f.read()
                result = process_content(content)
                data, skews, cpu, concurrencys = result["data"], result["skews"], result["cpu"], result["concurrencys"]
                f.close()
            index = np.array(concurrencys)
            log_index = [np.log2(item) for item in index]
            plot_data = 10000 / np.array(data[cpu][skews[0]]["latency"])* 1000
            total_tps_data.append(plot_data)
            plt.plot(log_index, plot_data, label="cpu={}".format(cpu))
            plt.xticks(log_index, index)
row_names = ['2', '4', '8', '16', '32']
col_names = ['1', '2', '4', '8', '16', '32', '64', '128', '256', '512', '1024', '2048', '4096']
# 创建DataFrame
df = pd.DataFrame(total_tps_data, index=row_names, columns=col_names)

# 将DataFrame写入Excel文件
df.to_excel('output.xlsx', sheet_name='tps')
plt.legend()
plt.savefig("./output.png")
print(total_data)