#!/usr/bin/env bash
set -a

# Setup
local_dir="$(dirname "$0")"
PARENT_DIR=`pwd`
GETH_DOCKER_CONTAINER_NAME="geth-node"
GETH_DOCKER_CONTAINER_WAS_RUNNING=`docker ps -a --filter "name=${DOCKER_CONTAINER_NAME}" --filter "status=running" --quiet`
echo "Running: [${GETH_DOCKER_CONTAINER_WAS_RUNNING}]"

################# Run Dependencies #########################
for path in ${local_dir}/test-dependencies/*; do
    [ -d "${path}" ] || continue # if not a directory, skip
    source "${path}/env_vars.sh" # Every dependency should have env_vars.sh + run.sh executable files
    echo "Executing [${path}/run.sh]"
    ${path}/run.sh
    if [ $? -ne 0 ]; then
        exit 1
    fi
done
############################################################

################# Prepare for tests ########################
# Even if other `env_vars.sh` might hold this variable
# Let's not count on it and be clear instead
if [ -z ${CENT_ETHEREUM_CONTRACTS_DIR} ]; then
    CENT_ETHEREUM_CONTRACTS_DIR=$GOPATH/src/github.com/CentrifugeInc/centrifuge-ethereum-contracts
fi

# If the contracts dir doesn't exist - it has likely not been "installed" yet
# Installing it in that case so it is usable
if [ ! -d ${CENT_ETHEREUM_CONTRACTS_DIR} ]; then
	echo "Ethereum contracts folder not found at ${CENT_ETHEREUM_CONTRACTS_DIR}. Checking them out."
	# git clone here instead of `go get` as `go get` defaults back to HTTPS which causes issues
    # with certificate-based github authentication
    mkdir -p ${CENT_ETHEREUM_CONTRACTS_DIR}
    git clone git@github.com:CentrifugeInc/centrifuge-ethereum-contracts.git ${CENT_ETHEREUM_CONTRACTS_DIR}
fi
cd ${CENT_ETHEREUM_CONTRACTS_DIR}
npm install

# Unlock User to Run Migration and Run it
docker run -it --net=host ethereum/client-go:$GETH_DOCKER_VERSION attach "${CENT_ETHEREUM_NODEURL}" --exec "personal.unlockAccount('0x${CENT_ETHEREUM_ACCOUNTS_MIGRATE_ADDRESS}', '${CENT_ETHEREUM_ACCOUNTS_MIGRATE_PASSWORD}')"
truffle migrate --network localgeth -f 2
status=$?

cd ${PARENT_DIR}

############################################################

################# Run Tests ################################
if [ $status -eq 0 ]; then
  statusAux=0
  for path in ${local_dir}/tests/*; do
    [ -x "${path}" ] || continue # if not an executable, skip
    ./$path
    statusAux="$(( $statusAux | $? ))"
  done
  # Store status of tests
  status=$statusAux
fi
############################################################

################# CleanUp ##################################
if [ -n "${GETH_DOCKER_CONTAINER_WAS_RUNNING}" ]; then
    echo "Container ${GETH_DOCKER_CONTAINER_NAME} was already running before the test setup. Not tearing it down as the assumption is that the container was started outside this context."
else
    echo "Bringing GETH Daemon Down"
    docker rm -f geth-node

    # Cleaning extra DAG file, so we do not cache it - travis
    if [[ "X${RUN_CONTEXT}" == "Xtravis" ]];
    then
      new_dag=`ls -ltr $DATA_DIR/.ethash/* | tail -1 | awk '{print $9}' | tr -d '\n'`
      rm -Rf $new_dag
    fi
fi
############################################################

################# Propagate test status ####################
exit $status
############################################################