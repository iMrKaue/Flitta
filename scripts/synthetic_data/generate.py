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


def choose_customer(customers):
    return random.choice(customers)


def choose_service(customer):
    if random.random() < 0.70:
        return customer["preferred_service"]

    return weighted_choice(SERVICES)


def choose_status(customer):
    weights = {
        "completed": STATUS_WEIGHTS["completed"],
        "cancelled": (
            STATUS_WEIGHTS["cancelled"]
            * customer["cancel_multiplier"]
        ),
        "no_show": (
            STATUS_WEIGHTS["no_show"]
            * customer["no_show_multiplier"]
        ),
    }

    statuses = list(weights.keys())
    status_weights = list(weights.values())

    return random.choices(
        statuses,
        weights=status_weights,
        k=1,
    )[0]


def service_slots(service_name):
    duration = SERVICES[service_name]["duration"]

    return duration // 60


def available_start_hours(occupied_hours, service_name):
    required_slots = service_slots(service_name)

    available = []

    for hour in range(OPENING_TIME.hour, CLOSING_TIME.hour):
        service_hours = [
            hour + offset
            for offset in range(required_slots)
        ]

        if service_hours[-1] >= CLOSING_TIME.hour:
            continue

        if any(service_hour in occupied_hours for service_hour in service_hours):
            continue

        available.append(hour)

    return available


def choose_start_hour(available_hours):
    if not available_hours:
        return None

    weights = [
        HOUR_WEIGHTS[hour]
        for hour in available_hours
    ]

    return random.choices(
        available_hours,
        weights=weights,
        k=1,
    )[0]


def generate_appointments(customers, business_days):
    appointments = []

    appointment_number = 1

    for business_day in business_days:
        occupied_hours = set()

        booked_customers = set()

        target = business_day["target_appointments"]

        attempts = 0
        max_attempts = target * 20

        while (
            len([
                appointment
                for appointment in appointments
                if appointment["date"] == business_day["date"]
            ]) < target
            and attempts < max_attempts
        ):
            attempts += 1

            customer = choose_customer(customers)

            if customer["phone"] in booked_customers:
                continue

            service_name = choose_service(customer)

            available_hours = available_start_hours(
                occupied_hours,
                service_name,
            )

            start_hour = choose_start_hour(available_hours)

            if start_hour is None:
                continue

            required_slots = service_slots(service_name)

            for offset in range(required_slots):
                occupied_hours.add(start_hour + offset)

            service = SERVICES[service_name]

            status = choose_status(customer)

            created_at = generate_created_at(
                business_day["date"],
                start_hour,
            )

            appointment_datetime = datetime.combine(
                business_day["date"],
                time(start_hour, 0),
            )

            cancelled_at = generate_cancelled_at(
                created_at,
                appointment_datetime,
                status,
            )

            reminder_sent, reminder_sent_at = generate_reminder(
                business_day["date"],
                start_hour,
                created_at,
                cancelled_at,
            )

            if status == "completed":
                updated_at = appointment_datetime + timedelta(
                    minutes=service["duration"]
                )
            elif status == "no_show":
                updated_at = appointment_datetime + timedelta(minutes=15)
            else:
                updated_at = cancelled_at

            appointments.append(
                {
                    "synthetic_id": appointment_number,
                    "client_id": CLIENT_ID,
                    "customer_name": customer["name"],
                    "customer_phone": customer["phone"],
                    "customer_profile": customer["profile"],
                    "service": service_name,
                    "date": business_day["date"],
                    "time": f"{start_hour:02d}:00",
                    "price_snapshot": service["price"],
                    "duration_snapshot": service["duration"],
                    "status": status,
                    "reminder_sent": reminder_sent,
                    "reminder_sent_at": reminder_sent_at,
                    "created_at": created_at,
                    "updated_at": updated_at,
                    "cancelled_at": cancelled_at,
                }
            )

            appointment_number += 1
            booked_customers.add(customer["phone"])

    appointments.sort(
        key=lambda appointment: (
            appointment["date"],
            appointment["time"],
        )
    )

    return appointments


def generate_created_at(appointment_date, start_hour):
    appointment_datetime = datetime.combine(
        appointment_date,
        time(start_hour, 0),
    )

    lead_days = random.choices(
        [0, 1, 2, 3, 5, 7, 10, 14, 21, 30],
        weights=[4, 8, 10, 12, 14, 14, 12, 10, 8, 4],
        k=1,
    )[0]

    lead_hours = random.randint(1, 8)

    created_at = (
        appointment_datetime
        - timedelta(days=lead_days)
        - timedelta(hours=lead_hours)
    )

    return created_at


def generate_cancelled_at(created_at, appointment_datetime, status):
    if status != "cancelled":
        return None

    available_seconds = int(
        (appointment_datetime - created_at).total_seconds()
    )

    if available_seconds <= 3600:
        return created_at + (
            appointment_datetime - created_at
        ) / 2

    minimum_offset = 30 * 60
    maximum_offset = available_seconds - (60 * 60)

    cancel_offset = random.randint(
        maximum_offset,
        maximum_offset,
    )

    return created_at + timedelta(seconds=cancel_offset)


def generate_reminder(
    appointment_date,
    start_hour,
    created_at,
    cancelled_at,
):
    appointment_datetime = datetime.combine(
        appointment_date,
        time(start_hour, 0),
    )

    hours_between = (
        appointment_datetime - created_at
    ).total_seconds() / 3600

    if hours_between < 24:
        reminder_probability = 0.55
    else:
        reminder_probability = 0.88

    reminder_sent = random.random() < reminder_probability

    if not reminder_sent:
        return False, None

    reminder_sent_at = appointment_datetime - timedelta(hours=24)

    if reminder_sent_at < created_at:
        reminder_sent_at = created_at + timedelta(minutes=30)

    if (
        cancelled_at is not None
        and reminder_sent_at >= cancelled_at
    ):
        return False, None

    return True, reminder_sent_at


def format_datetime(value):
    if value is None:
        return ""

    return value.strftime("%Y-%m-%d %H:%M:%S")


def write_appointments_csv(appointments):
    fieldnames = [
        "client_id",
        "name",
        "service",
        "date",
        "time",
        "customer_phone",
        "reminder_sent",
        "reminder_sent_at",
        "status",
        "created_at",
        "updated_at",
        "cancelled_at",
        "price_snapshot",
        "duration_snapshot",
    ]

    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    with OUTPUT_FILE.open(
        "w",
        newline="",
        encoding="utf-8",
    ) as csv_file:
        writer = csv.DictWriter(
            csv_file,
            fieldnames=fieldnames,
        )

        writer.writeheader()

        for appointment in appointments:
            writer.writerow(
                {
                    "client_id": appointment["client_id"],
                    "name": appointment["customer_name"],
                    "service": appointment["service"],
                    "date": appointment["date"].isoformat(),
                    "time": appointment["time"],
                    "customer_phone": appointment["customer_phone"],
                    "reminder_sent": str(
                        appointment["reminder_sent"]
                    ).lower(),
                    "reminder_sent_at": format_datetime(
                        appointment["reminder_sent_at"]
                    ),
                    "status": appointment["status"],
                    "created_at": format_datetime(
                        appointment["created_at"]
                    ),
                    "updated_at": format_datetime(
                        appointment["updated_at"]
                    ),
                    "cancelled_at": format_datetime(
                        appointment["cancelled_at"]
                    ),
                    "price_snapshot": (
                        f"{appointment['price_snapshot']:.2f}"
                    ),
                    "duration_snapshot": (
                        appointment["duration_snapshot"]
                    ),
                }
            )


def main():
    random.seed(SEED)

    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    customers = generate_customers()

    business_days = generate_business_days()

    appointments = generate_appointments(customers, business_days)

    write_appointments_csv(appointments)

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

    status_counts = {}

    for appointment in appointments:
        status = appointment["status"]
        status_counts[status] = status_counts.get(status, 0) + 1

    print()
    print("Distribuição de status:")

    for status, count in sorted(status_counts.items()):
        percentage = count / len(appointments)

        print(
            f"{status}: "
            f"{count} |"
            f"{percentage:.1%}"
        )

    profile_outcomes = {}

    for appointment in appointments:
        profile = appointment["customer_profile"]
        status = appointment["status"]

        if profile not in profile_outcomes:
            profile_outcomes[profile] = {
                "total": 0,
                "no_show": 0,
                "cancelled": 0,
            }

        profile_outcomes[profile]["total"] += 1

        if status == "no_show":
            profile_outcomes[profile]["no_show"] += 1

        if status == "cancelled":
            profile_outcomes[profile]["cancelled"] += 1

    print()
    print("Comportamento por perfil:")

    for profile, values in sorted(profile_outcomes.items()):
        no_show_rate = values["no_show"] / values["total"]
        cancel_rate = values["cancelled"] / values["total"]

        print(
            f"{profile}: "
            f"{values['total']} agendamentos | "
            f"no_show {no_show_rate:.1%} | "
            f"cancelamento {cancel_rate:.1%}"
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

    print()
    print(f"Agendamentos efetivamente encaixados: {len(appointments)}")
    print(
        "Agendamentos não encaixados por falta de espaço: "
        f"{planned_appointments - len(appointments)}"
    )

    service_counts = {}

    for appointment in appointments:
        service = appointment["service"]
        service_counts[service] = service_counts.get(service, 0) + 1

    print()
    print("Serviços efetivamente agendados:")

    for service, count in sorted(service_counts.items()):
        percentage = count / len(appointments)

        print(
            f"{service}: "
            f"{count} | "
            f"{percentage:.1%}"
        )

    print()
    print("Primeiros 10 agendamentos:")

    for appointment in appointments[:10]:
        print(
            f"{appointment['date']} "
            f"{appointment['time']} | "
            f"{appointment['customer_phone']} | "
            f"{appointment['customer_name']} | "
            f"{appointment['service']} | "
            f"{appointment['status']}"
        )

    print()
    print(f"CSV gerado: {OUTPUT_FILE}")
    print(f"Linhas de dados: {len(appointments)}")


if __name__ == "__main__":
    main()
