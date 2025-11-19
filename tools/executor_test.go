package tools

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDiffMaps(t *testing.T) {
	type args struct {
		map1 map[string]string
		map2 map[string]string
	}
	tests := []struct {
		wantOnlyInMap1      map[string]string
		wantOnlyInMap2      map[string]string
		wantDifferentValues map[string]string
		args                args
		name                string
	}{
		{
			name: "case 1",
			args: args{
				map1: map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key4": "value4",
				},
				map2: map[string]string{
					"key1": "value1",
					"key3": "value3",
					"key4": "value5",
				},
			},
			wantOnlyInMap1: map[string]string{
				"key2": "value2",
			},
			wantOnlyInMap2: map[string]string{
				"key3": "value3",
			},
			wantDifferentValues: map[string]string{
				"key4": "value4",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOnlyInMap1, gotOnlyInMap2, gotDifferentValues := diffMaps(tt.args.map1, tt.args.map2)
			if !reflect.DeepEqual(gotOnlyInMap1, tt.wantOnlyInMap1) {
				t.Errorf("diffMaps() gotOnlyInMap1 = %v, want %v", gotOnlyInMap1, tt.wantOnlyInMap1)
			}
			if !reflect.DeepEqual(gotOnlyInMap2, tt.wantOnlyInMap2) {
				t.Errorf("diffMaps() gotOnlyInMap2 = %v, want %v", gotOnlyInMap2, tt.wantOnlyInMap2)
			}
			if !reflect.DeepEqual(gotDifferentValues, tt.wantDifferentValues) {
				t.Errorf("diffMaps() gotDifferentValues = %v, want %v", gotDifferentValues, tt.wantDifferentValues)
			}
		})
	}
}

func TestExecutorStartStop(t *testing.T) {
	e := NewExecutor("/bin/sh", nil, "-c", `i=0; while [ $i -lt 10 ]; do echo "hello"; sleep 1; i=$((i + 1)); done`)
	err := e.Start()
	require.NoError(t, err)
	defer func() {
		_ = e.Stop()
	}()
	time.Sleep(time.Second * 2)
}

func TestExecutorRun(t *testing.T) {
	e := NewExecutor("/bin/sh", nil, "-c", `i=0; while [ $i -lt 10 ]; do echo "hello"; i=$((i + 1)); done`)
	require.NoError(t, e.Run())
}
