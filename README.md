# postmaster

The postmaster is responsible for queuing and sending emails. It currently uses
okq as a backing store and new emails are push'd into the end of the queue.
The postmaster exposes an API for queuing emails and internally also reads from
the backing store to send emails using SendGrid. Optionally Mongo can be used
to track statistics (received, open, etc) for each email. Additionally you can
store email preferences for each email address and emails will be silently
dropped if that email is opted out of that particular email.

## Sending

SendGrid is used to send emails and they're sent at most once. A `--sendgrid-key` is required.

## API

All requests against the API use JSON RPC 2.0. They must all be HTTP POSTs with a `Content-Type` of `application/json`.
The path of the request doesn't matter.

### Postmaster.Enqueue

An email must consist of a `to`, `from`, `html` (or `text`), and a `subject`.
Optional fields include `toName` and `fromName`. An `uniqueArgs` object can be
sent if you want to include optional SMTP headers. Finally, `flags` are used
to categorize the email and should be a bitwise number consisting of which
types of email this is.

**The first bit is reserved and should never be sent with ANY email**

Any email ending with `@test` will result in an error and can be used for unit
tests without risking sending an actual email.

Params:
```json
{
    "to": "test@test",
    "from": "test@test",
    "subject": "Test",
    "text": "Yo"
}
```

Returns:
```json
{
    "success": true
}
```

### Postmaster.UpdatePrefs

Store the bit types of emails to reject. Each bit sent will cause an email to
be silently dropped if it contains that bit in its `flags`.

If someone opts into ALL emails, send a `flags` value of 1, which no email
should contain. A `flags` value of 0 will result in an error.

Params:
```json
{
    "email": "test@test.com",
    "flags": 1
}
```

Returns:
```json
{
    "success": true
}
```

## Todo

* Rate-limit email addresses in order to prevent sending too many emails to the
same address too quickly
* Hash each email contents and email address to prevent sending the same email
to the same address twice
