package pfxlog

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/mgutz/ansi"
	"io"
	"strings"
	"time"
)

func Filter(sourceR io.Reader, options *Options) {
	r := bufio.NewReader(sourceR)
	var last time.Time
	lastSet := false
	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		line = autocut(line)
		msg := make(map[string]interface{})
		err = json.Unmarshal([]byte(line), &msg)
		if err != nil {
			fmt.Println(ansi.Yellow + strings.TrimSpace(line) + ansi.DefaultFG)
			continue
		}
		stamp, err := time.Parse(time.RFC3339, msg["time"].(string))
		if err != nil {
			panic(err)
		}
		if !lastSet {
			last = stamp
			lastSet = true
		}
		delta := stamp.Sub(last).Seconds()
		var level string
		switch msg["level"].(string) {
		case "panic":
			level = options.PanicLabel
		case "fatal":
			level = options.FatalLabel
		case "error":
			level = options.ErrorLabel
		case "warning":
			level = options.WarningLabel
		case "info":
			level = options.InfoLabel
		case "debug":
			level = options.DebugLabel
		case "trace":
			level = options.TraceLabel
		default:
			panic(fmt.Errorf("unknown (%s)", msg["level"].(string)))
		}
		var prefix string
		if v, found := msg["func"]; found {
			prefix = strings.TrimPrefix(v.(string), options.TrimPrefix)
		}
		if context, found := msg["_context"]; found {
			prefix += " [" + context.(string) + "]"
		}
		message := msg["msg"].(string)
		data := data(msg)
		if len(data) > 0 {
			fields := "{"
			field := 0
			for k, v := range data {
				if field > 0 {
					fields += " "
				}
				field++
				fields += fmt.Sprintf("%s=[%v]", k, v)
			}
			fields += "} "
			message = options.FieldsColor + fields + options.DefaultFgColor + message
		}
		var fmtTs string
		if options.AbsoluteTime {
			fmtTs = fmt.Sprintf("[%s]", stamp.Format(options.PrettyTimestampFormat))
		} else {
			fmtTs = fmt.Sprintf("[%8.3f]", delta)
		}
		fmt.Printf("%s %s %s: %s\n",
			options.TimestampColor+fmtTs+options.DefaultFgColor,
			level,
			options.FunctionColor+prefix+options.DefaultFgColor,
			message)
	}
}

func data(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range in {
		if k != "level" && k != "func" && k != "file" && k != "msg" && k != "time" && k != "_context" {
			out[k] = v
		}
	}
	return out
}

func autocut(inputLine string) string {
	idx := strings.IndexRune(inputLine, '{')
	if idx > -1 {
		return inputLine[idx:]
	}
	return inputLine
}
