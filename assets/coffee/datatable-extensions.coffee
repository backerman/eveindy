# Copyright © 2014–6 Brad Ackerman.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

jQuery.extend jQuery.fn.dataTableExt.oSort,
  "numeric-inf-pre": (a) ->
    if a is '∞' then +Infinity else parseFloat a

  "numeric-inf-asc": (a, b) ->
    ((a < b) ? -1 : ((a > b) ? 1 : 0))

  "numeric-inf-desc": (a, b) ->
    ((a < b) ? 1 : ((a > b) ? -1 : 0))
