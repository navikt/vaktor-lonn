# Vaktor Lønn

Dette er en komponent som regner ut lønn for beredskapsvakt.

## Antagelser og spørsmål

- Siden MinWinTid rapporterer på minutter, og vakttillegg regnes i timer, så vil Vaktor Lønn legge sammen alle minutter
  per individuelle vakttillegg i en periode, og så gjøre det om til timer. Dette samsvarer med hvordan økonomi regner
  lignende tillegg.
- Vaktor Lønn vil ikke regne vakttillegg for tid man ikke jobber i kjernetiden, da det da skal være andre på vakt, og
  beredskapsvakt er til for å dekke uforutsette hendelser utenom arbeidstid.
- Vaktor Lønn vil trekke fra tid som overstiger maks vaktperiode per dag.
- Vaktor Lønn vil ikke følge med på om man har mer enn lovlig mengde vakt i en periode, eller om man glemmer å føre
  timer.
- Vaktor Lønn vil hente beredskapstillegg, lønn, og timelister fra MinWinTid.
