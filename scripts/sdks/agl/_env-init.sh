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

. /etc/xdtrc

[ -z "$XDT_SDK" ] && XDT_SDK=/xdt/sdk

export SDK_FAMILY_NAME="agl"
export SDK_ROOT_DIR="$XDT_SDK"
export SDK_ENV_SETUP_FILENAME="environment-setup-*"
export SDK_DATABASE="http://iot.bzh/download/public/XDS/sdk/sdks_latest.json"

[ "$1" = "-print" ] && { env; }
