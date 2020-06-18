#!/bin/bash
#
# Code coverage generation
CVG_EXCLUDE="${CVG_EXCLUDE:-test}"

go test -coverpkg=./... -coverprofile ./coverage.cov $(go list ./... |  grep -Ev "${CVG_EXCLUDE}")

# Display the global code coverage.  This generates the total number the badge uses
go tool cover -func=coverage.cov ;

# If needed, generate HTML report
if [ "$1" == "html" ]; then
    go tool cover -html=coverage.cov -o coverage.html ;
fi



