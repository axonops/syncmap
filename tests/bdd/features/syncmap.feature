Feature: SyncMap public API

  Rule: Load

    Scenario: Present key returns value and found true
      Given an empty SyncMap of string to int
      And the key "hello" has been stored with value 42
      When I Load key "hello"
      Then the returned value is 42
      And the found flag is true

    Scenario: Missing key returns zero value and found false
      Given an empty SyncMap of string to int
      When I Load key "missing"
      Then the returned value is the zero value
      And the found flag is false

    Scenario: Zero value stored is distinguished from missing
      Given an empty SyncMap of string to int
      And the key "z" has been stored with value 0
      When I Load key "z"
      Then the returned value is 0
      And the found flag is true

  Rule: Store

    Scenario: Overwrite replaces value
      Given an empty SyncMap of string to int
      And the key "k" has been stored with value 1
      When I Store key "k" with value 2
      And I Load key "k"
      Then the returned value is 2
      And the found flag is true

    Scenario: Empty string key is valid
      Given an empty SyncMap of string to int
      When I Store key "" with value 99
      And I Load key ""
      Then the returned value is 99
      And the found flag is true

  Rule: LoadOrStore

    Scenario: Absent key stores value and returns it with loaded false
      Given an empty SyncMap of string to int
      When I LoadOrStore key "new" with value 10
      Then the returned value is 10
      And the loaded flag is false
      And the map contains key "new" with value 10

    Scenario: Present key loads existing value with loaded true
      Given an empty SyncMap of string to int
      And the key "existing" has been stored with value 55
      When I LoadOrStore key "existing" with value 0
      Then the returned value is 55
      And the loaded flag is true

    Scenario: Zero value stored is loadable by LoadOrStore
      Given an empty SyncMap of string to int
      And the key "z" has been stored with value 0
      When I LoadOrStore key "z" with value 7
      Then the returned value is 0
      And the loaded flag is true

  Rule: LoadAndDelete

    Scenario: Present key returns value and removes entry
      Given an empty SyncMap of string to int
      And the key "target" has been stored with value 77
      When I LoadAndDelete key "target"
      Then the returned value is 77
      And the found flag is true
      And the map does not contain key "target"

    Scenario: Missing key returns zero value and loaded false
      Given an empty SyncMap of string to int
      When I LoadAndDelete key "absent"
      Then the returned value is the zero value
      And the found flag is false

    Scenario: Zero value stored is distinguished from missing by loaded flag
      Given an empty SyncMap of string to int
      And the key "z" has been stored with value 0
      When I LoadAndDelete key "z"
      Then the returned value is 0
      And the found flag is true
      And the map does not contain key "z"

  Rule: Delete

    Scenario: Existing key is removed
      Given an empty SyncMap of string to int
      And the key "del" has been stored with value 5
      When I Delete key "del"
      Then the map does not contain key "del"

    Scenario: Deleting a missing key is a no-op
      Given an empty SyncMap of string to int
      When I Delete key "never-existed"
      Then no panic occurs
      And the map does not contain key "never-existed"

    Scenario: Double delete is a no-op
      Given an empty SyncMap of string to int
      And the key "d" has been stored with value 1
      When I Delete key "d"
      And I Delete key "d"
      Then no panic occurs
      And the map does not contain key "d"

  Rule: Swap

    Scenario: Absent key stores value and returns zero value with loaded false
      Given an empty SyncMap of string to int
      When I Swap key "new" with value 88
      Then the returned value is the zero value
      And the found flag is false
      And the map contains key "new" with value 88

    Scenario: Present key returns old value and overwrites with loaded true
      Given an empty SyncMap of string to int
      And the key "s" has been stored with value 10
      When I Swap key "s" with value 20
      Then the returned value is 10
      And the found flag is true
      And the map contains key "s" with value 20

    Scenario: Zero V is distinguished from absent by loaded flag
      Given an empty SyncMap of string to int
      And the key "z" has been stored with value 0
      When I Swap key "z" with value 1
      Then the returned value is 0
      And the found flag is true

  Rule: Clear

    Scenario: Clear on an empty map is a no-op
      Given an empty SyncMap of string to int
      When I Clear the map
      Then no panic occurs
      And Len equals 0

    Scenario: Clear on a populated map leaves it empty
      Given an empty SyncMap of string to int
      And the map contains the following entries
        | key | value |
        | a   | 1     |
        | b   | 2     |
        | c   | 3     |
      When I Clear the map
      Then Len equals 0
      And the map does not contain key "a"
      And the map does not contain key "b"
      And the map does not contain key "c"

  Rule: Range

    Scenario: Range visits every entry
      Given an empty SyncMap of string to int
      And the map contains the following entries
        | key | value |
        | x   | 10    |
        | y   | 20    |
        | z   | 30    |
      When I Range all entries
      Then Range visited exactly 3 entries

    Scenario: Early return stops iteration
      Given an empty SyncMap of string to int
      And the map contains the following entries
        | key | value |
        | a   | 1     |
        | b   | 2     |
        | c   | 3     |
        | d   | 4     |
        | e   | 5     |
      When I Range and stop after 2 entries
      Then Range visited exactly 2 entries

    Scenario: Empty map invokes callback zero times
      Given an empty SyncMap of string to int
      When I Range all entries
      Then Range visited exactly 0 entries

  Rule: Len

    Scenario: Empty map has Len of zero
      Given an empty SyncMap of string to int
      When I request Len
      Then Len returns 0

    Scenario: Len equals number of stored entries
      Given an empty SyncMap of string to int
      And the map contains the following entries
        | key | value |
        | p   | 1     |
        | q   | 2     |
        | r   | 3     |
      When I request Len
      Then Len returns 3

  Rule: Map

    Scenario: Snapshot matches stored entries
      Given an empty SyncMap of string to int
      And the map contains the following entries
        | key | value |
        | m   | 100   |
        | n   | 200   |
      When I request Map
      Then the snapshot length equals 2
      And the snapshot contains key "m" with value 100
      And the snapshot contains key "n" with value 200

  Rule: Keys

    Scenario: Keys matches stored keys
      Given an empty SyncMap of string to int
      And the map contains the following entries
        | key | value |
        | alpha | 1   |
        | beta  | 2   |
        | gamma | 3   |
      When I request Keys
      Then the captured keys length equals 3
      And the captured keys contain "alpha"
      And the captured keys contain "beta"
      And the captured keys contain "gamma"

    Scenario: Empty map returns empty keys slice
      Given an empty SyncMap of string to int
      When I request Keys
      Then the captured keys length equals 0

  Rule: Values

    Scenario: Values matches stored values
      Given an empty SyncMap of string to int
      And the map contains the following entries
        | key | value |
        | a   | 11    |
        | b   | 22    |
        | c   | 33    |
      When I request Values
      Then the captured values length equals 3
      And the captured values contain 11
      And the captured values contain 22
      And the captured values contain 33

    Scenario: Empty map returns empty values slice
      Given an empty SyncMap of string to int
      When I request Values
      Then the captured values length equals 0

  Rule: CompareAndSwap

    Scenario: Matching old value swaps and returns true
      Given an empty SyncMap of string to int
      And the key "k" has been stored with value 10
      When I CompareAndSwap key "k" from 10 to 20
      Then the swapped flag is true
      And the map contains key "k" with value 20

    Scenario: Mismatched old value does not swap and returns false
      Given an empty SyncMap of string to int
      And the key "k" has been stored with value 10
      When I CompareAndSwap key "k" from 99 to 20
      Then the swapped flag is false
      And the map contains key "k" with value 10

    Scenario: Missing key returns false and does not store
      Given an empty SyncMap of string to int
      When I CompareAndSwap key "absent" from 0 to 1
      Then the swapped flag is false
      And the map does not contain key "absent"

    Scenario: Zero V match swaps successfully
      Given an empty SyncMap of string to int
      And the key "z" has been stored with value 0
      When I CompareAndSwap key "z" from 0 to 1
      Then the swapped flag is true
      And the map contains key "z" with value 1

  Rule: CompareAndDelete

    Scenario: Matching value deletes entry and returns true
      Given an empty SyncMap of string to int
      And the key "k" has been stored with value 10
      When I CompareAndDelete key "k" expecting 10
      Then the deleted flag is true
      And the map does not contain key "k"

    Scenario: Mismatched value does not delete and returns false
      Given an empty SyncMap of string to int
      And the key "k" has been stored with value 10
      When I CompareAndDelete key "k" expecting 99
      Then the deleted flag is false
      And the map contains key "k" with value 10

    Scenario: Missing key returns false
      Given an empty SyncMap of string to int
      When I CompareAndDelete key "absent" expecting 0
      Then the deleted flag is false

  Rule: Concurrency

    Scenario: Multiple goroutines storing disjoint keys yield correct total count
      Given an empty SyncMap of string to int
      When 4 goroutines each Store 25 keys
      Then Len equals 100
