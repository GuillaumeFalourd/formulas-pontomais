@echo off
SETLOCAL

call:%~1
goto exit

:runFormula
  echo #####################################################################
  echo ##################### Banco de horas do time ########################
  echo #####################################################################
  rit pontomais relatorio banco-horas

  echo #####################################################################
  echo #################### Inconsistencia  de ponto #######################
  echo #####################################################################
  rit pontomais relatorio inconsistencia --rit_period="Ultimo mes" --rit_min_records="4"

  echo #####################################################################
  echo ##################### Solicitacoes abertas ##########################
  echo #####################################################################
  rit pontomais relatorio solicitacoes --rit_period="Ultimo mes"

  goto exit

:exit
  ENDLOCAL
  exit /b 0
