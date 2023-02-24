#!/bin/bash -eu

CONTAINER_NAME=sentinelnode
NODE_DIR="${HOME}/.sentinelnode"
NODE_IMAGE=ghcr.io/sentinel-official/dvpn-node:latest

function attach {
  id=$(docker ps --all --filter name="${CONTAINER_NAME}" --quiet)
  if [[ -n "${id}" ]]; then
    docker attach "${id}"
  else
    echo "Error: node is not running" && exit 1
  fi
}

function help {
  echo "Usage: ${0} [COMMAND]"
  echo ""
  echo "Commands:"
  echo "  attach     Attach to the already running node"
  echo "  help       Print the help message"
  echo "  init       Initialize the configuration"
  echo "  setup      Install the dependencies and setup the requirements"
  echo "  start      Start the node"
  echo "  stop       Stop the node"
  echo "  remove     Remove the node container"
  echo "  restart    Restart the node"
  echo "  update     Update the node to the latest version"
  exit 0
}

function init {
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
    code="${?}"
    if [[ "${code}" -ne 0 ]]; then
      exit 125
    fi
    if [[ -n "${output}" ]]; then
      echo "${output}"
    fi
    if [[ "${output}" == *"Error"* ]]; then
      exit 1
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

  function init_config {
    function help {
      echo "Usage: ${0} init config COMMAND OPTIONS"
      echo ""
      echo "Commands:"
      echo "  help    Print the help message"
      echo ""
      echo "Options:"
      echo "  -f, --force    Force the initialization"
      exit 0
    }

    local force=0

    if [[ "${#}" -gt 0 ]]; then
      case "${1}" in
        "-f" | "--force")
          force=1
          ;;
        "help")
          help
          ;;
        *)
          echo "Error: invalid command or option \"${1}\""
          exit 1
          ;;
      esac
    fi

    PUBLIC_IP=$(curl --silent https://icanhazip.com)

    local chain_rpc_address=https://rpc.sentinel.co:443
    local handshake_enable=false
    local keyring_backend=file
    local node_listen_on="0.0.0.0:${PORTS[0]}"
    local node_moniker=
    local node_price=
    local node_provider=
    local node_remote_url="https://${PUBLIC_IP}:${PORTS[0]}"
    local node_type="${NODE_TYPE}"

    echo "Initializing the configuration..."
    must_run config init --force="${force}"

    read -p "Enter chain_rpc_address [${chain_rpc_address}]:" -r input
    if [[ -n "${input}" ]]; then chain_rpc_address="${input}"; fi
    config_set "chain.rpc_address" "${chain_rpc_address}"

    read -p "Enter handshake_enable [${handshake_enable}]:" -r input
    if [[ -n "${input}" ]]; then handshake_enable="${input}"; fi
    config_set "handshake.enable" "${handshake_enable}"

    read -p "Enter keyring_backend [${keyring_backend}]:" -r input
    if [[ -n "${input}" ]]; then keyring_backend="${input}"; fi
    config_set "keyring.backend" "${keyring_backend}"

    read -p "Enter node_listen_on[$node_listen_on]:" -r input
    if [[ -n "${input}" ]]; then node_listen_on="${input}"; fi
    config_set "node.listen_on" "${node_listen_on}"

    read -p "Enter node_moniker:" -r input
    if [[ -n "${input}" ]]; then node_moniker="${input}"; fi
    config_set "node.moniker" "${node_moniker}"

    read -p "Enter node_price:" -r input
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

  function init_keys {
    read -p "Recover the existing account?:" -r input
    if [[ "${input}" == "yes" ]]; then
      must_run keys add --recover
    else
      must_run keys add
    fi
  }

  function init_v2ray {
    function help {
      echo "Usage: ${0} init v2ray COMMAND OPTIONS"
      echo ""
      echo "Commands:"
      echo "  help    Print the help message"
      echo ""
      echo "Options:"
      echo "  -f, --force    Force the initialization"
      exit 0
    }

    local force=0

    if [[ "${#}" -gt 0 ]]; then
      case "${1}" in
        "-f" | "--force")
          force=1
          ;;
        "help")
          help
          ;;
        *)
          echo "Error: invalid command or option \"${1}\""
          exit 1
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

  function init_wireguard {
    function help {
      echo "Usage: ${0} init wireguard COMMAND OPTIONS"
      echo ""
      echo "Commands:"
      echo "  help    Print the help message"
      echo ""
      echo "Options:"
      echo "  -f, --force    Force the initialization"
      exit 0
    }

    local force=0

    if [[ "${#}" -gt 0 ]]; then
      case "${1}" in
        "-f" | "--force")
          force=1
          ;;
        "help")
          help
          ;;
        *)
          echo "Error: invalid command or option \"${1}\""
          exit 1
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

  function init_all {
    function help {
      echo "Usage: ${0} init all COMMAND OPTIONS"
      echo ""
      echo "Commands:"
      echo "  help    Print the help message"
      echo ""
      echo "Options:"
      echo "  -f, --force    Force the initialization"
      exit 0
    }

    if [[ "${#}" -gt 0 ]]; then
      case "${1}" in
        "-f" | "--force") ;;

        "help")
          help
          ;;
        *)
          echo "Error: invalid command or option \"${1}\""
          exit 1
          ;;
      esac
    fi

    init_config "${@}"
    if [[ "${NODE_TYPE}" == "v2ray" ]]; then init_v2ray "${@}"; fi
    if [[ "${NODE_TYPE}" == "wireguard" ]]; then init_wireguard "${@}"; fi
    init_keys
  }

  function help {
    echo "Usage: ${0} init [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  all          Initialize everything"
    echo "  config       Initialize the config.toml file"
    echo "  help         Print the help message"
    echo "  keys         Initialize the keys"
    echo "  v2ray        Initialize the v2ray.toml file"
    echo "  wireguard    Initialize the wireguard.toml file"
    exit 0
  }

  case "${1}" in
    "all")
      shift
      init_all "${@}"
      ;;
    "config")
      shift
      init_config "${@}"
      ;;
    "" | "help")
      help
      ;;
    "keys")
      init_keys
      ;;
    "v2ray")
      shift
      init_v2ray "${@}"
      ;;
    "wireguard")
      shift
      init_wireguard "${@}"
      ;;
    *)
      echo "Error: invalid command or option \"${1}\""
      exit 1
      ;;
  esac
}

function setup {
  if [[ "$EUID" -ne 0 ]]; then
    echo "Error: please run this command with sudo privileges"
    exit 1
  fi

  function install_packages {
    echo "Installing the packages ${*}"
    for name in "${@}"; do
      if ! dpkg -s "${name}" &>/dev/null; then
        DEBIAN_FRONTEND=noninteractive apt-get install --yes "${name}"
      fi
    done
  }

  function setup_docker {
    function install {
      if ! command -v docker &>/dev/null; then
        curl -fsSL https://get.docker.com -o /tmp/get-docker.sh &&
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

    install &&
      setup_ipv6 &&
      systemctl restart docker
  }

  function setup_iptables {
    install_packages iptables-persistent &&
      rule=(POSTROUTING -s 2001:db8:1::/64 ! -o docker0 -j MASQUERADE) &&
      ip6tables -t nat -C "${rule[@]}" 2>/dev/null || ip6tables -t nat -A "${rule[@]}" &&
      ip6tables-save >/etc/iptables/rules.v6
  }

  function generate_tls {
    openssl req -new \
      -newkey ec \
      -pkeyopt ec_paramgen_curve:prime256v1 \
      -subj "/C=NA/ST=NA/L=./O=NA/OU=./CN=." \
      -x509 \
      -sha256 \
      -days 365 \
      -nodes \
      -out "${1}/tls.crt" \
      -keyout "${1}/tls.key"
  }

  apt-get update &&
    install_packages curl git openssl &&
    setup_docker &&
    setup_iptables &&
    mkdir -p "${NODE_DIR}" && generate_tls "${NODE_DIR}" &&
    docker pull "${NODE_IMAGE}"
}

function start {
  function help {
    echo "Usage: ${0} start COMMAND OPTIONS"
    echo ""
    echo "Commands:"
    echo "  help    Print the help message"
    echo ""
    echo "Options:"
    echo "  -a, --attach    Start the node container attached"
    exit 0
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
        help
        ;;
      *)
        echo "Error: invalid command or option \"${1}\""
        exit 1
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
    read_v2ray_config &&
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
    read_wireguard_config &&
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

function stop {
  id=$(docker ps --all --filter name="${CONTAINER_NAME}" --quiet)
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

function restart {
  function help {
    echo "Usage: ${0} restart COMMAND OPTIONS"
    echo ""
    echo "Commands:"
    echo "  help    Print the help message"
    echo ""
    echo "Options:"
    echo "  -a, --attach    Start the node container attached"
    exit 0
  }

  if [[ "${#}" -gt 0 ]]; then
    case "${1}" in
      "-a" | "--attach") ;;
      "help")
        help
        ;;
      *)
        echo "Error: invalid command or option \"${1}\""
        exit 1
        ;;
    esac
  fi

  stop &&
    remove &&
    start "${@}"
}

function update {
  stop &&
    remove &&
    docker rmi --force "${NODE_IMAGE}" &&
    docker pull "${NODE_IMAGE}"
}

case "${1}" in
  "attach")
    shift
    attach "${@}"
    ;;
  "" | "help")
    shift
    help
    ;;
  "init")
    shift
    init "${@}"
    ;;
  "setup")
    shift
    setup "${@}"
    ;;
  "start")
    shift
    start "${@}"
    ;;
  "stop")
    shift
    stop "${@}"
    ;;
  "remove")
    shift
    remove "${@}"
    ;;
  "restart")
    shift
    restart "${@}"
    ;;
  "run")
    shift
    must_run "${@}"
    ;;
  "update")
    shift
    update "${@}"
    ;;
  *)
    echo "Error: invalid command or option \"${1}\""
    exit 1
    ;;
esac
