## Setting up ZMQ to build on Windows

> Last updated Nov 2019 to be reworked for module directory structure.

0. Ensure that you've followed the steps leading up to this on the [README.md](README.md), this includes any dependenices. Clone this repository into the `%GOPATH%\src\github.com\edgexfoundry\` directory, using `git clone https://github.com/edgexfoundry/edgex-go.git %GOPATH%\src\github.com\edgexfoundry\edgex-go`.

1. Download & install the TDM-GCC from this URL: http://tdm-gcc.tdragon.net/download

2. Run `go get -v -x github.com/pebbe/zmq4`, **expect it to fail** due to a missing `zmq.h` file. This will create the directory structure needed in order to accomplish the steps below. This step can be skipped, but if skipped the directories in the following steps will have to be manually created.

3. Download the release version of ZMQ from AppVeyor builds from zmq: https://ci.appveyor.com/project/zeromq/libzmq. Verify it is a successful build and ensure  the platform architecture and VERSION is correct for your system.  Ensure the commit that is built matches the version you clone from source in step 3. Example: `Environment: platform=x64, configuration=Release, WITH_LIBSODIUM=ON, ENABLE_CURVE=ON, NO_PR=TRUE`. You can download the zip from the artifacts tab.

    3.a. Extract the contents of the downloaded zip file anywhere you prefer.

    3.b. Copy the `libsodium.dll` and `libzmq-v120-mt-4_x_x.dll` files from step 3 into `%GOPATH%\pkg\mod\github.com\pebbe\zmq4@v1.0.0\lib`; create the lib directory as needed. Rename `libzmq-v120-mt-4_x_x.dll` to `libzmq.dll`, so that gcc can link and find `-lzmq`. If this is not done, errors like this may occur: `C:/TDM-GCC-64/bin/../lib/gcc/x86_64-w64-mingw32/5.1.0/../../../../x86_64-w64-mingw32/bin/ld.exe:cannot find -lzmq`

> Note: As of Nov 2019, ZMQ 4.3.3 does not work and results in an error: https://github.com/pebbe/zmq4/issues/152/. 4.3.2 is recommended and tested on Win 10:
https://ci.appveyor.com/project/zeromq/libzmq/builds/25882500/job/s2yhbnjrdouq5qnm/artifacts
and
https://github.com/zeromq/libzmq/tree/a84ffa12b2eb3569ced199660bac5ad128bff1f0. 

4. In any directory you prefer, clone the source of ZMQ `git clone https://github.com/zeromq/libzmq.git` and ensure it matches the version/commit you downloaded from AppVeyor (i.e `git reset --hard a84ffa12`).

5. Copy over the `include` directory from the source downloaded in step 3 to  `%GOPATH%\pkg\mod\github.com\pebbe\zmq4@v1.0.0\`.

6. Set the following environment variables (run this in PowerShell):

```
$Env:CGO_CFLAGS="-I$Env:GOPATH\pkg\mod\github.com\pebbe\zmq4@v1.0.0\include"
$Env:CGO_LDFLAGS="-L$Env:GOPATH\pkg\mod\github.com\pebbe\zmq4@v1.0.0\lib"
```

After setting those, running `go get -v -x github.com/pebbe/zmq4` should succeed. If it fails, it most likely due to the `.h` files or `dll`s existing in a directory that doesnt match your `CGO_` environment variables. It may also fail due to the same error as before (see step 3.b) - try downloading the artifact from a different build on the [AppVeyor builds site](https://ci.appveyor.com/project/zeromq/libzmq).

> It is recommended to set these environment variables in your System or IDE so that you do not need to set them in every terminal session. This especially required if you wish to debug services that depend on ZMQ. 

7. Next, navigate to `$Env:GOPATH\src\github.com\edgexfoundry\edgex-go\cmd\core-data`. *If you didn't acquire dependencies previously, run `go get -v` to acquire all dependencies*. Run `go build -v`. If it complains about not finding zmq.h from auth.go, double check the CGO_ environment variables by running `go env`.

8. There should now be a `core-data.exe` file in your directory. The last step is to ensure that a copy of the DLL's `libsodium.dll` and `libzmq-v120-mt-4_x_x.dll` (yes, this one, not with the rename we did earlier) is in the same directory as your `.exe`. You will need to copy these .dlls in manually. 

9. Run `.\core-data.exe`.

> If you get a segfault, try setting `$Env:GOCACHE="off"` and rerun the build command from step 5