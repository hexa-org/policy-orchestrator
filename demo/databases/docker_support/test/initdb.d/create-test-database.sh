#!/bin/bash
set -e
psql -v on_error_stop=1 --username postgresql <<-EOSQL
  create database orchestrator_test;
  create user orchestrator with password 'orchestrator';
  grant all privileges on database orchestrator_test to orchestrator;
EOSQL