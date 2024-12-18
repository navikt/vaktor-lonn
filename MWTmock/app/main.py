import json
import random
from datetime import datetime, timedelta

from fastapi import FastAPI

app = FastAPI()


@app.get("/")
def read_root():
    return {"Hello": "World"}


def generate_dager(fra_dato, til_dato):
    fra = datetime.strptime(fra_dato, '%Y-%m-%d')
    til = datetime.strptime(til_dato, '%Y-%m-%d')
    salary = random.randint(400_000, 800_000)

    timesheet = []
    for date in (fra + timedelta(n) for n in range((til - fra).days + 1)):
        virkedag = "Virkedag"
        if date.weekday() == 5:
            virkedag = "Lørdag"
        elif date.weekday() == 6:
            virkedag = "Søndag"

        sheet = {
            "dato": date.isoformat(),
            "skjema_tid": 7.5,
            "skjema_navn": "Heltid 0800-1500 (2018)",
            "godkjent": 5,
            "virkedag": virkedag,
            "stemplinger": [
                {
                    "stempling_tid": (datetime(date.year, date.month, date.day, random.randint(7, 9),
                                               random.randint(0, 59))).isoformat(),
                    "navn": "Inn",
                    "type": "B1",
                    "fravar_kode": 0,
                    "fravar_kode_navn": "Ute",
                    "overtid_begrunnelse": None
                },
                {
                    "stempling_tid": (datetime(date.year, date.month, date.day, random.randint(14, 17),
                                               random.randint(0, 59))).isoformat(),
                    "navn": "Ut",
                    "type": "B2",
                    "fravar_kode": 0,
                    "fravar_kode_navn": "Ute",
                    "overtid_begrunnelse": None
                },
            ],
            "stillinger": [
                {
                    "post_id": "258",
                    "parttime_pct": 100,
                    "koststed": "855210",
                    "produkt": "000000",
                    "oppgave": "000000",
                    "rate_k001": salary,
                }
            ]
        }
        timesheet.append(sheet)

    return json.dumps(timesheet, ensure_ascii=False)


@app.get("/ords/dvh/dt_hr/vaktor/tiddata")
def mock(ident: str, fra_dato: str, til_dato: str):
    return {
        "nav_id": "123456",
        "resource_id": ident,
        "leder_resource_id": "654321",
        "leder_nav_id": "M654321",
        "leder_navn": "Kalpana, Bran",
        "leder_epost": "Bran.Kalpana@nav.no",
        "dager": generate_dager(fra_dato, til_dato)
    }


@app.post("/ords/dvh/oauth/token")
def token():
    return {
        "access_token": "super-secret-not-fake-token",
        "token_type": "bearer",
        "expires_in": 1
    }
