#!/bin/sh

SCRIPTPATH=$(dirname "$BASH_SOURCE")
. "${SCRIPTPATH}/parse_yaml.sh"
. "${SCRIPTPATH}/manage_hosts.sh"

eval $(parse_yaml "${SCRIPTPATH}/../config/config.yaml" "config_")

add_client() {
    local client_id=$1
    local client_secret=$2
    local callback_url=$3
    local logo_url=$4

    echo "Deleting ${client_id} ..."

    docker exec -it `docker ps -f name=${config_hydra_docker} -q` \
        hydra clients delete ${client_id} \
        --endpoint ${config_hydra_admin} \
    }

    echo "Adding ${client_id} ..."

    docker exec -it `docker ps -f name=${config_hydra_docker} -q` \
        hydra clients create \
        --endpoint ${config_hydra_admin} \
        --id ${client_id} \
        --secret ${client_secret} \
        --response-types code,id_token \
        --grant-types refresh_token,authorization_code \
        --scope openid,offline \
        --callbacks ${callback_url}
}

#idp must be last item

add_client $config_admin_client_id $config_admin_client_secret $config_admin_callback
add_client $config_portal_client_id $config_portal_client_secret $config_portal_callback
add_client $config_idp_client_id $config_idp_client_secret $config_idp_callback

echo "Please enter your password if requested."

remove_host $config_admin_host
add_host $config_admin_host

remove_host $config_portal_host
add_host $config_portal_host

remove_host $config_idp_host
add_host $config_idp_host
