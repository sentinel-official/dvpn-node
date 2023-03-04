#!/usr/bin/env bash

set -Eeou pipefail

CONTAINER_NAME=sentinelnode
NODE_DIR="${HOME}/.sentinelnode"
NODE_IMAGE=ghcr.io/sentinel-official/dvpn-node:latest

function stop {
  id=$(docker ps --filter name="${CONTAINER_NAME}" --quiet)
  if [[ -n "${id}" ]]; then
    docker stop "${id}"
  fi
}

function remove {
  id=$(docker ps --all --filter name="${CONTAINER_NAME}" --quiet)
  if [[ -n "${id}" ]]; then
    docker rm --force --volumes "${id}"
  fi
}

function cmd_attach {
  id=$(docker ps --filter name="${CONTAINER_NAME}" --quiet)
  if [[ -n "${id}" ]]; then
    docker attach "${id}"
  else
    echo "Error: node is not running" && return 1
  fi
}

function cmd_help {
  echo "Usage: ${0} [COMMAND]"
  echo ""
  echo "Commands:"
  echo "  attach     Attach to the already running node"
  echo "  help       Print the help message"
  echo "  init       Initialize the configuration"
  echo "  setup      Install the dependencies and set up the requirements"
  echo "  start      Start the node"
  echo "  status     Display the node's status along with the last 20 lines of the log"
  echo "  stop       Stop the node"
  echo "  remove     Remove the node container"
  echo "  restart    Restart the node"
  echo "  update     Update the node to the latest version"
}

function cmd_init {
  NODE_TYPE=wireguard
  mapfile -t PORTS < <(shuf -i 1024-65535 -n 2)

  function run {
    docker run \
      --interactive \
      --rm \
      --tty \
      --volume "${NODE_DIR}:/root/.sentinelnode" \
      "${NODE_IMAGE}" process "${@}"
  }

  function must_run {
    output=$(run "${@}")
    if [[ -n "${output}" ]]; then
      echo "${output}"
    fi
    if [[ "${output}" == *"Error"* ]]; then
      return 1
    fi
  }

  function config_set {
    echo "Setting the configuration key=${1}, value=${2}"
    must_run config set "${1}" "${2}"
  }

  function v2ray_config_set {
    echo "Setting the V2Ray configuration key=${1}, value=${2}"
    must_run v2ray config set "${1}" "${2}"
  }

  function wireguard_config_set {
    echo "Setting the WireGuard configuration key=${1}, value=${2}"
    must_run wireguard config set "${1}" "${2}"
  }

  function cmd_init_config {
    function generate_moniker {
      id=$(docker create "${NODE_IMAGE}") &&
        name=$(docker inspect --format "{{ .Name }}" "${id}" | cut -c 2-) &&
        docker rm --force --volumes "${id}" >/dev/null 2>&1 &&
        echo "${name}"
    }

    function query_min_price {
      curl -fsSL "https://lcd.sentinel.co/cosmos/params/v1beta1/params?key=MinPrice&subspace=vpn/node" |
        jq -r '.param.value' | jq -s '.[] | sort_by(.denom) | .[] | .amount + .denom' |
        sed -e 's/"//g' | sed -z 's/\n/,/g' | sed 's/.$//'
    }

    function cmd_help {
      echo "Usage: ${0} init config COMMAND OPTIONS"
      echo ""
      echo "Commands:"
      echo "  help    Print the help message"
      echo ""
      echo "Options:"
      echo "  -f, --force    Force the initialization"
    }

    local force=0

    if [[ "${#}" -gt 0 ]]; then
      case "${1}" in
        "-f" | "--force")
          force=1
          ;;
        "help")
          cmd_help && return 0
          ;;
        *)
          echo "Error: invalid command or option \"${1}\"" && return 1
          ;;
      esac
    fi

    PUBLIC_IP=$(curl -fsSL https://ifconfig.me)

    local chain_rpc_addresses="https://rpc.sentinel.co:443,https://rpc.sentinel.quokkastake.io:443,https://sentinel-rpc.badgerbite.io:443"
    local handshake_enable=false
    local keyring_backend=file
    local node_ipv4_address=
    local node_listen_on="0.0.0.0:${PORTS[0]}"
    local node_moniker && node_moniker=$(generate_moniker)
    local node_price && node_price=$(query_min_price)
    local node_provider=
    local node_remote_url="https://${PUBLIC_IP}:${PORTS[0]}"
    local node_type="${NODE_TYPE}"

    echo "Initializing the configuration..."
    must_run config init --force="${force}"

    read -p "Enter chain_rpc_addresses [${chain_rpc_addresses}]:" -r input
    if [[ -n "${input}" ]]; then chain_rpc_addresses="${input}"; fi
    config_set "chain.rpc_addresses" "${chain_rpc_addresses}"

    read -p "Enter handshake_enable [${handshake_enable}]:" -r input
    if [[ -n "${input}" ]]; then handshake_enable="${input}"; fi
    config_set "handshake.enable" "${handshake_enable}"

    read -p "Enter keyring_backend [${keyring_backend}]:" -r input
    if [[ -n "${input}" ]]; then keyring_backend="${input}"; fi
    config_set "keyring.backend" "${keyring_backend}"

    read -p "Enter node_ipv4_address:" -r input
    if [[ -n "${input}" ]]; then node_ipv4_address="${input}"; fi
    config_set "node.ipv4_address" "${node_ipv4_address}"

    read -p "Enter node_listen_on [${node_listen_on}]:" -r input
    if [[ -n "${input}" ]]; then node_listen_on="${input}"; fi
    config_set "node.listen_on" "${node_listen_on}"

    read -p "Enter node_moniker [${node_moniker}]:" -r input
    if [[ -n "${input}" ]]; then node_moniker="${input}"; fi
    config_set "node.moniker" "${node_moniker}"

    read -p "Enter node_price [${node_price}]:" -r input
    if [[ -n "${input}" ]]; then node_price="${input}"; fi
    config_set "node.price" "${node_price}"

    read -p "Enter node_provider:" -r input
    if [[ -n "${input}" ]]; then node_provider="${input}"; fi
    config_set "node.provider" "${node_provider}"

    read -p "Enter node_remote_url [${node_remote_url}]:" -r input
    if [[ -n "${input}" ]]; then node_remote_url="${input}"; fi
    config_set "node.remote_url" "${node_remote_url}"

    read -p "Enter node_type [${node_type}]:" -r input
    if [[ -n "${input}" ]]; then
      NODE_TYPE="${input}"
      node_type="${input}"
    fi
    config_set "node.type" "${node_type}"
  }

  function cmd_init_keys {
    read -p "Recover the existing account?:" -r input
    if [[ "${input}" == "yes" ]]; then
      must_run keys add --recover
    else
      must_run keys add
    fi
  }

  function cmd_init_v2ray {
    function cmd_help {
      echo "Usage: ${0} init v2ray COMMAND OPTIONS"
      echo ""
      echo "Commands:"
      echo "  help    Print the help message"
      echo ""
      echo "Options:"
      echo "  -f, --force    Force the initialization"
    }

    local force=0

    if [[ "${#}" -gt 0 ]]; then
      case "${1}" in
        "-f" | "--force")
          force=1
          ;;
        "help")
          cmd_help && return 0
          ;;
        *)
          echo "Error: invalid command or option \"${1}\"" && return 1
          ;;
      esac
    fi

    local listen_port=${PORTS[1]}
    local transport=grpc

    echo "Initializing the V2Ray configuration..."
    must_run v2ray config init --force="${force}"

    read -p "Enter vmess.listen_port [${listen_port}]:" -r input
    if [[ -n "${input}" ]]; then listen_port="${input}"; fi
    v2ray_config_set "vmess.listen_port" "${listen_port}"

    read -p "Enter vmess.transport [${transport}]:" -r input
    if [[ -n "${input}" ]]; then transport="${input}"; fi
    v2ray_config_set "vmess.transport" "${transport}"
  }

  function cmd_init_wireguard {
    function cmd_help {
      echo "Usage: ${0} init wireguard COMMAND OPTIONS"
      echo ""
      echo "Commands:"
      echo "  help    Print the help message"
      echo ""
      echo "Options:"
      echo "  -f, --force    Force the initialization"
    }

    local force=0

    if [[ "${#}" -gt 0 ]]; then
      case "${1}" in
        "-f" | "--force")
          force=1
          ;;
        "help")
          cmd_help && return 0
          ;;
        *)
          echo "Error: invalid command or option \"${1}\"" && return 1
          ;;
      esac
    fi

    local listen_port=${PORTS[1]}

    echo "Initializing the WireGuard configuration..."
    must_run wireguard config init --force="${force}"

    read -p "Enter listen_port [${listen_port}]:" -r input
    if [[ -n "${input}" ]]; then listen_port="${input}"; fi
    wireguard_config_set "listen_port" "${listen_port}"
  }

  function cmd_init_all {
    function cmd_help {
      echo "Usage: ${0} init all COMMAND OPTIONS"
      echo ""
      echo "Commands:"
      echo "  help    Print the help message"
      echo ""
      echo "Options:"
      echo "  -f, --force    Force the initialization"
    }

    if [[ "${#}" -gt 0 ]]; then
      case "${1}" in
        "-f" | "--force") ;;
        "help")
          cmd_help && return 0
          ;;
        *)
          echo "Error: invalid command or option \"${1}\"" && return 1
          ;;
      esac
    fi

    cmd_init_config "${@}"
    if [[ "${NODE_TYPE}" == "v2ray" ]]; then cmd_init_v2ray "${@}"; fi
    if [[ "${NODE_TYPE}" == "wireguard" ]]; then cmd_init_wireguard "${@}"; fi
    cmd_init_keys "${@}"
  }

  function cmd_help {
    echo "Usage: ${0} init [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  all          Initialize everything"
    echo "  config       Initialize the config.toml file"
    echo "  help         Print the help message"
    echo "  keys         Initialize the keys"
    echo "  v2ray        Initialize the v2ray.toml file"
    echo "  wireguard    Initialize the wireguard.toml file"
  }

  if [[ "${#}" -gt 0 ]]; then
    case "${1}" in
      "all")
        shift
        cmd_init_all "${@}"
        ;;
      "config")
        shift
        cmd_init_config "${@}"
        ;;
      "help")
        cmd_help
        ;;
      "keys")
        cmd_init_keys "${@}"
        ;;
      "v2ray")
        shift
        cmd_init_v2ray "${@}"
        ;;
      "wireguard")
        shift
        cmd_init_wireguard "${@}"
        ;;
      *)
        echo "Error: invalid command or option \"${1}\"" && return 1
        ;;
    esac
  else
    cmd_help
  fi
}

function cmd_setup {
  if [[ "$EUID" -ne 0 ]]; then
    echo "Error: please run this command with sudo privileges" && return 1
  fi

  function install_packages {
    echo "Installing the packages ${*}"
    DEBIAN_FRONTEND=noninteractive apt-get install --quiet --yes "${@}"
  }

  function setup_docker {
    function install {
      if ! command -v docker &>/dev/null; then
        curl -fsSL -o /tmp/get-docker.sh https://get.docker.com
        sh /tmp/get-docker.sh
      fi
    }

    function setup_ipv6 {
      cat <<EOF >/etc/docker/daemon.json
{
    "ipv6": true,
    "fixed-cidr-v6": "2001:db8:1::/64"
}
EOF
    }

    install
    setup_ipv6
    systemctl restart docker
  }

  function setup_iptables {
    install_packages iptables-persistent
    rule=(POSTROUTING -s 2001:db8:1::/64 ! -o docker0 -j MASQUERADE)
    ip6tables -t nat -C "${rule[@]}" 2>/dev/null || ip6tables -t nat -A "${rule[@]}"
    ip6tables-save >/etc/iptables/rules.v6
  }

  function setup_tls {
    openssl req -new \
      -newkey ec \
      -pkeyopt ec_paramgen_curve:prime256v1 \
      -subj "/C=NA/ST=NA/L=./O=NA/OU=./CN=." \
      -x509 \
      -sha256 \
      -days 365 \
      -nodes \
      -out "${NODE_DIR}/tls.crt" \
      -keyout "${NODE_DIR}/tls.key"
  }

  apt-get update
  install_packages curl git jq openssl
  setup_docker
  setup_iptables
  setup_tls
  docker pull "${NODE_IMAGE}"
}

function cmd_start {
  function cmd_help {
    echo "Usage: ${0} start COMMAND OPTIONS"
    echo ""
    echo "Commands:"
    echo "  help    Print the help message"
    echo ""
    echo "Options:"
    echo "  -a, --attach    Start the node container attached"
  }

  local detach=1
  local rm=0

  if [[ "${#}" -gt 0 ]]; then
    case "${1}" in
      "-a" | "--attach")
        detach=0
        rm=1
        ;;
      "help")
        cmd_help && return 0
        ;;
      *)
        echo "Error: invalid command or option \"${1}\"" && return 1
        ;;
    esac
  fi

  local node_api_port=
  local v2ray_vmess_port=
  local wireguard_port=

  function read_config {
    found=false
    while IFS= read -r line; do
      if [[ "${line}" == "[node]" ]]; then
        found=true
      fi
      if [[ "${found}" ]] && [[ "${line}" == *"listen_on"* ]]; then
        node_api_port=$(echo "${line}" | cut -d '=' -f 2 | cut -d '"' -f 2 | cut -d ':' -f 2)
      fi
      if [[ "${found}" ]] && [[ "${line}" == *"type"* ]]; then
        NODE_TYPE=$(echo "${line}" | cut -d '=' -f 2 | cut -d '"' -f 2)
      fi
    done <"${NODE_DIR}/config.toml"
  }

  function read_wireguard_config {
    while IFS= read -r line; do
      if [[ "${line}" == *"listen_port"* ]]; then
        wireguard_port=$(echo "${line}" | cut -d '=' -f 2 | cut -d ' ' -f 2)
      fi
    done <"${NODE_DIR}/wireguard.toml"
  }

  function read_v2ray_config {
    found=false
    while IFS= read -r line; do
      if [[ "${line}" == "[vmess]" ]]; then
        found=true
      fi
      if [[ "${found}" ]] && [[ "${line}" == *"listen_port"* ]]; then
        v2ray_vmess_port=$(echo "${line}" | cut -d '=' -f 2 | cut -d ' ' -f 2)
      fi
    done <"${NODE_DIR}/v2ray.toml"
  }

  read_config
  if [[ "${NODE_TYPE}" == "v2ray" ]]; then
    read_v2ray_config
    docker run \
      --detach="${detach}" \
      --name="${CONTAINER_NAME}" \
      --rm="${rm}" \
      --tty \
      --volume "${NODE_DIR}:/root/.sentinelnode" \
      --publish "${node_api_port}:${node_api_port}/tcp" \
      --publish "${v2ray_vmess_port}:${v2ray_vmess_port}/tcp" \
      "${NODE_IMAGE}" process start
  fi
  if [[ "${NODE_TYPE}" == "wireguard" ]]; then
    read_wireguard_config
    docker run \
      --detach="${detach}" \
      --name="${CONTAINER_NAME}" \
      --rm="${rm}" \
      --tty \
      --volume /lib/modules:/lib/modules \
      --volume "${NODE_DIR}:/root/.sentinelnode" \
      --cap-drop ALL \
      --cap-add NET_ADMIN \
      --cap-add NET_BIND_SERVICE \
      --cap-add NET_RAW \
      --cap-add SYS_MODULE \
      --sysctl net.ipv4.ip_forward=1 \
      --sysctl net.ipv6.conf.all.disable_ipv6=0 \
      --sysctl net.ipv6.conf.all.forwarding=1 \
      --sysctl net.ipv6.conf.default.forwarding=1 \
      --publish "${node_api_port}:${node_api_port}/tcp" \
      --publish "${wireguard_port}:${wireguard_port}/udp" \
      "${NODE_IMAGE}" process start
  fi
}

function cmd_status {
  id=$(docker ps --all --filter name="${CONTAINER_NAME}" --quiet)
  if [[ -n "${id}" ]]; then
    running=$(docker container inspect --format "{{ .State.Running }}" "${CONTAINER_NAME}")
    if [[ "${running}" == "true" ]]; then
      echo "Node is running..."
    else
      exit_code=$(docker container inspect --format "{{ .State.ExitCode }}" "${CONTAINER_NAME}")
      echo "Node is not running and has exited with code ${exit_code}"
    fi
    docker logs --tail 20 "${CONTAINER_NAME}"
  else
    echo "Error: node container does not exist" && return 1
  fi
}

function cmd_stop {
  id=$(docker ps --filter name="${CONTAINER_NAME}" --quiet)
  if [[ -n "${id}" ]]; then
    docker stop "${id}"
  else
    echo "Error: node is not running" && return 1
  fi
}

function cmd_remove {
  id=$(docker ps --all --filter name="${CONTAINER_NAME}" --quiet)
  if [[ -n "${id}" ]]; then
    docker rm --force --volumes "${id}"
  else
    echo "Error: node container does not exist" && return 1
  fi
}

function cmd_restart {
  function cmd_help {
    echo "Usage: ${0} restart COMMAND OPTIONS"
    echo ""
    echo "Commands:"
    echo "  help    Print the help message"
    echo ""
    echo "Options:"
    echo "  -a, --attach    Start the node container attached"
  }

  if [[ "${#}" -gt 0 ]]; then
    case "${1}" in
      "-a" | "--attach") ;;
      "help")
        cmd_help && return 0
        ;;
      *)
        echo "Error: invalid command or option \"${1}\"" && return 1
        ;;
    esac
  fi

  stop
  remove
  cmd_start "${@}"
}

function cmd_update {
  stop
  remove
  docker rmi --force "${NODE_IMAGE}"
  docker pull "${NODE_IMAGE}"
}

if [[ "${#}" -gt 0 ]]; then
  case "${1}" in
    "attach")
      shift
      cmd_attach "${@}"
      ;;
    "help")
      shift
      cmd_help
      ;;
    "init")
      shift
      cmd_init "${@}"
      ;;
    "setup")
      shift
      cmd_setup "${@}"
      ;;
    "start")
      shift
      cmd_start "${@}"
      ;;
    "status")
      shift
      cmd_status "${@}"
      ;;
    "stop")
      shift
      cmd_stop "${@}"
      ;;
    "remove")
      shift
      cmd_remove "${@}"
      ;;
    "restart")
      shift
      cmd_restart "${@}"
      ;;
    "update")
      shift
      cmd_update "${@}"
      ;;
    *)
      echo "Error: invalid command or option \"${1}\"" && return 1
      ;;
  esac
else
  cmd_help
fi
