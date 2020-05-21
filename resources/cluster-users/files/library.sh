#!/bin/bash

export GLOBAL_RETRY_TIMER=15

function LogInAsUser() {
    # export local envs to make then avaliable in the internal shells
    export USER="$1"
    export PASS="$2"
    export DEX_SERVICE_SERVICE_HOST
    export DEX_SERVICE_SERVICE_PORT_HTTP
    export IAM_KUBECONFIG_SVC_FQDN

    # Handle registration request
    timeout ${GLOBAL_RETRY_TIMER} bash -c 'until curl -s -k -f -X GET -H "Content-Type: application/x-www-form-urlencoded" "${DEX_SERVICE_SERVICE_HOST}:${DEX_SERVICE_SERVICE_PORT_HTTP}/auth?response_type=id_token%20token&client_id=kyma-client&redirect_uri=http://127.0.0.1:5555/callback&scope=openid%20profile%20email%20groups&nonce=vF7FAQlqq41CObeUFYY0ggv1qEELvfHaXQ0ER4XM" > registration_request ; do sleep 5; done'
    export REQUEST_ID=$(grep '/auth/local?req' < registration_request | cut -d '"' -f 2 | cut -d '?' -f 2)
    rm -f registration_request

    # Handle login request
    timeout ${GLOBAL_RETRY_TIMER} bash -c 'until curl -s -X POST -F "login=${USER}" -F "password=${PASS}" "${DEX_SERVICE_SERVICE_HOST}:${DEX_SERVICE_SERVICE_PORT_HTTP}/auth/local?${REQUEST_ID}" ; do sleep 5; done'

    # Handle approval response
    timeout ${GLOBAL_RETRY_TIMER} bash -c 'until curl -s -X GET "${DEX_SERVICE_SERVICE_HOST}:${DEX_SERVICE_SERVICE_PORT_HTTP}/approval?${REQUEST_ID}" > approval_response ; do sleep 5; done'
    export AUTH_TOKEN=$(cat approval_response | grep -o -P '(?<=id_token=).*(?=&amp;state)')
    rm -f approval_response

    # Get kubeconfig
    timeout ${GLOBAL_RETRY_TIMER} bash -c 'until curl -s -k -f -H "Authorization: Bearer ${AUTH_TOKEN}" "${IAM_KUBECONFIG_SVC_FQDN}/kube-config" -o "${PWD}/kubeconfig-${USER}" ; do sleep 5; done'
    if [[ ! -s "${PWD}/kubeconfig-${USER}" ]]; then
        echo "---> KUBECONFIG not created, or is empty!"
        exit 1
    fi
    echo "---> Login Successful! Created ${PWD}/kubeconfig-${USER}"
}

function CreateBindings() {
    export DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
    timeout ${GLOBAL_RETRY_TIMER} bash -c 'until kubectl create namespace "${CUSTOM_NAMESPACE}" ; do sleep 5; done'
    timeout ${GLOBAL_RETRY_TIMER} bash -c 'until kubectl create rolebinding namespace-developer --clusterrole=kyma-developer --user=${DEVELOPER_EMAIL} -n ${CUSTOM_NAMESPACE} ; do sleep 5; done'
    timeout ${GLOBAL_RETRY_TIMER} bash -c 'until kubectl create -f "${DIR}/kyma-test-bindings.yaml" ; do sleep 5; done'
}

function DeleteBindings() {
    export DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
    timeout ${GLOBAL_RETRY_TIMER} bash -c 'until kubectl delete -f "${DIR}/kyma-test-bindings.yaml" ; do sleep 5; done'
    timeout ${GLOBAL_RETRY_TIMER} bash -c 'until kubectl delete namespace "${CUSTOM_NAMESPACE}" ; do sleep 5; done'
}

function testPermissions() {
    export OPERATION="$1"
    export RESOURCE="$2"
    export TEST_NS="$3"
    local EXPECTED="$4"
    local TEST="not-set-yet"

    if [[ "${TEST_NS}" != "--all-namespaces" ]]; then
        TEST_NS="-n${TEST_NS}"
    fi

    SECONDS=0
    until [[ ${SECONDS} -ge ${GLOBAL_RETRY_TIMER} ]]; do
        set +e
        TEST=$(kubectl auth can-i "${OPERATION}" "${RESOURCE}" "${TEST_NS}")
        set -e
        if [[ "${TEST}" == "${EXPECTED}" ]]; then
            echo "----> PASSED"
            return 0
        fi
        sleep 5
    done

    echo "----> |FAIL| Expected: ${EXPECTED}, Actual: ${TEST}"
    return 1
}

function testComponent() {
    local -r userEmail="${1}"
    local -r testNamespace="${2}"
    local -r viewAllowed="${3}"
    local -r editAllowed="${4}"
    shift 4
    local -r resources=("$@")

    local viewText=""
    if [[ "${viewAllowed}" == "no" ]]; then
        viewText=" NOT"
    fi

    local editText=""
    if [[ "${editAllowed}" == "no" ]]; then
        editText=" NOT"
    fi

    readonly viewText editText
    # View
    for resource in "${resources[@]}"; do
        for operation in "${VIEW_OPERATIONS[@]}"; do
            echo "--> ${userEmail} should${viewText} be able to ${operation} ${resource} CR in ${testNamespace}"
            testPermissions "${operation}" "${resource}" "${testNamespace}" "${viewAllowed}"
        done
    done
    # Edit
    for resource in "${resources[@]}"; do
        for operation in "${EDIT_OPERATIONS[@]}"; do
            echo "--> ${userEmail} should${editText} be able to ${operation} ${resource} CR in ${testNamespace}"
            testPermissions "${operation}" "${resource}" "${testNamespace}" "${editAllowed}"
        done
    done
}

function testComponentClusterScoped() {
    local -r userEmail="${1}"
    local -r viewAllowed="${2}"
    local -r editAllowed="${3}"
    shift 3
    local -r resources=("$@")

    local viewText=""
    if [[ "${viewAllowed}" == "no" ]]; then
        viewText=" NOT"
    fi

    local editText=""
    if [[ "${editAllowed}" == "no" ]]; then
        editText=" NOT"
    fi

    readonly viewText editText
    # View
    for resource in "${resources[@]}"; do
        for operation in "${VIEW_OPERATIONS[@]}"; do
            echo "--> ${userEmail} should${viewText} be able to ${operation} ${resource} CR"
            testPermissionsClusterScoped "${operation}" "${resource}" "${viewAllowed}"
        done
    done
    # Edit
    for resource in "${resources[@]}"; do
        for operation in "${EDIT_OPERATIONS[@]}"; do
            echo "--> ${userEmail} should${editText} be able to ${operation} ${resource} CR"
            testPermissionsClusterScoped "${operation}" "${resource}" "${editAllowed}"
        done
    done
}

function testDescribe() {
	local RESOURCE="$1"
	local TEST_NS="$2"
	local EXPECTED="$3"

	if [[ "${TEST_NS}" != "--all-namespaces" ]]; then
		TEST_NS="-n${TEST_NS}"
	fi

	sleep 0.1

	local result=1
	set +e
	kubectl describe "${RESOURCE}" "${TEST_NS}" > /dev/null
	result=$?
	set -e

	local IS_OK="false"

	if [[ "${EXPECTED}" == "yes" ]] && [[ ${result} -eq 0 ]]; then
		IS_OK="true"
	fi
	if [[ "${EXPECTED}" == "no" ]] && [[ ${result} -ne 0 ]]; then
		IS_OK="true"
	fi

	if [[ "${IS_OK}" == "true" ]]; then
			echo "----> PASSED"
			return 0
	fi

	echo "----> |FAIL| Expected: ${EXPECTED}, Actual: ${TEST}"
	return 1
}

function testDescribeClusterScoped() {
	local RESOURCE="$1"
	local EXPECTED="$2"

	testDescribe "${RESOURCE}" --all-namespaces "${EXPECTED}"
}

function cleanup() {
    unset KUBECONFIG
    DeleteBindings
}