import pandas as pd
import matplotlib.pyplot as plt

import os

# outputFolder = "output"
outputFolder = "./cdsanalysis/output"

def plot_credit_termstructure(filename: str, data: pd.DataFrame):
    """
    This function plots the data in the DataFrame.
    The data is indexed by date in lines and tenors in columns.

    It plots the timeseries for each tenor in the same plot.
    """
    fig, ax = plt.subplots(figsize=(9, 4.5))
    plt.xlabel("Date")
    plt.ylabel("Credit spread")

    for tenor in data.columns:
        plt.plot(df.index, df[tenor], label=tenor)

    plt.title(filename.strip(".csv"), fontsize=15)
    plt.xticks(rotation=30, fontsize=10)

    minDate = df.index.min() - pd.Timedelta(days=1)
    maxDate = df.index.max() + pd.Timedelta(days=1)
    plt.xlim(minDate, maxDate)
    plt.legend(loc="upper left")

    plt.tight_layout()
    plt.savefig('./cdsanalysis/plots/'+filename+'.png',format="png")

if __name__ == "__main__":
    # Look for the list of csv files in the output folder.
    for filename in os.listdir(outputFolder):
        if filename.endswith(".csv"):
            print(filename)

        # Read the datas for the various csv files.
        df = pd.read_csv(outputFolder + '/' + filename)
        df["date"] = pd.to_datetime(df["date"])
        df.set_index("date", inplace=True)
        print(df)

        plot_credit_termstructure(filename.strip(".csv"), df)

    