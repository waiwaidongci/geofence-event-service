# Bug Reproduction

## Bug

Location attributes and geofence transition snapshots share mutable maps and pointers. Reusing an attributes map for a later report can change an earlier stored location and the state used for enter, dwell, and exit decisions.

## Trigger

Create a location with an attributes map, retain it in a transition state, then mutate and reuse the original map for another report. Read the first location and the transition's entered and last-location snapshots afterward.

## Observed Error

The earlier snapshots contain values from the later report. The targeted regression tests fail because constructor, clone, entered-state, and last-location snapshots are not isolated.
