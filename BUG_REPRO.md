# Bug Reproduction

## Bug

Webhook delivery drops the caller context and transport error chain, does not reliably close response bodies, and records delivery failures as successful terminal state.

## Trigger

Send a webhook through a cancelable request context while the transport returns an error, a non-success response, or body read and close failures. Inspect the persisted delivery state after the call.

## Observed Error

Cancellation does not stop the request, errors cannot be matched through `errors.Is`, response resources remain open, and failed delivery attempts can be stored as sent.
