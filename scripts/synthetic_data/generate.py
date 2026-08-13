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

CUSTOMER_COUNT = 120

FIRST_NAMES = [
    "Ana",
    "Beatriz",
    "Camila",
    "Carla",
    "Daniela",
    "Fernanda",
    "Gabriela",
    "Isabela",
    "Juliana",
    "Larissa",
    "Mariana",
    "Natalia",
    "Patricia",
    "Renata",
    "Vanessa",
]

LAST_NAMES = [
    "Almeida",
    "Barbosa",
    "Cardoso",
    "Costa",
    "Ferreira",
    "Gomes",
    "Lima",
    "Martins",
    "Oliveira",
    "Pereira",
    "Ribeiro",
    "Rocha",
    "Santos",
    "Silva",
    "Souza",
]

CUSTOMER_PROFILES = {
    "reliable": {
        "weight": 0.65,
        "no_show_multiplier": 0.45,
        "cancel_multiplier": 0.75,
    },
    "regular": {
        "weight": 0.25,
        "no_show_multiplier": 1.00,
        "cancel_multiplier": 1.00,
    },
    "risky": {
        "weight": 0.10,
        "no_show_multiplier": 2.50,
        "cancel_multiplier": 1.40,
    },
}

WEEKDAY_DEMAND = {
    0: (3, 4), # Segunda
    1: (3, 5), # Terça
    2: (4, 5), # Quarta
    3: (4, 6), # Quinta
    4: (5, 7), # Sexta
    5: (5, 7), # Sábado
    6: (0, 0), # Domingo fechado
}

HOUR_WEIGHTS = {
    9: 0.70,
    10: 1.00,
    11: 1.10,
    12: 0.65,
    13: 0.75,
    14: 0.90,
    15: 1.10,
    16: 1.25,
    17: 1.20,
}


def weighted_choice(options):
    names = list(options.keys())
    weights = [options[name]["weight"] for name in names]

    return random.choices(
        names,
        weights=weights,
        k=1,
    )[0]


def generate_customers():
    customers = []

    for index in range(1, CUSTOMER_COUNT + 1):
        first_name = random.choice(FIRST_NAMES)
        last_name = random.choice(LAST_NAMES)

        profile_name = weighted_choice(CUSTOMER_PROFILES)
        profile = CUSTOMER_PROFILES[profile_name]

        preferred_service = weighted_choice(SERVICES)

        customers.append(
            {
                "id": index,
                "name": f"{first_name} {last_name}",
                "phone": f"55119000{index:05d}",
                "profile": profile_name,
                "preferred_service": preferred_service,
                "no_show_multiplier": profile["no_show_multiplier"],
                "cancel_multiplier": profile["cancel_multiplier"],
            }
        )

    return customers


def generate_business_days():
    business_days = []

    current_date = START_DATE

    while current_date <= END_DATE:
        min_demand, max_demand = WEEKDAY_DEMAND[current_date.weekday()]

        if max_demand > 0:
            target_appointments = random.randint(
                min_demand,
                max_demand,
            )

            business_days.append(
                {
                    "date": current_date,
                    "weekday": current_date.weekday(),
                    "target_appointments": target_appointments,
                }
            )

        current_date += timedelta(days=1)

    return business_days


def main():
    random.seed(SEED)

    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    customers = generate_customers()

    business_days = generate_business_days()

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

    print()
    print(f"Clientes sintéticos: {len(customers)}")

    profile_counts = {}

    for customer in customers:
        profile = customer["profile"]
        profile_counts[profile] = profile_counts.get(profile, 0) + 1

    for profile, count in sorted(profile_counts.items()):
        print(f"{profile}: {count}")

    print()
    print("Primeiros clientes gerados:")

    for customer in customers[:5]:
        print(
            f"{customer['phone']} | "
            f"{customer['name']} | "
            f"{customer['profile']} | "
            f"preferência: {customer['preferred_service']}"
        )

    print()
    print(f"Dias úteis simulados: {len(business_days)}")

    planned_appointments = sum(
        day["target_appointments"]
        for day in business_days
    )

    print(
        "Agendamentos planejados antes de considerar duração: "
        f"{planned_appointments}"
    )

    weekday_counts = {}

    for day in business_days:
        weekday = day["weekday"]

        if weekday not in weekday_counts:
            weekday_counts[weekday] = {
                "days": 0,
                "appointments": 0,
        }

        weekday_counts[weekday]["days"] += 1
        weekday_counts[weekday]["appointments"] += day["target_appointments"]

    print()
    print("Demanda planejada por dia da semana:")

    weekday_names = {
        0: "Segunda",
        1: "Terça",
        2: "Quarta",
        3: "Quinta",
        4: "Sexta",
        5: "Sábado",
    }

    for weekday, values in sorted(weekday_counts.items()):
        average = values["appointments"] / values["days"]

        print(
            f"{weekday_names[weekday]}: "
            f"{values['days']} dias | "
            f"{values['appointments']} atendimentos | "
            f"média {average:.2f}/dia"
        )


if __name__ == "__main__":
    main()
