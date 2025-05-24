#!/bin/bash

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))

SQDB=sq:"$(realpath $HOME/.mozilla/firefox/*.default-release/cookies.sqlite)?mode=ro"

set -x

TYPE_COMMENT='{{ . }} is a browser cookie.'
FUNC_COMMENT='{{ . }} retrieves cookies.'
FIELDS='Expiry int64,Host string,Name string,Value string,Path string,IsSecure bool,IsHTTPOnly bool'
dbtpl query "$SQDB" \
  --type Cookie \
  --type-comment="$TYPE_COMMENT" \
  --func Cookies \
  --func-comment="$FUNC_COMMENT" \
  --fields="$FIELDS" \
  --trim \
  --strip \
  --interpolate \
  --out=$SRC/models \
  --single=models.go \
<< 'ENDSQL'
/* %%host string,interpolate%% */
SELECT
  expiry,
  host,
  name,
  value,
  path,
  isSecure,
  isHttpOnly
FROM moz_cookies
ENDSQL

FUNC_COMMENT='{{ . }} retrieves cookies like the host.'
dbtpl query "$SQDB" \
  --type Cookie \
  --func CookiesLikeHost \
  --func-comment="$FUNC_COMMENT" \
  --fields="$FIELDS" \
  --trim \
  --strip \
  --append \
  --out=$SRC/models \
  --single=models.go \
<< 'ENDSQL'
SELECT
  host,
  name,
  value,
  path,
  expires_utc,
  isSecure,
  is_httponly
FROM cookies
WHERE host_key LIKE %%host string%%
ENDSQL
