#!/bin/sh -ex

export AWS_DEFAULT_REGION=us-east-1
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test

if [ -f /status/localstack/1 ]; then
    echo "Skipped localstack create role"
else
    aws --endpoint-url=http://localstack:4566 iam create-role \
        --role-name "vui-iam-role" \
        --assume-role-policy-document '{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Principal": {"AWS": "arn:aws:iam::000000000000:root"}, "Action": "sts:AssumeRole"}]}'
    touch /status/localstack/1
fi

if [ -f /status/localstack/2 ]; then
    echo "Skipped localstack attach policy"
else
    aws --endpoint-url=http://localstack:4566 iam attach-role-policy \
        --role-name "vui-iam-role" \
        --policy-arn "arn:aws:iam::aws:policy/IAMFullAccess"
    touch /status/localstack/2
fi

echo "Completed"
touch /status/localstack/completed
