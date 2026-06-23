#!/bin/bash
set -e

sed -i "s/__FUTU_LOGIN_ACCOUNT__/${FUTU_LOGIN_ACCOUNT}/g" FutuOpenD.xml
sed -i "s/__FUTU_LOGIN_PWD_MD5__/${FUTU_LOGIN_PWD_MD5}/g" FutuOpenD.xml

exec ./FutuOpenD -cfg_file=./FutuOpenD.xml
