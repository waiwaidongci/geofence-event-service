# Bug Reproduction

## Bug

Webhook delivery state does not fully transition across failure and retry, and subscription filters retain caller-owned slices and maps. A successful retry can keep an old error or timestamp while later caller mutations silently change an existing subscription.

## Trigger

Mark one delivery failed and then sent, and reuse or mutate the slices and map passed to a subscription after construction or cloning.

## Observed Error

The delivery record reports stale failure state after success, and the subscription begins matching a different set of events. The targeted regression tests fail on retry state cleanup and filter snapshot isolation.
