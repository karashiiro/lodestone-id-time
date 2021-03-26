from numpy import sin
import pandas as pd
from scipy.optimize import curve_fit

def objective(x, m, b1, a, b2):
    return b1 + m * x + a * sin(b2 * x)

def main():
    # characters_cvt.csv just has the created_at column converted into float form
    df = pd.read_csv("characters_cvt.csv")
    popt, _ = curve_fit(
        objective,
        df["id"].to_numpy(),
        df["created_at"].to_numpy(),
        p0=[9.3, 41217, 40, 1 / 1200000]
    )
    print(popt)

if __name__ == "__main__":
    main()