
# magusd submission

## Install the operator

### Option 1 - pre-built container (linux/arm64)

```bash
make
make install
make deploy
```

### Option 2 - build/deploy

```bash
make
make install
make docker-build IMG="youruser/image:tag"
make docker-push IMG="youruser/image:tag"
make deploy
```

## Create the secret and senderconfig

### Option 1 kustomize

```bash
cp config/samples/samples-secret.yaml config/samples/secret.yaml
vim config/samples/secret.yaml # fill in the values with base64 encoded text
kubectl apply -k config/samples/
```

### Option 2 manifests

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