#!/bin/sh
set -x

# Build documentation

cd /docbuild
mkdir _static
make html

# Add documentation for service interfaces

raml2html core-data.raml -o _build/html/core-data.html
raml2html core-metadata.raml -o _build/html/core-metadata.html
raml2html core-command.raml -o _build/html/core-command.html
raml2html support-logging.raml -o _build/html/support-logging.html
raml2html support-notifications/raml/support-notifications.raml -o _build/html/support-notifications.html
raml2html device-virtual/raml/device-virtual.raml -o _build/html/device-virtual.html
raml2html support-rulesengine/raml/support-rulesengine.raml -o _build/html/support-rulesengine.html
raml2html export-client.raml -o _build/html/export-client.html

# Check for broken links in HTML

cd _build/html
linkchecker index.html
