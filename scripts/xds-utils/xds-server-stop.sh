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

# Stop it gracefully
pkill -INT xds-server
sleep 1

# Seems no stopped, nasty kill
nbProc=$(ps -ef |grep xds-server |grep -v grep |wc -l)
if [ "$nbProc" != "0" ]; then
    pkill -KILL xds-server
fi

nbProc=$(ps -ef |grep syncthing |grep -v grep |wc -l)
if [ "$nbProc" != "0" ]; then
    pkill -KILL syncthing
    pkill -KILL syncthing-inotify
fi

exit 0
