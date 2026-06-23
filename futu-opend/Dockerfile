FROM ubuntu:18.04

RUN apt-get update && apt-get install -y \
    wget \
    tar \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /opt/futu

COPY dist/Futu_OpenD_10.7.6728_Ubuntu18.04.tar.gz .

RUN tar -xzf Futu_OpenD_10.7.6728_Ubuntu18.04.tar.gz && \
    rm Futu_OpenD_10.7.6728_Ubuntu18.04.tar.gz && \
    mv Futu_OpenD_10.7.6728_Ubuntu18.04/Futu_OpenD_10.7.6728_Ubuntu18.04 FutuOpenD && \
    rm -rf Futu_OpenD_10.7.6728_Ubuntu18.04

RUN chmod +x FutuOpenD/FutuOpenD FutuOpenD/FTUpdate FutuOpenD/FTWebSocket

ENV FUTU_LOGIN_ACCOUNT=""
ENV FUTU_LOGIN_PWD_MD5=""
ENV FUTU_OPEND_IP="0.0.0.0"
ENV FUTU_OPEND_PORT="11111"

EXPOSE 11111

WORKDIR /opt/futu/FutuOpenD

COPY <<'EOF' /opt/futu/FutuOpenD/FutuOpenD.xml
<futu_opend>
	<ip>127.0.0.1</ip>
	<api_port>11111</api_port>
	<login_account>__FUTU_LOGIN_ACCOUNT__</login_account>
	<login_pwd_md5>__FUTU_LOGIN_PWD_MD5__</login_pwd_md5>
	<lang>chs</lang>
	<log_level>info</log_level>
	<push_proto_type>0</push_proto_type>
	<price_reminder_push>1</price_reminder_push>
	<auto_hold_quote_right>1</auto_hold_quote_right>
	<future_trade_api_time_zone>UTC+8</future_trade_api_time_zone>
	<pdt_protection>1</pdt_protection>
	<dtcall_confirmation>1</dtcall_confirmation>
</futu_opend>
EOF

COPY <<'EOF' /opt/futu/start.sh
#!/bin/bash
set -e

sed -i "s/__FUTU_LOGIN_ACCOUNT__/${FUTU_LOGIN_ACCOUNT}/g" FutuOpenD.xml
sed -i "s/__FUTU_LOGIN_PWD_MD5__/${FUTU_LOGIN_PWD_MD5}/g" FutuOpenD.xml

exec ./FutuOpenD -cfg_file=./FutuOpenD.xml
EOF

RUN chmod +x /opt/futu/start.sh

CMD ["/opt/futu/start.sh"]
