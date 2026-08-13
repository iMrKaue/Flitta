import csv
import random
from datetime import date, datetime, time, timedelta
from pathlib import Path

SEED = 42

CLIENT_ID = 19

START_DATE = date(2026, 2, 1)
END_DATE = date(2026, 8, 12)

OPENING_TIME = time(9, 0)
CLOSING_TIME = time(18, 0)

OUTPUT_DIR = Path(__file__).parent / "output"
OUTPUT_FILE = OUTPUT_DIR / "synthetic_appointments.csv"

SERVICES = {
    "Corte": {
        "duration": 60,
        "price": 55.00,
        "weight": 0.50,
    },

    "Escova": {
        "duration": 60,
        "price": 70.00,
        "weight": 0.35,
    },

    "Progressiva": {
        "duration": 180,
        "price": 250.00,
        "weight": 0.15,
    },
}

STATUS_WEIGHTS = {
    "completed": 0.82,
    "cancelled": 0.11,
    "no_show": 0.07,
}


def main():
    random.seed(SEED)

    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    print("Flitta synthetic data generator")
    print(f"Seed: {SEED}")
    print(f"Client ID: {CLIENT_ID}")
    print(f"Período: {START_DATE} até {END_DATE}")
    print(f"Arquivo de saída: {OUTPUT_FILE}")
    print()

    for service_name, config in SERVICES.items():
        print(
            f"{service_name}: "
            f"{config['duration']} min | "
            f"R$ {config['price']:.2f} | "
            f"peso {config['weight']:.0%}"
        )


if __name__ == "__main__":
    main()
