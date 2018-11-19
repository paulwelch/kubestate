#!/usr/bin/env bash

# Copyright © 2017 The Kubicorn Authors
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

read -r -d '' EXPECTED <<EOF
// Copyright © DATE The Kubicorn Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
EOF

FILES=$(find . -name "*.go" -not -path "./vendor/*")
for FILE in $FILES; do
	if [ "$FILE" == "./bootstrap/bootstrap.go" ]; then
            continue
    fi
    # Replace the actual year with DATE so we can ignore the year when
    # checking for the license header.
    CONTENT=$(head -n 13 $FILE | sed -E -e 's/Copyright © [0-9]+/Copyright © DATE/')
    if [ "$CONTENT" != "$EXPECTED" ]; then
        # Replace DATE with the current year.
        EXPECTED=$(echo "$EXPECTED" | sed -E -e "s/Copyright © DATE/Copyright © $(date +%Y)/")
		echo -e "$EXPECTED\n\n$(cat $FILE)" > $FILE
    fi
done
