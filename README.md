
# magusd submission

## Install the operator

```bash
make
make install

#to use the pre-built image docker.io/unclelobs/mailerlite-controller
make deploy

#to build your own
make docker-build IMG="youruser/image:tag"
make docker-push IMG="youruser/image:tag"

make deploy
```

## Create the secret and senderconfig

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: mailgun-token
data:
  provider:  mailgun
  domain: $(echo -n "$DOMAIN" | base64)
  token: $(echo -n "$TOKEN" | base64)
EOF
```

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: k8s.magusd.com.my.domain/v1
kind: EmailSenderConfig
metadata:
  name: mailgun
spec:
  apiTokenSecretRef: mailgun-token
  senderEmail: ${SENDER_EMAIL}
EOF
```

## Send an email with mailgun

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: k8s.magusd.com.my.domain/v1
kind: Email
metadata:
  name: email-sample
spec:
  senderConfigRef: mailgun
  recipientEmail: recipient@gmail.com
  subject: hello
  body: world mailgun!
EOF
```