# Vaktor Lønn

Dette er en komponent som regner ut lønn for beredskapsvakt i NAV IT.
Lønnen blir beregnet basert på vaktperioden din minus arbeidet tid.

## Flyten i Vaktor

```mermaid
sequenceDiagram
actor Vakthaver
actor Vaktsjef
actor leder as NAV IT Leder
participant Plan as Vaktor Plan
participant Lønn as Vaktor Lønn
Plan-->>Plan: Endt vaktperiode
Plan->>Vakthaver: Ber om godkjenning av periode
Vakthaver-->>Plan: Godkjenner vaktperiode
Plan->>Vaktsjef: Ber om godkjenning av vaktperiode
Vaktsjef-->>Plan: Godkjenner vaktperiode
Plan->>Lønn: Godkjent vaktperiode
Lønn-->>Plan: Periode mottatt
loop Every hour
  Lønn->>MinWinTid: Ber om arbeidstid i vaktperiode
  MinWinTid-->>Lønn: Arbeidstid
  Lønn-->>Lønn: Sjekk om arbeidstid er godkjent av personalleder
  Lønn-->>Lønn: Sjekk at det ikke er ferie i vaktperioden
  Lønn-->>Lønn: Beregner utbetaling av kronetillegg og<br/>overtidstillegg for vaktperioden
  Lønn->>Plan: Utbetaling for vaktperiode
end
Plan->>leder: Transaksjonsfil sendes for godkjenning via e-post
leder-->>Økonomi: Godkjenner, og melder i fra til Økonomi
```

## Utvikling

Det er satt opp CI/CD for automatisk utrulling av kodebasen.
I `dev` har vi lagd en mock av MinWinTid som automatisk genererer arbeidstid innenfor vaktperioden man tester mot.
Foreløpig satt til å kjøre utregning hvert 5 minutt.

### Lokalt

For å kjøre lokalt trenger man en egen Postgres database, tilgang til Azure AD, og mock av MinWinTid.

```shell
make env # krever tilgang til GCP
make db
make mock # i et eget shell
make local
```
