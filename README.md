# Vaktor Lønn

Dette er en komponent som regner ut lønn for beredskapsvakt i NAV IT.
Lønnen blir beregnet basert på vaktperioden din minus arbeidet tid.

## Flyten i Vaktor

```mermaid
sequenceDiagram
actor Vakthaver
actor Vaktsjef
actor BDM
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
Plan->>Fullmaktregister: Henter BDM for vakthaver
Fullmaktregister-->>Plan: Liste over BDMer for vakthaver
Plan->>BDM: Ber om godkjenning av utbetalinger
BDM-->>Plan: Godkjenner vakthaver sin utbetalinger
Plan-->>Økonomi: Sender godkjente vaktperioder
```

## Dataflyt i Vaktor

```mermaid
flowchart LR
  subgraph NAIS
    vp(Vaktor Plan)
    vl(Vaktor Lønn)
    pgvp[("Vaktor Plan (10år lagring)")]
    pgvl[(Vaktor Lønn)]
    vp<-- "BMD (ident)" -->Fullmaktsregister
  end

  vp-- "vaktplan (ident, vaktplan)" -->vl
  vl-- "beregning (sum, timer)" -->vp
  vl<-- "Ident, vaktplan (begge slettes etter beregning)" -->pgvl

  vp-- "Vaktplan (ident), beregning (sum, timer) " -->pgvp

  subgraph Azure
    vp-- "Innlogging/SSO" -->AzureAD
  end

  subgraph on-prem
    vp-- "beregning (artskoder, sum, timer) per ident" -->ØT
    vl<-- "timelister, satser, lønn" -->Datavarehus
  end
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
