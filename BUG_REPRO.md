# BUG_REPRO

## Bug

An acknowledged event does not retain a consistent `closed` terminal state. The closed timestamp can be shared across clones, closed events disappear from the closed-state filter, and the memory repository restores the previous acknowledged state during update.

## Trigger

Run the focused lifecycle, clone-isolation, closed-filter, and memory-persistence tests against the initial bug snapshot.

## Error

```text
--- FAIL: TestEventLifecycleTransitionP01 (0.00s)
    lifecycle_grading_test.go:19: close did not reach a consistent terminal state: Status:"acknowledged"
--- FAIL: TestClosedEventCloneTimestampIsolationP01 (0.00s)
    lifecycle_grading_test.go:33: clone shares ClosedAt pointer
--- FAIL: TestClosedEventFilterP01 (0.00s)
    lifecycle_grading_test.go:41: closed event disappeared from the closed-state filter
--- FAIL: TestClosedEventPersistenceP01 (0.00s)
    event_lifecycle_grading_test.go:28: closed transition was not persisted: Status:"acknowledged"
```
