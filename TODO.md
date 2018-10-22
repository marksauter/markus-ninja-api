Better response from /confirm_verification route
fix update queries to eliminate identical row updates that can cause various
  side effects

finish updating data types to reflect new filters - see lesson/study for example

fix order fields for each data type, make sure they match the schema

THERE IS PROBABLY SOME SORT OF MEMORY LEAK, BECAUSE THE SERVER SLOWS DOWN THE
LONGER IT RUNS...
  pgx prepare statements seems to be the main culprit, need to figure out how to
  use it appropriately...

Fields that can be used for ordering, cannot not be null.
  - e.g. study.advanced_at cannot be null, need to change this. 
