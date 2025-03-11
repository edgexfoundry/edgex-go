echo off
REM #	Copyright 2019 NetFoundry, Inc.
REM #
REM #	Licensed under the Apache License, Version 2.0 (the "License");
REM #	you may not use this file except in compliance with the License.
REM #	You may obtain a copy of the License at
REM #
REM #	https://www.apache.org/licenses/LICENSE-2.0
REM #
REM #	Unless required by applicable law or agreed to in writing, software
REM #	distributed under the License is distributed on an "AS IS" BASIS,
REM #	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
REM #	See the License for the specific language governing permissions and
REM #	limitations under the License.

SET lib=%1
SET pin=%2

rmdir /q /s softhsm-testdata
mkdir softhsm-testdata

SET SOFTHSM2_CONF=softhsm2.conf

echo on
softhsm2-util --init-token --free --label 'ziti-test-token' --so-pin %pin% --pin %pin%
"c:\Program Files\OpenSC Project\OpenSC\tools\pkcs11-tool.exe" --module %lib% -p %pin% -k --key-type rsa:2048 --id 01 --label ziti-rsa-key
"c:\Program Files\OpenSC Project\OpenSC\tools\pkcs11-tool.exe" --module %lib% -p %pin% -k --key-type EC:prime256v1 --id 02 --label ziti-ecdsa-key
