#!/bin/bash
if [[ ${CIRCLE_BUILD_NUM} ]]; then
  iteration=${CIRCLE_BUILD_NUM}
else
  iteration=0
fi

for package_type in deb ; do
  fpm -t ${package_type} \
    -s dir \
    --name kapo \
    --version $(./kapo --version | awk '{print $3}') \
    --iteration ${iteration} \
    --license gplv3 \
    --vendor 'Sam Pointer' \
    --provides kapo \
    --architecture $(uname -m) \
    --maintainer sam@outsidethe.net \
    --description "Wrap any command in a status socket." \
    --url "https://github.com/sampointer/kapo" \
    --prefix /usr/local/bin \
    kapo
done

if [[ ${CIRCLE_ARTIFACTS} ]]; then
  cp kapo ${CIRCLE_ARTIFACTS}
  cp kapo*.deb ${CIRCLE_ARTIFACTS}
fi