## Setting up ZMQ to build on Windows

0. Ensure that you've followed the steps leading up to this on the [README.md](README.md). Clone this repository into the `%GOPATH%\src\github.com\edgexfoundry\` directory, using `git clone https://github.com/edgexfoundry/edgex-go.git %GOPATH%\src\github.com\edgexfoundry\edgex-go`. If `dep` is available (it should be since it's a prerequisite), navigate to `%GOPATH%\src\github.com\edgexfoundry\edgex-go` and run `dep ensure -v`. Otherwise, some of the below `go get` steps will need to be run before attempting to build (skippable commands will be denoted).

1. Download & install the TDM-GCC from this URL: http://tdm-gcc.tdragon.net/download

2. Run `go get -v -x github.com/pebbe/zmq4`, **expect it to fail** due to a missing `zmq.h` file. This will create the directory structure needed in order to accomplish the steps below. This step can be skipped, but if skipped the directories in the following steps will have to be manually created.

3. Download the release version of ZMQ from AppVeyor builds from zmq: https://ci.appveyor.com/project/zeromq/libzmq. Be sure it is a successful build and be sure a choose the the correct platform architecture and VERSION for your system.  Ensure the commit that is built matches the version you clone from source in step 3. Example: `Environment: platform=x64, configuration=Release, WITH_LIBSODIUM=ON, ENABLE_CURVE=ON, NO_PR=TRUE`. You can download the zip from the artifacts tab.

    3.a. Extract the contents of the downloaded zip file anywhere you prefer.

    3.b. Copy the `libsodium.dll` and `libzmq-v120-mt-4_x_x.dll` files from step 3 into `%GOPATH%\src\github.com\pebbe\zmq4\usr\local\lib`; create all of the directories if you need to. Rename `libzmq-v120-mt-4_x_x.dll` to `libzmq.dll`, so that gcc can link and find `-lzmq`. If this is not done, errors like this may occur: `C:/TDM-GCC-64/bin/../lib/gcc/x86_64-w64-mingw32/5.1.0/../../../../x86_64-w64-mingw32/bin/ld.exe:cannot find -lzmq`

3. In any directory you prefer, clone the source of ZMQ that matches the version/commit you downloaded from AppVeyor: `git clone https://github.com/zeromq/libzmq.git`.

4. Copy over the `include` directory from the source downloaded in step 3 to  `%GOPATH%\src\github.com\pebbe\zmq4\usr\local\include`, ensuring that any necessary directories that don't exist are created.

5. Set the following environment variables (run this in PowerShell):

```
$Env:CGO_CFLAGS="-I%GOPATH%\src\github.com\pebbe\zmq4\usr\local\include"
$Env:CGO_LDFLAGS="-L%GOPATH%\src\github.com\pebbe\zmq4\usr\local\lib"
```
> Replace %GOPATH% with your actual path as PowerShell doesn't seem to be able resolve %GOPATH%. For example C:\users\\[windowsusername]\go

After setting those, running `go get -v -x github.com/pebbe/zmq4` should succeed. If it fails, it most likely due to the `.h` files or `dll`s existing in a directory that doesnt match your `CGO_` environment variables. It may also fail due to the same error as before (see step 3.b) - try downloading the artifact from a different build on the [AppVeyor builds site](https://ci.appveyor.com/project/zeromq/libzmq).

5. Next, navigate to `%GOPATH%\github.com\edgexfoundry\edgex-go\cmd\core-data`. *If you didn't run `dep ensure -v` previously, run `go get -v` to acquire all dependencies*. Run `go build -v`. If it complains about not finding zmq.h from auth.go, double check the CGO_ environment variables by running `go env`.

6. There should now be a `core-data.exe` file in your directory. The last step is to ensure that a copy of the DLL's `libsodium.dll` and `libzmq-v120-mt-4_x_x.dll` (yes, this one, not with the rename we did earlier) is in the same directory as your `.exe`.

7. Run `.\core-data.exe`.

> If you get a segfault, try setting `$Env:GOCACHE="off"` and rerun the build command from step 5