#!/bin/bash
###########################################################################
# Copyright 2017 IoT.bzh
#
# author: Sebastien Douheret <sebastien@iot.bzh>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
###########################################################################

SDK_AGL_BASEURL="https://download.automotivelinux.org/AGL"
SDK_AGL_IOTBZH_BASEURL="http://iot.bzh/download/public/XDS/sdk"

# Define urls where SDKs can be downloaded
DOWNLOADABLE_URLS="
    ${SDK_AGL_BASEURL}/snapshots/master/latest/*/deploy/sdk

    ${SDK_AGL_BASEURL}/release/dab/3.99.3/m3ulcb-nogfx/deploy/sdk
    ${SDK_AGL_BASEURL}/release/dab/4.0.2/*/deploy/sdk

    ${SDK_AGL_BASEURL}/release/eel/4.99.4/*/deploy/sdk
    ${SDK_AGL_BASEURL}/release/eel/latest/*/deploy/sdk

    ${SDK_AGL_IOTBZH_BASEURL}
"

###


# Compute full urls list  (parse '*' characters)
urls=""
for url in $(echo $DOWNLOADABLE_URLS); do
    if [[ "$url" = *"*"* ]]; then
        bUrl=$(echo $url | cut -d'*' -f 1)
        eUrl=$(echo $url | cut -d'*' -f 2)
        dirs=$(curl -s ${bUrl} | grep '\[DIR\]' | grep -oP  'href="[^"]*"' | cut -d'"' -f 2)
        for dir in $(echo $dirs); do
            urls="$urls ${bUrl::-1}/${dir::-1}/${eUrl:1}"
        done
    else
        urls="$urls $url"
    fi
done

# Compute list of available/installable SDKs
sdksList=" "
for url in $(echo $urls); do
    htmlPage=$(curl -s --connect-timeout 10 "${url}/")
    files=$(echo ${htmlPage} | egrep -o 'href="[^"]*.sh"' | cut -d '"' -f 2)
    if [ "$?" != "0" ] || [ "${files}" = "" ]; then
        echo " IGNORED ${url}: no valid files found"
        continue
    fi

    for sdkFile in $(echo ${files}); do

        # assume that sdk name follow this format :
        #  _PROFILE_-_COMPILER_ARCH_-_TARGET_-crosssdk-_ARCH_-toolchain-_VERSION_.sh
        # for example:
        #  poky-agl-glibc-x86_64-agl-demo-platform-crosssdk-corei7-64-toolchain-4.0.1.sh

        [[ "${sdkFile}" != *"crosssdk"* ]] && { echo " IGNORED ${sdkFile}, not a valid sdk file"; continue; }

        echo "Processing ${sdkFile}"
        profile=$(echo "${sdkFile}" | sed -r 's/(.*)-glibc.*/\1/')
        version=$(echo "${sdkFile}" | sed -r 's/.*toolchain-(.*).sh/\1/')
        arch=$(echo "${sdkFile}" | sed -r 's/.*crosssdk-(.*)-toolchain.*/\1/')

        endUrl=${url#$SDK_AGL_BASEURL}
        if [ "${endUrl::4}" = "http" ]; then
            name=${profile}_${arch}_${version}
        else
            name=$(echo "AGL-$(echo ${endUrl} | cut -d'/' -f2,3,4,5)" | sed s:/:-:g)
        fi

        [ "${profile}" = "" ] && { echo " ERROR: profile not set" continue; }
        [ "${version}" = "" ] && { echo " ERROR: version not set" continue; }
        [ "${arch}" = "" ] && { echo " ERROR: arch not set" continue; }
        [ "${name}" = "" ] && { name=${profile}_${arch}_${version}; }

        sdkDate="$(echo "${htmlPage}" |egrep -o ${sdkFile/+/\\+}'</a>.*[0-9\-]+ [0-9]+:[0-9]+' |cut -d'>' -f 4|cut -d' ' -f1,2)"
        sdkSize="$(echo "${htmlPage}" |egrep -o  "${sdkFile/+/\\+}.*${sdkDate}.*[0-9\.MG]+</td>" |cut -d'>' -f7 |cut -d'<' -f1)"
        md5sum="$(wget -q -O - ${url}/${sdkFile/.sh/.md5} |cut -d' ' -f1)"

        read -r -d '' res <<- EndOfMessage
{
    "name":         "${name}",
    "description":  "AGL SDK ${arch} (version ${version})",
    "profile":      "${profile}",
    "version":      "${version}",
    "arch":         "${arch}",
    "path":         "",
    "url":          "${url}/${sdkFile}",
    "status":       "Not Installed",
    "date":         "${sdkDate}",
    "size":         "${sdkSize}",
    "md5sum":       "${md5sum}",
    "setupFile":    ""
},
EndOfMessage

        sdksList="${sdksList}${res}"
    done
done

OUT_FILE=$(dirname "$0")/sdks_$(date +"%F_%H%m").json

echo "[" > ${OUT_FILE}
echo "${sdksList::-1}" >> ${OUT_FILE}
echo "]" >> ${OUT_FILE}

echo "SDKs list successfully saved in ${OUT_FILE}"
