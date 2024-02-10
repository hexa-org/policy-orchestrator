#!/bin/bash
set -e
psql -v on_error_stop=1 --username postgresql <<-EOSQL
  create database orchestrator_development;
  create user orchestrator with password 'orchestrator';
  grant all privileges on database orchestrator_development to orchestrator;
EOSQL

cp /docker-entrypoint-initdb.d/ca-cert.pem /var/lib/postgresql/data
cp /docker-entrypoint-initdb.d/server-cert.pem /var/lib/postgresql/data
cp /docker-entrypoint-initdb.d/server-key.pem /var/lib/postgresql/data

chmod og-rwx /var/lib/postgresql/data/ca-cert.pem
chmod og-rwx /var/lib/postgresql/data/server-cert.pem
chmod og-rwx /var/lib/postgresql/data/server-key.pem

cat <<EOF >> /var/lib/postgresql/data/postgresql.conf
ssl=on
ssl_ca_file='ca-cert.pem'
ssl_cert_file='server-cert.pem'
ssl_key_file='server-key.pem'
ssl_ciphers='HIGH:MEDIUM:+3DES:!aNULL'
EOF

cat <<EOF >> /var/lib/postgresql/data/pg_hba.conf
hostssl all orchestrator 0.0.0.0/0 md5 clientcert=verify-ca
EOF
