package tools

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"sync"
	"syscall"

	"github.com/go-faster/errors"
	jsoniter "github.com/json-iterator/go"
)

type (
	fieldsCollector struct {
		fields          map[string]string
		buf             bytes.Buffer
		m               sync.Mutex
		unmarshalErrors int
	}

	// PatternDetector special structure what can handle stdout/err and detect pattern in log
	PatternDetector struct {
		pattern string
		buf     bytes.Buffer
		m       sync.Mutex
		count   int
	}

	// BaseExecutor base interface for executor in order to create executors for not golang services
	BaseExecutor interface {
		Start() error
		Run() error
		Stop() error
		IsDebug() bool
	}

	// Executor helper to start binaries with log pattern detection
	Executor struct {
		stdoutDetector *PatternDetector
		stderrDetector *PatternDetector
		fieldsParser   *fieldsCollector
		cmd            *exec.Cmd
		outputFile     *os.File
		errorsFile     *os.File
		debug          bool
	}
)

const (
	FieldNameHeader   = "field name"
	TypeHeader        = "type"
	DescriptionHeader = "description"
	TrueValue         = "true"
)

var json = jsoniter.ConfigFastest

func diffMaps(map1, map2 map[string]string) (onlyInMap1, onlyInMap2, differentValues map[string]string) {
	onlyInMap1 = make(map[string]string)
	onlyInMap2 = make(map[string]string)
	differentValues = make(map[string]string)

	for key, val := range map1 {
		if val2, ok := map2[key]; !ok {
			onlyInMap1[key] = val
		} else if val != val2 {
			differentValues[key] = val
		}
	}

	for key, val := range map2 {
		if _, ok := map1[key]; !ok {
			onlyInMap2[key] = val
		}
	}

	return onlyInMap1, onlyInMap2, differentValues
}

func newFieldsCollector() *fieldsCollector {
	return &fieldsCollector{
		fields: make(map[string]string),
	}
}

func processHeaders(headers []string) (fieldNameIndex, typeIndex int, err error) {
	fieldNameIndex, typeIndex = -1, -1
	columnMap := make(map[string]int)
	for index, header := range headers {
		columnMap[header] = index
	}
	for _, e := range []string{FieldNameHeader, TypeHeader, DescriptionHeader} {
		_, ok := columnMap[e]
		if !ok {
			return 0, 0, fmt.Errorf("header %s is missing", e)
		}
		switch e {
		case FieldNameHeader:
			fieldNameIndex = columnMap[e]
		case TypeHeader:
			typeIndex = columnMap[e]
		}
	}

	return fieldNameIndex, typeIndex, nil
}

func printInformation(actualFields, onlyInfFile, differentTypes map[string]string) {
	if len(actualFields) != 0 {
		fmt.Println("=================================================")
		fmt.Println("Fields in log are missing in file. Please add them to file")
		fmt.Printf("%s,%s,%s\n", FieldNameHeader, TypeHeader, DescriptionHeader)
		for k, v := range actualFields {
			fmt.Printf("%s,%s,\n", k, v)
		}
		fmt.Println("=================================================")
	}

	if len(onlyInfFile) != 0 {
		fmt.Println("=================================================")
		fmt.Println("Fields in file are missing in log. Please remove them from file")
		for k, v := range onlyInfFile {
			fmt.Println(k, "=", v)
		}
		fmt.Println("=================================================")
	}

	if len(differentTypes) != 0 {
		fmt.Println("=================================================")
		fmt.Println("Fields with different types. Please fix type")
		for k, v := range differentTypes {
			fmt.Println(k, "=", v)
		}
		fmt.Println("=================================================")
	}
}

func trimSpacesInRecords(items []string) []string {
	result := make([]string, 0, len(items))
	for _, e := range items {
		result = append(result, strings.TrimSpace(e))
	}
	return result
}

func validateLogFields(fieldsCollectorFilePath string, fields map[string]string, unmarshalErrors int) error {
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	f, err := os.Open(fieldsCollectorFilePath)
	if err != nil {
		fmt.Println("please provide file with all log fields", fieldsCollectorFilePath)
		return err
	}
	defer f.Close()
	csvReader := csv.NewReader(f)
	fieldNameIndex, typeIndex := -1, -1
	fieldsInFile := make(map[string]string)

	for {
		records, readErr := csvReader.Read()
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return readErr
		}
		records = trimSpacesInRecords(records)
		if fieldNameIndex == -1 {
			var processErr error
			fieldNameIndex, typeIndex, processErr = processHeaders(records)
			if processErr != nil {
				return processErr
			}
			continue
		}
		fieldsInFile[records[fieldNameIndex]] = records[typeIndex]
	}
	var failed bool

	m1, m2, dm := diffMaps(fields, fieldsInFile)
	if len(m1) != 0 || len(m2) != 0 || len(dm) != 0 {
		printInformation(m1, m2, dm)
		failed = true
	}
	if unmarshalErrors != 0 {
		fmt.Println("found json unmarshal errors: ", unmarshalErrors)
		failed = true
	}

	if failed {
		return fmt.Errorf("validation of fileds failed, please check the output")
	}
	return nil
}

func (fc *fieldsCollector) Write(p []byte) (n int, err error) {
	fc.m.Lock()
	defer fc.m.Unlock()
	n, err = fc.buf.Write(p)

	for {
		l, err2 := fc.buf.ReadBytes('\n')
		if err2 == nil {
			var m map[string]interface{}
			err = json.Unmarshal(l, &m)
			if err != nil {
				fc.unmarshalErrors++
				fmt.Println("ERROR DURING UNMARSHALING OF LOG LINE: ", string(l), "ERROR: ", err)
				continue
			}

			for k, v := range m {
				_, ok := fc.fields[k]
				if ok {
					continue
				}
				fc.fields[k] = reflect.TypeOf(v).String()
			}
			continue
		}
		break
	}
	return n, err
}

// Write implementation Writer interface
func (pt *PatternDetector) Write(p []byte) (n int, err error) {
	pt.m.Lock()
	defer pt.m.Unlock()
	n, err = pt.buf.Write(p)

	for {
		l, err2 := pt.buf.ReadString('\n')
		if err2 == nil {
			if strings.Contains(l, pt.pattern) {
				pt.count++
			}
			continue
		}
		break
	}

	return n, err
}

// Start starts the binary but does not wait for it to complete.
func (b *Executor) Start() error {
	return b.cmd.Start()
}

// Run executes the binary and waits for it to complete.
func (b *Executor) Run() error {
	if err := b.cmd.Run(); err != nil {
		return err
	}
	if err := b.checkOutput(); err != nil {
		return err
	}
	fmt.Println("run done process", b.cmd.Path)
	return nil
}

func (b *Executor) IsDebug() bool {
	return b.debug
}

// Stop sends SIGTERM to the binary and waits for it to exit.
func (b *Executor) Stop() error {
	fmt.Println("sending signal to process during stop, process=", b.cmd.Path, b.cmd.Process.Pid)

	if err := b.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return err
	}

	fmt.Println("waiting for process", b.cmd.Path)

	if err := b.cmd.Wait(); err != nil {
		fmt.Println("failed to wait for process", err)
		return err
	}

	if err := b.checkOutput(); err != nil {
		fmt.Println("failed to check output", err)
		return err
	}

	if b.outputFile != nil {
		if err := b.outputFile.Close(); err != nil {
			fmt.Printf("failed to close output file: %v\n", err)
		}
	}

	if b.errorsFile != nil {
		if err := b.errorsFile.Close(); err != nil {
			fmt.Printf("failed to close errors file: %v\n", err)
		}
	}

	fmt.Println("stop done process", b.cmd.Path)

	return nil
}

func (b *Executor) checkOutput() error {
	fmt.Println("checking output", b.cmd.Path)
	if b.stdoutDetector != nil && (b.stdoutDetector.count != 0 || b.stderrDetector.count != 0) {
		return fmt.Errorf("exit code is 0, but race condition found")
	}
	return nil
}

func debugExecutor(b string, m map[string]string, args ...string) *Executor {
	port := "2345"
	if p := os.Getenv("GOAT_REMOTE_DEBUG_PORT"); p != "" {
		port = p
	}

	args = append([]string{
		"--listen=:" + port, "--headless=true", "--api-version=2",
		"--accept-multiclient", "exec", b, "--",
	}, args...)
	e := directExecutor("dlv", m, args...)
	e.debug = true

	return e
}

func directExecutor(binary string, envs map[string]string, args ...string) *Executor {
	// fmt.Println("create binary executor", binary, envs, args)
	fmt.Println("create binary executor", binary, args)

	cmd := exec.Command(binary, args...)
	cmd.Env = os.Environ()
	for k, v := range envs {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	b := &Executor{
		cmd: cmd,
	}

	b.stdoutDetector = &PatternDetector{
		pattern: "WARNING: DATA RACE",
	}
	b.stderrDetector = &PatternDetector{
		pattern: "WARNING: DATA RACE",
	}

	stdOutWriters := []io.Writer{b.stdoutDetector}
	stdErrWriters := []io.Writer{b.stderrDetector, os.Stderr}

	disableStdout := os.Getenv("GOAT_DISABLE_STDOUT") == TrueValue

	if !disableStdout {
		stdOutWriters = append(stdOutWriters, os.Stdout)
	}

	outputFilePath := os.Getenv("GOAT_OUTPUT_FILE")

	if outputFilePath != "" {
		if outputFile, err := os.Create(outputFilePath); err != nil {
			fmt.Printf("failed to create output file %s: %v, using stdout\n", outputFilePath, err)
		} else {
			b.outputFile = outputFile
			stdOutWriters = append(stdOutWriters, outputFile)
		}
	}

	errorsFilePath := os.Getenv("GOAT_OUTPUT_ERRORS_FILE")

	if errorsFilePath != "" {
		if errorsFile, err := os.Create(errorsFilePath); err != nil {
			fmt.Printf("failed to create errors file %s: %v, using stderr\n", errorsFilePath, err)
		} else {
			b.errorsFile = errorsFile
			stdErrWriters = append(stdErrWriters, errorsFile)
		}
	}

	if getFieldsCollectorFilePath() != "" {
		b.fieldsParser = newFieldsCollector()
		stdOutWriters = append(stdOutWriters, b.fieldsParser)
	}

	cmd.Stdout = io.MultiWriter(stdOutWriters...)
	cmd.Stderr = io.MultiWriter(stdErrWriters...)

	return b
}

func NewExecutor(binary string, envs map[string]string, args ...string) *Executor {
	if strings.ToLower(os.Getenv("GOAT_REMOTE_DEBUG")) == TrueValue {
		return debugExecutor(binary, envs, args...)
	}
	return directExecutor(binary, envs, args...)
}
