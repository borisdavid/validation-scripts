import pandas as pd
import matplotlib.pyplot as plt
from matplotlib.dates import DateFormatter
import json

import os

outputFolder = "./backtestingplot/output"

def plot_backtesting_performance(filename: str,data: pd.DataFrame):
    """
    This function plots the data in the DataFrame.
    All data will be available in the dataframe directly.
    
    Four plots will be produced : 
    - on top-left, a plot of the monthly timeseries of prices
    - on top-right, a plot of the monthly timeseries of rolling performance at 1 year.
    - on bottom-left, a plot of the monthly timeseries of rolling volatility at 1 year.
    - on bottom-right, a plot showing the monthly drawdown.
    """
    fig, axs = plt.subplots(2, 2, figsize=(14, 10), gridspec_kw={'height_ratios': [1, 1.2]})

    # Resample to monthly if not already
    monthly_data = data.resample('M').last()

    fig, axs = plt.subplots(2, 2, figsize=(16, 10), sharex=True)
    date_format = DateFormatter("%Y-%m")

    # Top-left: prices
    axs[0, 0].plot(monthly_data.index, monthly_data['0'], color='blue')
    axs[0, 0].set_title("Monthly Prices")
    axs[0, 0].set_ylabel("Price")
    axs[0, 0].grid(True)

    # Top-right: 1Y rolling performance
    axs[0, 1].plot(monthly_data.index, monthly_data['1'], color='green')
    axs[0, 1].set_title("1Y Rolling Performance (Monthly)")
    axs[0, 1].set_ylabel("Performance")
    axs[0, 1].grid(True)

    # Bottom-left: 1Y rolling volatility
    axs[1, 0].plot(monthly_data.index, monthly_data['2'], color='orange')
    axs[1, 0].set_title("1Y Rolling Volatility (Monthly)")
    axs[1, 0].set_ylabel("Volatility")
    axs[1, 0].grid(True)

    # Bottom-right: drawdown
    axs[1, 1].plot(monthly_data.index, monthly_data['3'], color='red')
    axs[1, 1].fill_between(monthly_data.index, monthly_data['3'], 0, color='red', alpha=0.3)
    axs[1, 1].set_title("Drawdown (Monthly)")
    axs[1, 1].set_ylabel("Drawdown")
    axs[1, 1].grid(True)

    for ax in axs.flat:
        ax.xaxis.set_major_formatter(date_format)
        for label in ax.get_xticklabels():
            label.set_rotation(45)

    plt.tight_layout()
    plt.savefig(filename)
    plt.close()

def load_output_dataframe(filename: str):
    """
    Load the output dataframe from the given filename.
    The file is expected to be in JSON format.
    """
    # Open and read the JSON file
    with open(filename, 'r') as file:
        data = json.load(file)  # this converts the JSON into a Python dictionary

    dfs = []

    # usableData = data['positions']['0']['metrics']
    usableData = data['metrics']
    for metric_key in usableData:
        metric_data = usableData[metric_key]['values']
        df = pd.DataFrame(metric_data)

        df["timestamp"] = pd.to_datetime(df["timestamp"])
        df.rename(columns={"timestamp": "date"}, inplace=True)

        df.set_index("date", inplace=True)

        df.rename(columns={"value": metric_key}, inplace=True)

        dfs.append(df)

    final_df = pd.concat(dfs, axis=1)
    print(final_df)

    return final_df
        

df = load_output_dataframe(outputFolder + '/output.json')
plot_backtesting_performance(outputFolder + '/backtesting_performance.png', df)