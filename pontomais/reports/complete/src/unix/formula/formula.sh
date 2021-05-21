#!/bin/sh

runFormula() {
  echo "#####################################################################"
  echo "##################### Banco de horas do time ########################"
  echo "#####################################################################"
  rit pontomais relatorio banco-horas

  echo "#####################################################################"
  echo "#################### Inconsistencia  de ponto #######################"
  echo "#####################################################################"
  rit pontomais relatorio inconsistencia --rit_period="Ultima semana" --rit_min_records="4"

  echo "#####################################################################"
  echo "##################### Solicitacoes abertas ########################"
  echo "#####################################################################"
  rit pontomais relatorio solicitacoes --rit_period="Ultimo mes"
}
