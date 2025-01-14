#!/usr/bin/env bash

cat <<EOF
                            __                __
    ____  __  ______  _____/ /_  ____  __  __/ /_
   / __ \/ / / / __ \/ ___/ __ \/ __ \/ / / / __/
  / /_/ / /_/ / / / / /__/ / / / /_/ / /_/ / /_
 / .___/\__,_/_/ /_/\___/_/ /_/\____/\__,_/\__/
/_/

EOF
pass_count=0
fail_count=0

tests=(
    "Simple cloud setup|punchout -config-file-path=config-good.toml -db-path='db.db' -jira-url='https://jira.company.com' -jira-installation-type 'cloud' -jira-token='XXX' -jira-username='example@example.com' -list-config|0"
    "Simple onpremise setup|punchout -config-file-path=config-good.toml -db-path='db.db' -jira-url='https://jira.company.com' -jira-installation-type 'onpremise' -jira-token='XXX' -list-config|0"
    "No installation type provided; should fallback to onpremise|punchout -config-file-path=config-good.toml -db-path='db.db' -jira-url='https://jira.company.com' -jira-token='XXX' -list-config|0"
    "Fallback comment|punchout -config-file-path=config-good.toml -db-path='db.db' -fallback-comment='test' -jira-url='https://jira.company.com' -jira-token='XXX' -list-config|0"
    "Incorrect installation type provided|punchout -config-file-path=config-good.toml -db-path='db.db' -jira-url='https://jira.company.com' -jira-installation-type 'blah' -jira-token='XXX' -list-config|1"
    "No token provided|punchout -config-file-path=config-good.toml -db-path='db.db' -jira-url='https://jira.company.com' -jira-installation-type 'onpremise' -list-config|1"
    "No username provided for cloud installation|punchout -config-file-path=config-good.toml -db-path='db.db' -jira-url='https://jira.company.com' -jira-installation-type 'cloud' -jira-token='XXX' -list-config|1"
    "Incorrect config file path|punchout -config-file-path=config-absent.toml -db-path='db.db' -jira-url='https://jira.company.com' -jira-token='XXX' -list-config|1"
    "Bad config file|punchout -config-file-path=config-bad.toml -db-path='db.db' -jira-url='https://jira.company.com' -jira-token='XXX' -list-config|1"
    "Empty jira url|punchout -config-file-path=config-good.toml -db-path='db.db' -jira-url='' -jira-token='XXX' -list-config|1"
    "Empty jira jira token|punchout -config-file-path=config-good.toml -db-path='db.db' -jira-url='https://jira.company.com' -jira-token='' -list-config|1"
    "Empty jira jira username|punchout -config-file-path=config-good.toml -db-path='db.db' -jira-url='https://jira.company.com' -jira-token='XXX' -jira-installation-type 'cloud' -jira-username='' -list-config|1"
    "Incorrect value for time delta|punchout -config-file-path=config-good.toml -db-path='db.db' -jira-url='https://jira.company.com' -jira-token='XXX' -jira-time-delta-mins='blah' -list-config|1"
    "Incorrect fallback comment|punchout -config-file-path=config-good.toml -db-path='db.db' -fallback-comment='  ' -jira-url='https://jira.company.com' -jira-token='XXX' -list-config|1"
)

for test in "${tests[@]}"; do
    IFS='|' read -r title cmd expected_exit_code <<<"$test"

    echo "> $title"
    echo "$cmd"
    echo
    eval "$cmd"
    exit_code=$?
    if [ $exit_code -eq $expected_exit_code ]; then
        echo "✅ command behaves as expected"
        ((pass_count++))
    else
        echo "❌ command returned $exit_code, expected $expected_exit_code"
        ((fail_count++))
    fi
    echo
    echo "==============================="
    echo
done

echo "Summary:"
echo "- Passed: $pass_count"
echo "- Failed: $fail_count"

if [ $fail_count -gt 0 ]; then
    exit 1
else
    exit 0
fi
