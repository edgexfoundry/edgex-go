## Setting up ZMQ to build on Windows

1. Download & install the TDM-GCC: [http://tdm-gcc.tdragon.net/download]

2. Download the release version of ZMQ from AppVeyor builds from zmq: https://ci.appveyor.com/project/zeromq/libzmq. Be sure it is a successful build and be sure a choose the the correct platform architecture and VERSION for your system.  Ensure the commit that is built matches the version you clone from source in step 3. Example: `Environment: platform=x64, configuration=Release, WITH_LIBSODIUM=ON, ENABLE_CURVE=ON, NO_PR=TRUE`. You can download the zip from the artifacts tab

    2.b. After extracting copy the `libsodium.dll` and `libzmq-v120-mt-4_x_x.dll` from step 2 into `%GOPATH%\src\github.com\pebbe\zmq4\usr\local\lib`. You'll need to to rename `libzmq-v120-mt-4_x_x.dll` to `libzmq.dll` so that gcc can link and find `-lzmq`. If you don't you'll get an error akin to : `C:/TDM-GCC-64/bin/../lib/gcc/x86_64-w64-mingw32/5.1.0/../../../../x86_64-w64-mingw32/bin/ld.exe:cannot find -lzmq`

3. Clone the source of ZMQ that matches the version/commit you downloaded from AppVeyor: https://github.com/zeromq/libzmq.git. 

4. Copy over the `include` directory from the source we downloaded in step 3 to  `%GOPATH%\src\github.com\pebbe\zmq4\usr\local\include` 

5. Set the following environment variables (powershell):

```
$Env:CGO_CFLAGS="-I%GOPATH%\src\github.com\pebbe\zmq4\usr\local\include"
$Env:CGO_LDFLAGS="-L%GOPATH%\src\github.com\pebbe\zmq4\usr\local\lib"
```
> Replace %GOPATH% with your actual path as `go build` doesn't seem to be able resolve %GOPATH%. For example C:\users\\[windowsusername]\go

After setting those, running `go get -v -x github.com/pebbe/zmq4` should succeed. If it fails, it most likely due to the .h files or dlls existing in a directory that doesnt match your CGO_ environment variables.

5. After thats done, the next step is running the `go build cmd/core-data`. If it complains about not finding zmq.h from auth.go, double check your CGO_ environment variables by running `go env`

6. You should now have `core-data.exe` in your directory. The last step is to ensure that you have a copy of the DLLS `libsodium.dll` and `libzmq-v120-mt-4_x_x.dll` (yes, this one, not with the rename we did earlier) in the same directory as your `.exe`. 

7. Run it!

>If you get a segfault, try setting `$Env:GOCACHE="off"` and rerun the build command from step 5