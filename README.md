# SMTP HOOK service

## Dev run

```bash
cp .env.example .env
# add variables to env, then
docker-compose up
```

## Send an email from localhost

```bash
swaks --to mail@domain.com --from anymain@gmail.com --server localhost:1025 -header "Subject: Test letter" --body "Sample email body in plaintext"
```

This will send the letter to localhost with smtp-hook and then to a webhook, that you provided in `.env` file.
