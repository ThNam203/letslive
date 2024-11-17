#openssl req -newkey rsa:4096 \
#            -x509 \
#            -sha256 \
#            -days 3650 \
#            -nodes \
#            -out server.crt \
#           -keyout server.key \
#            -subj "/C=SI/ST=Ljubljana/L=Ljubljana/O=Security/OU=IT Department/CN=www.example.com"

openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout localhost.key -out localhost.crt -config localhost.cnf
