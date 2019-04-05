
docker exec -it hydra hydra clients create \
    --endpoint http://hydra:4445 \
    --id idp \
    --secret idp-secret \
    --response-types code,id_token \
    --grant-types refresh_token,authorization_code \
    --scope openid,offline \
    --callbacks https://id.hiveon.net/callback