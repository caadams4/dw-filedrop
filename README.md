# dw-filedrop

Validate a payroll CSV against Drovewell schema v5, encrypt it with PGP and
drop it on sftp.drovewell.com.

    dw-filedrop validate [-schema FILE] FILE.csv
    dw-filedrop encrypt  [-recipient R] FILE.csv
    dw-filedrop send     [-schema FILE] FILE.csv
    dw-filedrop version

Flags go before the file name (Go flag package, not my fault).

## Flags

| flag | command | default |
|---|---|---|
| -schema | validate, send | schema/payroll_v5.csv |
| -recipient | encrypt | $DW_PGP_RECIPIENT |

## Config

Env only, see .env.example. send needs DW_SFTP_HOST, DW_SFTP_USER,
DW_SFTP_KEY_FILE and DW_PGP_RECIPIENT. DW_SFTP_PORT defaults to 22,
DW_SFTP_DIR to /incoming.

Needs `gpg` and `sftp` on PATH.

## Exit codes

| code | meaning |
|---|---|
| 0 | ok |
| 2 | bad usage / file not found |
| 3 | file failed validation |
| 4 | schema file could not be loaded |
| 5 | gpg failed |
| 6 | sftp failed |
| 7 | missing or bad DW_* env |

## Build

    go build -o bin/dw-filedrop ./cmd/dw-filedrop

Needs go 1.22.
