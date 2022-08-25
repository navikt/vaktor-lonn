# Vaktor Lønn

Dette er en POC for utregning av lønn for beredskapsvakt.

Foreløpig er det et internal repo, da jeg må undersøke om satser og utregninger er sensitiv informasjon.

## Antagelser og spørsmål

- Siden MinWinTid rapporterer på minutter, og vakttillegg regnes i timer, så vil Vaktor Lønn legge sammen alle minutter
  per individuelle vakttillegg i en periode, og så gjøre det om til timer. Dette samsvarer med 

## Spørsmål

Alle spørsmål er tenkt med sommertid, så lenge ikke annet er nevnt.

- Kan Vaktor anta at en vakthaver avspaserer fleksitid i det tidsrommet som er mest gunstig? Eller må de som gå vakt
  føre nøyaktig avspasering. Dette er ikke et krav i MinWinTid og det vanlige er at det blir automatisk ført som 
  avspasering av fleksitid.
  - Eksempel: MinWinTid rapporterer arbeidstid fra 0800-1400. Som betyr at vakthaver kun har jobbet 6t. Det betyr at
    personen avspaserer 1t. Det er mest gunstig at vakthaver avspaserer fra kl14-kl15, i stedet for kl07-kl08 da 
    sistnevnte går under skiftillegg.
  - Man kan ikke ha vakt i kjernetiden. Og hvis man bare jobber i kjernetiden, så antar Vaktor Lønn at noen TM passer 
    på systemene de timene som ikke dekkes av kjernetiden. Dette betyr også at man ikke får betalt vakt for denne
    perioden. Om sommeren er det 1,5t som ikke dekkes av kjernetiden, mens om vinteren er det snakk om 2,25t.
- Hva skal Vaktor Lønn gjøre hvis en vakthaver har vakt mer enn 17t vakt på en dag?
  - Dette vil også være et problem i romjulen, da arbeidstiden kun er 5t 45m. Da vil man oppnå en vakttid på 18t 15min.
- Skal Vaktor Lønn varsle om at en vakthaver har mer vakt i en periode enn lov?
- Hvor kan man finne aktuelle tillegg? I HTA (akademikerne) står det at 0620 perioden godtgjøres 15 pr. løpende time, og 2006 
  godtgjøres 25 pr. løpende time.
  - Vaktor Lønn får ansvar for å følge opp satser hver gang de endres. Forhåpentligvis kan Økonomi gi oss disse fortløpende.
- Beredskapsavtalen vi nå har sier at registrering av døgnkontinuerlig beredskapsvakt skal skje i 
  tidsregistreringssystemet. Spesifiserer også MinWinTid. Dette må nok endres.
  - Dette må vi ta opp med MDBA, for avtalen må endres.
