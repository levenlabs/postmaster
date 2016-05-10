# postmaster

The postmaster is responsible for queuing and sending emails. It currently uses
okq as a backing store and new emails are push'd into the end of the queue.
The postmaster exposes an API for queuing emails and internally also reads from
the backing store to send emails using SendGrid. Optionally Mongo can be used
to track statistics (received, open, etc) for each email. Additionally you can
store email preferences for each email address and emails will be silently
dropped if that email is opted out of that particular email.

## Prerequisites

You must have a SendGrid account and pass your key via `--sendgrid-key`.
Optionally, in order to store statistics you must be running a MongoDB instance
and send the address to `--mongo-addr`. You must also publicly expose the
postmaster webhook port to the Internet. Do NOT expose the RPC port.

In order to provide resiliency against the service crashing before it had a
chance to process a webhook or an email, or to run multiple instances of
postmaster, an instance of [okq](https://github.com/mc0/okq) can be run and
passed as `--okq-addr`. All jobs will be held in okq until they are processed.

## Version

The running postmaster with `--version` prints the version number. This is only
supported in releases unless you set the version when compiling:
```
go build -ldflags "-X github.com/levenlabs/golib/genapi.Version 'myversion'" ./...
```

## Webhook

The webhook port can be specified via `--webhook-addr`. In
order to minimally protect the webhook endpoint, you can specify a basic auth
password with `--webhook-pass` that will be required for all webhook requests.
Currently postmaster supports the following actions:

* Delivered (flag 2)
* Mark as Spam (flag 4)
* Bounced (flag 8)
* Dropped (flag 16)
* Opened (flag 32)

Those flags will be bitwise or'd together as `StateFlags`.

## API

All requests against the API use JSON RPC 2.0. They must all be HTTP POSTs with
a `Content-Type` of `application/json`. The path of the request doesn't matter.

### Postmaster.Enqueue

An email must consist of a `to`, `from`, `html` (or `text`), and a `subject`.
Optional fields include `toName` and `fromName`. An `uniqueArgs` object can be
sent if you want to include optional SMTP headers. Finally, `flags` are used
to categorize the email and should be a bitwise number consisting of which
types of email this is.

**The first bit is reserved and should never be sent with ANY email**

Any email ending with `@test` will result in an error and can be used for unit
tests without risking sending an actual email.

Optionally, an `uniqueID` can be sent with the email and will be stored with the
stats for the email. You can use this in conjunction with
`Postmaster.GetLastEmail` to verify that you didn't already send an email to a
user within a certain threshold of time.

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
    "email": "test@test",
    "flags": 1
}
```

Returns:
```json
{
    "success": true
}
```

### Postmaster.GetLastEmail

Get the last email sent to `to` with the `uniqueID`. You must be running with
statistics and pass a `uniqueID` in Enqueue in order to use this functionality.

Params:
```json
{
    "to": "test@test",
    "uniqueID": "user_15_favorited_user_12"
}
```

Returns:
```json
{
    "stat": {
        "recipient": "test@test.com",
        "emailFlags": 0,
        "stateFlags": 32,
        "uniqueID": "user_15_favorited_user_12",
        "tsCreated": 1449264108,
        "tsUpdated": 1449264208,
        "error": ""
    }
}
```

Returns if non found:
```json
{
    "stat": null
}
```

