import json
import random

from fastapi import FastAPI
from datetime import datetime, timedelta

app = FastAPI()


@app.get("/")
def read_root():
    return {"Hello": "World"}


def generate_dager(fra_dato, til_dato):
    fra = datetime.strptime(fra_dato, '%Y-%m-%d')
    til = datetime.strptime(til_dato, '%Y-%m-%d')

    timesheet = []
    for date in (fra + timedelta(n) for n in range((til - fra).days + 1)):
        virkedag = "Virkedag"
        if date.weekday() == 5:
            virkedag = "Lørdag"
        elif date.weekday() == 6:
            virkedag = "Søndag"

        sheet = {
            "dato": date.isoformat(),
            "skjema_tid": 7,
            "skjema_navn": "Heltid 0800-1500 (2018)",
            "godkjent": 5,
            "ansatt_dato_godkjent_av": "m654321",
            "godkjent_dato": (til + timedelta(days=10)).isoformat(),
            "virkedag": virkedag,
            "Stemplinger": [
                {
                    "Stempling_Tid": (datetime(date.year, date.month, date.day, random.randint(7, 9),
                                               random.randint(0, 59))).isoformat(),
                    "Navn": "Inn",
                    "Type": "B1",
                    "Fravar_kode": 0,
                    "Fravar_kode_navn": "Ute"
                },
                {
                    "Stempling_Tid": (datetime(date.year, date.month, date.day, random.randint(14, 17),
                                               random.randint(0, 59))).isoformat(),
                    "Navn": "Ut",
                    "Type": "B2",
                    "Fravar_kode": 0,
                    "Fravar_kode_navn": "Ute"
                },
            ],
            "Stillinger": [
                {
                    "koststed": "855210",
                    "formal": "000000",
                    "aktivitet": "000000",
                    "RATE_K001": random.randint(400_000, 800_000),
                    "RATE_K170": 35,
                    "RATE_K171": 10,
                    "RATE_K172": 20,
                    "RATE_K160": 15,
                    "RATE_K161": 55,
                }
            ]
        }
        timesheet.append(sheet)

    return json.dumps(timesheet, ensure_ascii=False)


@app.get("/json/Hr/Vaktor/Vaktor_Tiddata")
def mock(ident: str, fra_dato: str, til_dato: str):
    return {
        "Vaktor.Vaktor_TiddataResponse": {
            "Vaktor.Vaktor_TiddataResult": [
                {
                    "Vaktor.nav_id": "123456",
                    "Vaktor.resource_id": ident,
                    "Vaktor.leder_resource_id": "654321",
                    "Vaktor.leder_nav_id": "M654321",
                    "Vaktor.leder_navn": "Kalpana, Bran",
                    "Vaktor.leder_epost": "Bran.Kalpana@nav.no",
                    "Vaktor.dager": generate_dager(fra_dato, til_dato)
                }
            ]
        }
    }
