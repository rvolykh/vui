FROM hashicorp/vault:1.20

RUN apk update && \
    apk add --no-cache aws-cli kubectl

COPY scripts/ /vui/
COPY cfg/aws /root/.aws/config

ENV VAULT_ADDR=http://vault:8200 \
    VAULT_TOKEN=vui-sandbox-token \
    AWS_ACCESS_KEY_ID=test \
    AWS_SECRET_ACCESS_KEY=test \
    AWS_SESSION_TOKEN=test

ENTRYPOINT ["Bad usage"]
