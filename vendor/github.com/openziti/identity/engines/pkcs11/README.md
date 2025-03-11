PKCS#11 engine
==============

# Testing
Tests in this package use [Software HSM](https://github.com/opendnssec/SoftHSMv2) for testing. 
The tests will skip if softhsm2 driver is not found. Install `softhsm2` if you want to run these tests.
Directory `softhsm-testdata` contains predefined token with keys expected by the tests.

If the driver is installed but not being detected you can set the `SOFTHSM2_LIB` environment variable to
inform the test where to find the library:

    # linux
    $ export SOFTHSM2_LIB=/usr/local/lib/softhsm/libsofthsm2.so
    
    # windows
    SET SOFTHSM2_LIB=V:/dev/tools/softhsm2/lib/softhsm2-x64.dll