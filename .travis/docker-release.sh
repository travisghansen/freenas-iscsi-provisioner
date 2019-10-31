#!/bin/bash

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin

if [[ -n "${TRAVIS_TAG}" ]];then
	docker build --pull -t travisghansen/freenas-iscsi-provisioner:${TRAVIS_TAG} .
	docker push travisghansen/freenas-iscsi-provisioner:${TRAVIS_TAG}
elif [[ -n "${TRAVIS_BRANCH}" ]];then
	if [[ "${TRAVIS_BRANCH}" == "master" ]];then
		docker build --pull -t travisghansen/freenas-iscsi-provisioner:latest .
		docker push travisghansen/freenas-iscsi-provisioner:latest
	else
		docker build --pull -t travisghansen/freenas-iscsi-provisioner:${TRAVIS_BRANCH} .
		docker push travisghansen/freenas-iscsi-provisioner:${TRAVIS_BRANCH}
	fi
else
	:
fi
