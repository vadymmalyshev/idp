```
docker-compose up -d

./scripts/populate.sh

go mod vendor
go mod tidy

go build ./...
go run .
```

FYI
https://medium.com/12plus1/oauth2-with-ory-hydra-vapor-3-and-ios-12-ca0e61c28f5a

https://github.com/segmentio/ory-hydra/blob/master/docs/oauth2.md


Local certs generated by https://github.com/FiloSottile/mkcert
# IDP


Create IDP client

```
docker exec -it `docker ps -f name=idp_hydra -q` \
    hydra clients create \
    --endpoint http://hydra:4445 \
    --id idp \
    --secret idp-secret \
    --response-types code,id_token \
    --grant-types refresh_token,authorization_code \
    --scope openid,offline \
    --callbacks https://id.hiveon.net:3000/callback
```

Create IDP test example client

```
docker exec -it `docker ps -f name=idp_hydra -q` \
    hydra clients create \
    --endpoint http://hydra:4445 \
    --id id-portal \
    --secret id-portal-secret \
    --response-types code,id_token \
    --grant-types refresh_token,authorization_code \
    --scope openid,offline \
    --callbacks http://127.0.0.1:8000/callback
```

FYI
https://medium.com/12plus1/oauth2-with-ory-hydra-vapor-3-and-ios-12-ca0e61c28f5a

https://github.com/segmentio/ory-hydra/blob/master/docs/oauth2.md


Local certs generated by https://github.com/FiloSottile/mkcert
# IDP


Create IDP client

```
docker exec -it `docker ps -f name=idp_hydra -q` \
    hydra clients create \
    --endpoint http://hydra:4445 \
    --id idp \
    --secret idp-secret \
    --response-types code,id_token \
    --grant-types refresh_token,authorization_code \
    --scope openid,offline \
    --callbacks https://127.0.0.1:5555/callback
```

Create IDP test example client

```
docker exec -it `docker ps -f name=idp_hydra -q` \
    hydra clients create \
    --endpoint http://hydra:4445 \
    --id id-portal \
    --secret id-portal-secret \
    --response-types code,id_token \
    --grant-types refresh_token,authorization_code \
    --scope openid,offline \
    --callbacks http://127.0.0.1:8000/callback
```

FYI
https://medium.com/12plus1/oauth2-with-ory-hydra-vapor-3-and-ios-12-ca0e61c28f5a

https://github.com/segmentio/ory-hydra/blob/master/docs/oauth2.md


Local certs generated by https://github.com/FiloSottile/mkcert