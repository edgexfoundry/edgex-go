/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package debugz

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func GenerateStack() string {
	return generateStack(1024*1024, true)
}

func GenerateLocalStack() string {
	return generateStack(8*1024, false)
}

func generateStack(size int, all bool) string {
	stackBuf := make([]byte, size)
	size = runtime.Stack(stackBuf, all)
	for size == len(stackBuf) {
		size = len(stackBuf) * 2
		stackBuf = make([]byte, size)
		size = runtime.Stack(stackBuf, all)
	}
	return string(stackBuf[:size])
}

func DumpStack() {
	fmt.Println(GenerateStack())
}

func DumpLocalStack() {
	fmt.Println(GenerateLocalStack())
}

func AddStackDumpHandler() {
	go func() {
		signalC := make(chan os.Signal, 1)
		signal.Notify(signalC, syscall.SIGQUIT)
		for range signalC {
			fmt.Printf("\n DUMPING STACK AS REQUESTED BY SIGQUIT \n\n%v\n", GenerateStack())
		}
	}()
}

func DumpStackOnTick(tickerChan <-chan time.Time, fileFormatter func(time.Time) string) {
	go func() {
		for t := range tickerChan {
			if err := DumpStackToFile(fileFormatter(t)); err != nil {
				fmt.Printf("error dumping stackdump to file [%v]\n", err)
			}
		}
	}()
}

func DumpStackToFile(fileName string) error {
	out, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer func() { _ = out.Close() }()

	stackDump := GenerateStack()
	_, err = out.WriteString(stackDump)
	if err != nil {
		return err
	}
	return nil
}
